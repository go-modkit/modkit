package httpserver

import (
	"log/slog"
	"net/http"

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
	router.Use(modkithttp.RequestLogger(logger))
	if err := registerRoutes(modkithttp.AsRouter(router), boot.Controllers); err != nil {
		return boot, nil, err
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
