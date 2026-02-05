package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/middleware"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/app"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/logging"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
	"github.com/go-modkit/modkit/modkit/kernel"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var registerRoutes = modkithttp.RegisterRoutes

func BuildAppHandler(opts app.Options) (*kernel.App, http.Handler, error) {
	mod := app.NewModule(opts)
	boot, err := kernel.Bootstrap(mod)
	if err != nil {
		return nil, nil, err
	}

	logger := logging.New().With(slog.String("scope", "httpserver"))
	router := modkithttp.NewRouter()
	root := modkithttp.AsRouter(router)
	root.Use(modkithttp.RequestLogger(logger))

	var registerErr error
	root.Group("/api/v1", func(r modkithttp.Router) {
		r.Use(middleware.NewCORS(middleware.CORSConfig{
			AllowedOrigins: opts.CORSAllowedOrigins,
			AllowedMethods: opts.CORSAllowedMethods,
			AllowedHeaders: nil,
		}))
		r.Use(middleware.NewRateLimit(middleware.RateLimitConfig{
			RequestsPerSecond: opts.RateLimitPerSecond,
			Burst:             opts.RateLimitBurst,
		}))
		r.Use(middleware.NewTiming(logger))
		if err := registerRoutes(r, boot.Controllers); err != nil {
			registerErr = err
		}
	})
	if registerErr != nil {
		return boot, nil, registerErr
	}
	router.Get("/swagger/*", httpSwagger.WrapHandler)
	router.Get("/docs/*", httpSwagger.WrapHandler)
	router.Get("/docs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	}))

	return boot, router, nil
}

func BuildHandler(opts app.Options) (http.Handler, error) {
	_, handler, err := BuildAppHandler(opts)
	return handler, err
}
