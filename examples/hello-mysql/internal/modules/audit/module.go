package audit

import (
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/users"
	"github.com/go-modkit/modkit/modkit/module"
)

const TokenService module.Token = "audit.service"

type Options struct {
	Users module.Module
}

type Module struct {
	opts Options
}

func NewModule(opts Options) module.Module {
	return &Module{opts: opts}
}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name:    "audit",
		Imports: []module.Module{m.opts.Users},
		Providers: []module.ProviderDef{
			{
				Token: TokenService,
				Build: func(r module.Resolver) (any, error) {
					usersSvc, err := module.Get[users.Service](r, users.TokenService)
					if err != nil {
						return nil, err
					}
					return NewService(usersSvc), nil
				},
			},
		},
	}
}
