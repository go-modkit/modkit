package main

import (
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/go-modkit/modkit/examples/hello-postgres/internal/app"
	mkhttp "github.com/go-modkit/modkit/modkit/http"
	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

type HealthController struct{}

func (c *HealthController) RegisterRoutes(r mkhttp.Router) {
	r.Handle(http.MethodGet, "/health", http.HandlerFunc(c.health))
}

func (c *HealthController) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
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

func main() {
	app, err := kernel.Bootstrap(&RootModule{})
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}

	router := mkhttp.NewRouter()
	if err := mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers); err != nil {
		log.Fatalf("routes: %v", err)
	}

	log.Println("Server starting on http://localhost:8080")
	log.Println("Try: curl http://localhost:8080/health")
	if err := mkhttp.Serve(":8080", router); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
