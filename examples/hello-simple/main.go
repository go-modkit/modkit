// Package main demonstrates a minimal modkit application.
//
// This example shows the core concepts without any external dependencies:
// - Module definition with providers and controllers
// - Dependency injection via token resolution
// - HTTP route registration
//
// Run with: go run main.go
// Test with: curl http://localhost:8080/greet
package main

import (
	"encoding/json"
	"log"
	"net/http"

	mkhttp "github.com/aryeko/modkit/modkit/http"
	"github.com/aryeko/modkit/modkit/kernel"
	"github.com/aryeko/modkit/modkit/module"
)

// Tokens identify providers
const (
	TokenGreeting module.Token = "greeting.message"
	TokenCounter  module.Token = "greeting.counter"
)

// Counter tracks greeting count (demonstrates stateful providers).
// Note: not concurrency-safe; real applications should use sync.Mutex or atomic.
type Counter struct {
	count int
}

func (c *Counter) Increment() int {
	c.count++
	return c.count
}

// GreetingController handles HTTP requests
type GreetingController struct {
	message string
	counter *Counter
}

func (c *GreetingController) RegisterRoutes(r mkhttp.Router) {
	r.Handle(http.MethodGet, "/greet", http.HandlerFunc(c.Greet))
	r.Handle(http.MethodGet, "/health", http.HandlerFunc(c.Health))
}

func (c *GreetingController) Greet(w http.ResponseWriter, r *http.Request) {
	count := c.counter.Increment()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": c.message,
		"count":   count,
	})
}

func (c *GreetingController) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// AppModule is the root module
type AppModule struct {
	message string
}

func NewAppModule(message string) *AppModule {
	return &AppModule{message: message}
}

func (m *AppModule) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "app",
		Providers: []module.ProviderDef{
			{
				Token: TokenGreeting,
				Build: func(r module.Resolver) (any, error) {
					return m.message, nil
				},
			},
			{
				Token: TokenCounter,
				Build: func(r module.Resolver) (any, error) {
					return &Counter{}, nil
				},
			},
		},
		Controllers: []module.ControllerDef{
			{
				Name: "GreetingController",
				Build: func(r module.Resolver) (any, error) {
					msg, err := r.Get(TokenGreeting)
					if err != nil {
						return nil, err
					}
					counter, err := r.Get(TokenCounter)
					if err != nil {
						return nil, err
					}
					return &GreetingController{
						message: msg.(string),
						counter: counter.(*Counter),
					}, nil
				},
			},
		},
	}
}

func main() {
	// Create and bootstrap the app module
	appModule := NewAppModule("Hello from modkit!")

	app, err := kernel.Bootstrap(appModule)
	if err != nil {
		log.Fatalf("Failed to bootstrap: %v", err)
	}

	// Create router and register controllers
	router := mkhttp.NewRouter()
	if err := mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers); err != nil {
		log.Fatalf("Failed to register routes: %v", err)
	}

	// Start server
	log.Println("Server starting on http://localhost:8080")
	log.Println("Try: curl http://localhost:8080/greet")
	log.Println("Try: curl http://localhost:8080/health")

	if err := mkhttp.Serve(":8080", router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
