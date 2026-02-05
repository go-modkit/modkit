package app

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	"github.com/go-modkit/modkit/modkit/module"
)

type noopResolver struct{}

func (noopResolver) Get(module.Token) (any, error) {
	return nil, nil
}

func TestModule_DefinitionIncludesImports(t *testing.T) {
	mod := NewModule(Options{
		HTTPAddr: ":8080",
		MySQLDSN: "user:pass@tcp(localhost:3306)/app",
		Auth: auth.Config{
			Secret:   "test-secret",
			Issuer:   "test-issuer",
			TTL:      time.Minute,
			Username: "demo",
			Password: "demo",
		},
	})
	def := mod.Definition()

	if def.Name == "" {
		t.Fatalf("expected module name")
	}

	if len(def.Imports) != 4 {
		t.Fatalf("expected 4 imports, got %d", len(def.Imports))
	}

	seen := map[string]bool{}
	for _, imp := range def.Imports {
		seen[imp.Definition().Name] = true
	}

	for _, name := range []string{"database", "auth", "users", "audit"} {
		if !seen[name] {
			t.Fatalf("expected import %s", name)
		}
	}
}

func TestModule_DefinitionIncludesProviders(t *testing.T) {
	mod := NewModule(Options{
		HTTPAddr:           ":8080",
		MySQLDSN:           "user:pass@tcp(localhost:3306)/app",
		CORSAllowedOrigins: []string{"http://localhost:3000"},
		CORSAllowedMethods: []string{"GET", "POST"},
		CORSAllowedHeaders: []string{"Content-Type"},
		RateLimitPerSecond: 5.0,
		RateLimitBurst:     10,
		Auth: auth.Config{
			Secret:   "test-secret",
			Issuer:   "test-issuer",
			TTL:      time.Minute,
			Username: "demo",
			Password: "demo",
		},
	})
	def := mod.Definition()

	if len(def.Providers) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(def.Providers))
	}

	providerTokens := make(map[module.Token]bool)
	for _, provider := range def.Providers {
		providerTokens[provider.Token] = true
	}

	if !providerTokens[CorsMiddlewareToken] {
		t.Fatal("expected CorsMiddlewareToken provider")
	}
	if !providerTokens[RateLimitMiddlewareToken] {
		t.Fatal("expected RateLimitMiddlewareToken provider")
	}
	if !providerTokens[TimingMiddlewareToken] {
		t.Fatal("expected TimingMiddlewareToken provider")
	}
}

func TestModule_ProviderBuildsCORS(t *testing.T) {
	mod := NewModule(Options{
		HTTPAddr:           ":8080",
		MySQLDSN:           "user:pass@tcp(localhost:3306)/app",
		CORSAllowedOrigins: []string{"http://example.com"},
		CORSAllowedMethods: []string{"GET"},
		CORSAllowedHeaders: []string{"X-Custom"},
		Auth: auth.Config{
			Secret:   "test-secret",
			Issuer:   "test-issuer",
			TTL:      time.Minute,
			Username: "demo",
			Password: "demo",
		},
	})
	def := mod.Definition()

	var corsProvider module.ProviderDef
	for _, provider := range def.Providers {
		if provider.Token == CorsMiddlewareToken {
			corsProvider = provider
			break
		}
	}

	if corsProvider.Token == "" {
		t.Fatal("expected CORS provider")
	}

	middleware, err := corsProvider.Build(noopResolver{})
	if err != nil {
		t.Fatalf("build CORS middleware: %v", err)
	}

	mwFunc, ok := middleware.(func(http.Handler) http.Handler)
	if !ok {
		t.Fatalf("expected func(http.Handler) http.Handler, got %T", middleware)
	}
	if mwFunc == nil {
		t.Fatal("expected non-nil middleware function")
	}
}

func TestModule_ProviderBuildsRateLimit(t *testing.T) {
	mod := NewModule(Options{
		HTTPAddr:           ":8080",
		MySQLDSN:           "user:pass@tcp(localhost:3306)/app",
		RateLimitPerSecond: 10.5,
		RateLimitBurst:     20,
		Auth: auth.Config{
			Secret:   "test-secret",
			Issuer:   "test-issuer",
			TTL:      time.Minute,
			Username: "demo",
			Password: "demo",
		},
	})
	def := mod.Definition()

	var rateLimitProvider module.ProviderDef
	for _, provider := range def.Providers {
		if provider.Token == RateLimitMiddlewareToken {
			rateLimitProvider = provider
			break
		}
	}

	if rateLimitProvider.Token == "" {
		t.Fatal("expected RateLimit provider")
	}

	middleware, err := rateLimitProvider.Build(noopResolver{})
	if err != nil {
		t.Fatalf("build RateLimit middleware: %v", err)
	}

	mwFunc, ok := middleware.(func(http.Handler) http.Handler)
	if !ok {
		t.Fatalf("expected func(http.Handler) http.Handler, got %T", middleware)
	}
	if mwFunc == nil {
		t.Fatal("expected non-nil middleware function")
	}
}

func TestModule_ProviderBuildsTiming(t *testing.T) {
	mod := NewModule(Options{
		HTTPAddr: ":8080",
		MySQLDSN: "user:pass@tcp(localhost:3306)/app",
		Auth: auth.Config{
			Secret:   "test-secret",
			Issuer:   "test-issuer",
			TTL:      time.Minute,
			Username: "demo",
			Password: "demo",
		},
	})
	def := mod.Definition()

	var timingProvider module.ProviderDef
	for _, provider := range def.Providers {
		if provider.Token == TimingMiddlewareToken {
			timingProvider = provider
			break
		}
	}

	if timingProvider.Token == "" {
		t.Fatal("expected Timing provider")
	}

	middleware, err := timingProvider.Build(noopResolver{})
	if err != nil {
		t.Fatalf("build Timing middleware: %v", err)
	}

	mwFunc, ok := middleware.(func(http.Handler) http.Handler)
	if !ok {
		t.Fatalf("expected func(http.Handler) http.Handler, got %T", middleware)
	}
	if mwFunc == nil {
		t.Fatal("expected non-nil middleware function")
	}
}

func TestModule_DefinitionIncludesController(t *testing.T) {
	mod := NewModule(Options{
		HTTPAddr: ":8080",
		MySQLDSN: "user:pass@tcp(localhost:3306)/app",
		Auth: auth.Config{
			Secret:   "test-secret",
			Issuer:   "test-issuer",
			TTL:      time.Minute,
			Username: "demo",
			Password: "demo",
		},
	})
	def := mod.Definition()

	if len(def.Controllers) != 1 {
		t.Fatalf("expected 1 controller, got %d", len(def.Controllers))
	}

	if def.Controllers[0].Name != HealthControllerID {
		t.Fatalf("expected controller name %q, got %q", HealthControllerID, def.Controllers[0].Name)
	}
}

func TestModule_ControllerBuilds(t *testing.T) {
	mod := NewModule(Options{
		HTTPAddr: ":8080",
		MySQLDSN: "user:pass@tcp(localhost:3306)/app",
		Auth: auth.Config{
			Secret:   "test-secret",
			Issuer:   "test-issuer",
			TTL:      time.Minute,
			Username: "demo",
			Password: "demo",
		},
	})
	def := mod.Definition()

	controller, err := def.Controllers[0].Build(noopResolver{})
	if err != nil {
		t.Fatalf("build controller: %v", err)
	}

	ctrl, ok := controller.(*Controller)
	if !ok {
		t.Fatalf("expected *Controller, got %T", controller)
	}
	if ctrl == nil {
		t.Fatal("expected non-nil controller")
	}
}
