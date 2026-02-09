package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-modkit/modkit/examples/hello-postgres/internal/app"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

type HealthController struct{}

func (c *HealthController) RegisterRoutes(r modkithttp.Router) {
	r.Handle(http.MethodGet, "/health", http.HandlerFunc(c.health))
}

func (c *HealthController) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type RootModule struct{}

func (m *RootModule) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "api",
		Imports: []module.Module{
			app.NewModule(),
		},
		Controllers: []module.ControllerDef{{
			Name: "HealthController",
			Build: func(_ module.Resolver) (any, error) {
				return &HealthController{}, nil
			},
		}},
	}
}

var registerRoutes = modkithttp.RegisterRoutes

func BuildAppHandler() (*kernel.App, http.Handler, error) {
	boot, err := kernel.Bootstrap(&RootModule{})
	if err != nil {
		return nil, nil, err
	}

	router := modkithttp.NewRouter()
	root := modkithttp.AsRouter(router)

	var registerErr error
	root.Group("/api/v1", func(r modkithttp.Router) {
		if err := registerRoutes(r, boot.Controllers); err != nil {
			registerErr = err
		}
	})
	if registerErr != nil {
		return boot, nil, registerErr
	}

	return boot, router, nil
}

func BuildHandler() (http.Handler, error) {
	_, handler, err := BuildAppHandler()
	return handler, err
}
