package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/app"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/logging"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
	"github.com/go-modkit/modkit/modkit/kernel"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var registerRoutes = modkithttp.RegisterRoutes

func BuildAppHandler() (*kernel.App, http.Handler, error) {
	mod := app.NewModule()
	boot, err := kernel.Bootstrap(mod)
	if err != nil {
		return nil, nil, err
	}

	logger := logging.New().With(slog.String("scope", "httpserver"))
	router := modkithttp.NewRouter()
	root := modkithttp.AsRouter(router)
	root.Use(modkithttp.RequestLogger(logger))

	corsAny, err := boot.Get(app.CorsMiddlewareToken)
	if err != nil {
		return boot, nil, err
	}
	cors, ok := corsAny.(func(http.Handler) http.Handler)
	if !ok {
		return boot, nil, fmt.Errorf("cors middleware: expected func(http.Handler) http.Handler, got %T", corsAny)
	}

	rateLimitAny, err := boot.Get(app.RateLimitMiddlewareToken)
	if err != nil {
		return boot, nil, err
	}
	rateLimit, ok := rateLimitAny.(func(http.Handler) http.Handler)
	if !ok {
		return boot, nil, fmt.Errorf("rate limit middleware: expected func(http.Handler) http.Handler, got %T", rateLimitAny)
	}

	timingAny, err := boot.Get(app.TimingMiddlewareToken)
	if err != nil {
		return boot, nil, err
	}
	timing, ok := timingAny.(func(http.Handler) http.Handler)
	if !ok {
		return boot, nil, fmt.Errorf("timing middleware: expected func(http.Handler) http.Handler, got %T", timingAny)
	}

	var registerErr error
	root.Group("/api/v1", func(r modkithttp.Router) {
		r.Use(cors)
		r.Use(rateLimit)
		r.Use(timing)
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

func BuildHandler() (http.Handler, error) {
	_, handler, err := BuildAppHandler()
	return handler, err
}
