package http

import (
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router provides a minimal method-based handler registration API.
type Router interface {
	Handle(method string, pattern string, handler http.Handler)
}

// RouteRegistrar defines a controller that can register its HTTP routes.
type RouteRegistrar interface {
	RegisterRoutes(router Router)
}

type routerAdapter struct {
	chi.Router
}

func (r routerAdapter) Handle(method string, pattern string, handler http.Handler) {
	r.Method(method, pattern, handler)
}

// AsRouter adapts a chi router to the minimal Router interface.
func AsRouter(router chi.Router) Router {
	return routerAdapter{Router: router}
}

// NewRouter creates a chi router with baseline middleware for the HTTP adapter.
func NewRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	return router
}

// RegisterRoutes invokes controller route registration functions.
func RegisterRoutes(router Router, controllers map[string]any) error {
	keys := make([]string, 0, len(controllers))
	for name := range controllers {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	registrars := make([]RouteRegistrar, 0, len(keys))
	for _, name := range keys {
		controller := controllers[name]
		registrar, ok := controller.(RouteRegistrar)
		if !ok {
			return &RouteRegistrationError{Name: name}
		}
		registrars = append(registrars, registrar)
	}

	for _, registrar := range registrars {
		registrar.RegisterRoutes(router)
	}

	return nil
}
