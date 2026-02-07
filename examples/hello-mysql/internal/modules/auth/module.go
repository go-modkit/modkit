package auth

import (
	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/modkit/module"
)

const (
	TokenMiddleware module.Token = "auth.middleware"
	TokenHandler    module.Token = "auth.handler"
)

type Options struct {
	Config module.Module
}

type Module struct {
	opts Options
}

type AuthModule = Module

func NewModule(opts Options) module.Module {
	if opts.Config == nil {
		opts.Config = configmodule.NewModule(configmodule.Options{})
	}
	return &Module{opts: opts}
}

func (m Module) Definition() module.ModuleDef {
	configMod := m.opts.Config
	if configMod == nil {
		configMod = configmodule.NewModule(configmodule.Options{})
	}

	return module.ModuleDef{
		Name:      "auth",
		Imports:   []module.Module{configMod},
		Providers: Providers(),
		Controllers: []module.ControllerDef{
			{
				Name: "AuthController",
				Build: func(r module.Resolver) (any, error) {
					return module.Get[*Handler](r, TokenHandler)
				},
			},
		},
		Exports: []module.Token{TokenMiddleware},
	}
}
