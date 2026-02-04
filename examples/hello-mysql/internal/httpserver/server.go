package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/aryeko/modkit/examples/hello-mysql/internal/modules/app"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/logging"
	modkithttp "github.com/aryeko/modkit/modkit/http"
	"github.com/aryeko/modkit/modkit/kernel"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func BuildHandler(opts app.Options) (http.Handler, error) {
	mod := app.NewModule(opts)
	boot, err := kernel.Bootstrap(mod)
	if err != nil {
		return nil, err
	}

	logger := logging.New().With(slog.String("scope", "httpserver"))
	router := modkithttp.NewRouter()
	router.Use(modkithttp.RequestLogger(logger))
	if err := modkithttp.RegisterRoutes(modkithttp.AsRouter(router), boot.Controllers); err != nil {
		return nil, err
	}
	router.Get("/swagger/*", httpSwagger.WrapHandler)
	router.Get("/docs/*", httpSwagger.WrapHandler)
	router.Get("/docs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	}))

	return router, nil
}
