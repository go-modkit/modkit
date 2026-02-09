package database

import (
	"context"
	"database/sql"

	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/mysql"
	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
	"github.com/go-modkit/modkit/modkit/module"
)

const (
	// TokenDB is kept for backwards compatibility with the existing hello-mysql
	// token path.
	TokenDB module.Token = sqlmodule.TokenDB

	TokenDialect module.Token = sqlmodule.TokenDialect
)

type Options struct {
	Config module.Module
}

type Module struct {
	opts Options
}

type DatabaseModule = Module

func NewModule(opts Options) module.Module {
	if opts.Config == nil {
		opts.Config = configmodule.DefaultModule()
	}
	return &Module{opts: opts}
}

func (m Module) Definition() module.ModuleDef {
	configMod := m.opts.Config
	if configMod == nil {
		configMod = configmodule.DefaultModule()
	}

	var db *sql.DB
	return module.ModuleDef{
		Name:    "database",
		Imports: []module.Module{configMod},
		Providers: []module.ProviderDef{
			{
				Token: TokenDB,
				Build: func(r module.Resolver) (any, error) {
					dsn, err := module.Get[string](r, configmodule.TokenMySQLDSN)
					if err != nil {
						return nil, err
					}

					db, err = mysql.Open(dsn)
					if err != nil {
						return nil, err
					}
					return db, nil
				},
				Cleanup: func(ctx context.Context) error {
					return CleanupDB(ctx, db)
				},
			},
			{
				Token: TokenDialect,
				Build: func(_ module.Resolver) (any, error) {
					return sqlmodule.DialectMySQL, nil
				},
			},
		},
		Exports: []module.Token{TokenDB, TokenDialect},
	}
}
