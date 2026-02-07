package app

import (
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/middleware"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/audit"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/database"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/users"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/logging"
	"github.com/go-modkit/modkit/modkit/module"
)

const (
	HealthControllerID       = "HealthController"
	CorsMiddlewareToken      = module.Token("app.cors_middleware")
	RateLimitMiddlewareToken = module.Token("app.rate_limit_middleware")
	TimingMiddlewareToken    = module.Token("app.timing_middleware")
)

type Module struct{}

type AppModule = Module

func NewModule() module.Module {
	return &Module{}
}

func (m *Module) Definition() module.ModuleDef {
	cfgModule := configmodule.NewModule(configmodule.Options{})
	dbModule := database.NewModule(database.Options{Config: cfgModule})
	authModule := auth.NewModule(auth.Options{Config: cfgModule})
	usersModule := users.NewModule(users.Options{Database: dbModule, Auth: authModule})
	auditModule := audit.NewModule(audit.Options{Users: usersModule})

	return module.ModuleDef{
		Name:    "app",
		Imports: []module.Module{cfgModule, dbModule, authModule, usersModule, auditModule},
		Providers: []module.ProviderDef{
			{
				Token: CorsMiddlewareToken,
				Build: func(r module.Resolver) (any, error) {
					origins, err := module.Get[[]string](r, configmodule.TokenCORSAllowedOrigins)
					if err != nil {
						return nil, err
					}
					methods, err := module.Get[[]string](r, configmodule.TokenCORSAllowedMethods)
					if err != nil {
						return nil, err
					}
					headers, err := module.Get[[]string](r, configmodule.TokenCORSAllowedHeaders)
					if err != nil {
						return nil, err
					}

					return middleware.NewCORS(middleware.CORSConfig{
						AllowedOrigins: origins,
						AllowedMethods: methods,
						AllowedHeaders: headers,
					}), nil
				},
			},
			{
				Token: RateLimitMiddlewareToken,
				Build: func(r module.Resolver) (any, error) {
					perSecond, err := module.Get[float64](r, configmodule.TokenRateLimitPerSecond)
					if err != nil {
						return nil, err
					}
					burst, err := module.Get[int](r, configmodule.TokenRateLimitBurst)
					if err != nil {
						return nil, err
					}

					return middleware.NewRateLimit(middleware.RateLimitConfig{
						RequestsPerSecond: perSecond,
						Burst:             burst,
					}), nil
				},
			},
			{
				Token: TimingMiddlewareToken,
				Build: func(_ module.Resolver) (any, error) {
					logger := logging.New().With("scope", "middleware")
					return middleware.NewTiming(logger), nil
				},
			},
		},
		Controllers: []module.ControllerDef{
			{
				Name: HealthControllerID,
				Build: func(r module.Resolver) (any, error) {
					return NewController(), nil
				},
			},
		},
	}
}
