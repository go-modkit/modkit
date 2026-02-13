package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
	"github.com/go-modkit/modkit/modkit/module"
)

const (
	driverName     = "postgres"
	moduleNameBase = "data.postgres"
)

// Options configures a Postgres provider module.
type Options struct {
	// Config provides Postgres configuration tokens (DSN, pool settings, ping timeout).
	Config module.Module
	// Name namespaces exported SQL contract tokens via sqlmodule.NamedTokens.
	Name string
}

// Module provides a Postgres-backed *sql.DB and dialect token.
type Module struct {
	opts Options
}

// NewModule constructs a Postgres provider module.
func NewModule(opts Options) module.Module {
	if opts.Config == nil {
		opts.Config = configModule(opts.Name)
	}
	return &Module{opts: opts}
}

// Definition returns the module definition for graph construction.
func (m *Module) Definition() module.ModuleDef {
	configMod := m.opts.Config
	if configMod == nil {
		configMod = configModule(m.opts.Name)
	}

	toks, err := sqlmodule.NamedTokens(m.opts.Name)
	if err != nil {
		return invalidModuleDef(err)
	}

	var db *sql.DB
	return module.ModuleDef{
		Name:    moduleName(m.opts.Name),
		Imports: []module.Module{configMod},
		Providers: []module.ProviderDef{
			{
				Token: toks.DB,
				Build: func(r module.Resolver) (any, error) {
					built, buildErr := buildDB(r, toks.DB)
					if buildErr != nil {
						return nil, buildErr
					}
					db = built
					return db, nil
				},
				Cleanup: func(ctx context.Context) error {
					return CleanupDB(ctx, db)
				},
			},
			{
				Token: toks.Dialect,
				Build: func(_ module.Resolver) (any, error) {
					return sqlmodule.DialectPostgres, nil
				},
			},
		},
		Exports: []module.Token{toks.DB, toks.Dialect},
	}
}

func moduleName(name string) string {
	if name == "" {
		return moduleNameBase
	}
	return moduleNameBase + "." + name
}

func invalidModuleDef(err error) module.ModuleDef {
	return module.ModuleDef{
		Name: moduleNameBase + ".invalid",
		Controllers: []module.ControllerDef{{
			Name: "InvalidPostgresModule",
			Build: func(_ module.Resolver) (any, error) {
				return nil, err
			},
		}},
	}
}

func buildDB(r module.Resolver, dbToken module.Token) (*sql.DB, error) {
	dsn, err := module.Get[string](r, TokenDSN)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("dsn: %w", err)}
	}
	maxOpen, err := module.Get[int](r, TokenMaxOpenConns)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("max_open_conns: %w", err)}
	}
	maxIdle, err := module.Get[int](r, TokenMaxIdleConns)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("max_idle_conns: %w", err)}
	}
	maxIdleSet, err := module.Get[bool](r, tokenMaxIdleConnsSet)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("max_idle_conns_set: %w", err)}
	}
	maxLifetime, err := module.Get[time.Duration](r, TokenConnMaxLifetime)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("conn_max_lifetime: %w", err)}
	}
	connectTimeout, err := module.Get[time.Duration](r, TokenConnectTimeout)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("connect_timeout: %w", err)}
	}

	if maxOpen < 0 {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageInvalidConfig, Err: fmt.Errorf("max_open_conns must be >= 0")}
	}
	if maxIdle < 0 {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageInvalidConfig, Err: fmt.Errorf("max_idle_conns must be >= 0")}
	}
	if maxLifetime < 0 {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageInvalidConfig, Err: fmt.Errorf("conn_max_lifetime must be >= 0")}
	}
	if connectTimeout < 0 {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageInvalidConfig, Err: fmt.Errorf("connect_timeout must be >= 0")}
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageOpen, Err: err}
	}

	if maxOpen > 0 {
		db.SetMaxOpenConns(maxOpen)
	}
	if maxIdleSet {
		db.SetMaxIdleConns(maxIdle)
	}
	if maxLifetime > 0 {
		db.SetConnMaxLifetime(maxLifetime)
	}

	if connectTimeout == 0 {
		return db, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StagePing, Err: err}
	}

	return db, nil
}
