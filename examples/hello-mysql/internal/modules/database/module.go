package database

import (
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
	return module.ModuleDef{
		Name: "database",
		Providers: []module.ProviderDef{
			{
				Token: TokenDB,
				Build: func(r module.Resolver) (any, error) {
					return mysql.Open(m.opts.DSN)
				},
			},
		},
		Exports: []module.Token{TokenDB},
	}
}
