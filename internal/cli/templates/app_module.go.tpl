package app

import (
	"net/http"

	mkhttp "github.com/go-modkit/modkit/modkit/http"
	"github.com/go-modkit/modkit/modkit/module"
)

// AppModule is the root module of the application.
type AppModule struct{}

func (m *AppModule) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "app",
		Controllers: []module.ControllerDef{
			{
				Name: "HealthController",
				Build: func(r module.Resolver) (any, error) {
					return &HealthController{}, nil
				},
			},
		},
	}
}

type HealthController struct{}

func (c *HealthController) RegisterRoutes(r mkhttp.Router) {
	r.Handle(http.MethodGet, "/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
}
