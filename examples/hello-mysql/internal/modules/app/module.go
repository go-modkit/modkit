package app

import (
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/middleware"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/audit"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
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

type Options struct {
	HTTPAddr           string
	MySQLDSN           string
	Auth               auth.Config
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	RateLimitPerSecond float64
	RateLimitBurst     int
}

type Module struct {
	opts Options
}

type AppModule = Module

func NewModule(opts Options) module.Module {
	return &Module{opts: opts}
}

func (m *Module) Definition() module.ModuleDef {
	dbModule := database.NewModule(database.Options{DSN: m.opts.MySQLDSN})
	authModule := auth.NewModule(auth.Options{Config: m.opts.Auth})
	usersModule := users.NewModule(users.Options{Database: dbModule, Auth: authModule})
	auditModule := audit.NewModule(audit.Options{Users: usersModule})

	return module.ModuleDef{
		Name:    "app",
		Imports: []module.Module{dbModule, authModule, usersModule, auditModule},
		Providers: []module.ProviderDef{
			{
				Token: CorsMiddlewareToken,
				Build: func(_ module.Resolver) (any, error) {
					return middleware.NewCORS(middleware.CORSConfig{
						AllowedOrigins: m.opts.CORSAllowedOrigins,
						AllowedMethods: m.opts.CORSAllowedMethods,
						AllowedHeaders: nil,
					}), nil
				},
			},
			{
				Token: RateLimitMiddlewareToken,
				Build: func(_ module.Resolver) (any, error) {
					return middleware.NewRateLimit(middleware.RateLimitConfig{
						RequestsPerSecond: m.opts.RateLimitPerSecond,
						Burst:             m.opts.RateLimitBurst,
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
