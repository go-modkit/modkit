package auth

import (
	"time"

	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/modkit/module"
)

func Providers() []module.ProviderDef {
	return []module.ProviderDef{
		{
			Token: TokenMiddleware,
			Build: func(r module.Resolver) (any, error) {
				cfg, err := loadConfig(r)
				if err != nil {
					return nil, err
				}
				return NewJWTMiddleware(cfg), nil
			},
		},
		{
			Token: TokenHandler,
			Build: func(r module.Resolver) (any, error) {
				cfg, err := loadConfig(r)
				if err != nil {
					return nil, err
				}
				return NewHandler(cfg), nil
			},
		},
	}
}

func loadConfig(r module.Resolver) (Config, error) {
	secret, err := module.Get[string](r, configmodule.TokenJWTSecret)
	if err != nil {
		return Config{}, err
	}
	issuer, err := module.Get[string](r, configmodule.TokenJWTIssuer)
	if err != nil {
		return Config{}, err
	}
	ttl, err := module.Get[time.Duration](r, configmodule.TokenJWTTTL)
	if err != nil {
		return Config{}, err
	}
	username, err := module.Get[string](r, configmodule.TokenAuthUsername)
	if err != nil {
		return Config{}, err
	}
	password, err := module.Get[string](r, configmodule.TokenAuthPassword)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Secret:   secret,
		Issuer:   issuer,
		TTL:      ttl,
		Username: username,
		Password: password,
	}, nil
}
