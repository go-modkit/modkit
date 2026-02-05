package database

import (
	"context"
	"database/sql"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/mysql"
	"github.com/go-modkit/modkit/modkit/module"
)

const TokenDB module.Token = "database.db"

type Options struct {
	DSN string
}

type Module struct {
	opts Options
}

type DatabaseModule = Module

func NewModule(opts Options) module.Module {
	return &Module{opts: opts}
}

func (m Module) Definition() module.ModuleDef {
	var db *sql.DB
	return module.ModuleDef{
		Name: "database",
		Providers: []module.ProviderDef{
			{
				Token: TokenDB,
				Build: func(r module.Resolver) (any, error) {
					var err error
					db, err = mysql.Open(m.opts.DSN)
					if err != nil {
						return nil, err
					}
					return db, nil
				},
				Cleanup: func(ctx context.Context) error {
					return CleanupDB(ctx, db)
				},
			},
		},
		Exports: []module.Token{TokenDB},
	}
}
