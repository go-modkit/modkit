package auth

import "github.com/go-modkit/modkit/modkit/module"

const (
	TokenMiddleware module.Token = "auth.middleware"
	TokenHandler    module.Token = "auth.handler"
)

type Options struct {
	Config Config
}

type Module struct {
	opts Options
}

type AuthModule = Module

func NewModule(opts Options) module.Module {
	return &Module{opts: opts}
}

func (m Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name:      "auth",
		Providers: Providers(m.opts.Config),
		Controllers: []module.ControllerDef{
			{
				Name: "AuthController",
				Build: func(r module.Resolver) (any, error) {
					handlerAny, err := r.Get(TokenHandler)
					if err != nil {
						return nil, err
					}
					return handlerAny, nil
				},
			},
		},
	}
}
