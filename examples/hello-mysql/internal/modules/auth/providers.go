package auth

import "github.com/go-modkit/modkit/modkit/module"

func Providers(cfg Config) []module.ProviderDef {
	return []module.ProviderDef{
		{
			Token: TokenMiddleware,
			Build: func(r module.Resolver) (any, error) {
				return NewJWTMiddleware(cfg), nil
			},
		},
		{
			Token: TokenHandler,
			Build: func(r module.Resolver) (any, error) {
				return NewHandler(cfg), nil
			},
		},
	}
}
