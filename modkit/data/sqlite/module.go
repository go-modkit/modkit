package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
	"github.com/go-modkit/modkit/modkit/module"
)

const (
	driverName     = "sqlite3"
	moduleNameBase = "data.sqlite"
)

// Options configures a SQLite provider module.
type Options struct {
	// Config provides SQLite configuration tokens (path/DSN, DSN options, ping timeout).
	Config module.Module
	// Name namespaces exported SQL contract tokens via sqlmodule.NamedTokens.
	Name string
}

// Module provides a SQLite-backed *sql.DB and dialect token.
type Module struct {
	opts Options
}

// NewModule constructs a SQLite provider module.
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
					return sqlmodule.DialectSQLite, nil
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
			Name: "InvalidSQLiteModule",
			Build: func(_ module.Resolver) (any, error) {
				return nil, err
			},
		}},
	}
}

func buildDB(r module.Resolver, dbToken module.Token) (*sql.DB, error) {
	path, err := module.Get[string](r, TokenPath)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("path: %w", err)}
	}
	busyTimeout, err := module.Get[time.Duration](r, TokenBusyTimeout)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("busy_timeout: %w", err)}
	}
	journalMode, err := module.Get[string](r, TokenJournalMode)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("journal_mode: %w", err)}
	}
	connectTimeout, err := module.Get[time.Duration](r, TokenConnectTimeout)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageResolveConfig, Err: fmt.Errorf("connect_timeout: %w", err)}
	}

	if busyTimeout < 0 {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageInvalidConfig, Err: fmt.Errorf("busy_timeout must be >= 0")}
	}
	if connectTimeout < 0 {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageInvalidConfig, Err: fmt.Errorf("connect_timeout must be >= 0")}
	}

	dsn := buildDSN(path, busyTimeout, journalMode)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, &BuildError{Provider: driverName, Token: dbToken, Stage: StageOpen, Err: err}
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

func buildDSN(base string, busyTimeout time.Duration, journalMode string) string {
	journalMode = strings.TrimSpace(journalMode)

	params := url.Values{}
	if busyTimeout > 0 {
		params.Set("_busy_timeout", strconv.FormatInt(int64(busyTimeout/time.Millisecond), 10))
	}
	if journalMode != "" {
		params.Set("_journal_mode", journalMode)
	}
	if len(params) == 0 {
		return base
	}

	enc := params.Encode()
	if strings.Contains(base, "?") {
		if strings.HasSuffix(base, "?") || strings.HasSuffix(base, "&") {
			return base + enc
		}
		return base + "&" + enc
	}
	return base + "?" + enc
}
