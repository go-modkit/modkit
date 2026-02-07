package app

import (
	"errors"
	"net/http"
	"testing"

	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/modkit/module"
)

type mapResolver map[module.Token]any

func (r mapResolver) Get(token module.Token) (any, error) {
	v, ok := r[token]
	if !ok {
		return nil, errors.New("missing token")
	}
	return v, nil
}

func TestModule_DefinitionIncludesImports(t *testing.T) {
	mod := NewModule()
	def := mod.Definition()

	if def.Name == "" {
		t.Fatalf("expected module name")
	}

	if len(def.Imports) != 5 {
		t.Fatalf("expected 5 imports, got %d", len(def.Imports))
	}

	seen := map[string]bool{}
	for _, imp := range def.Imports {
		seen[imp.Definition().Name] = true
	}

	for _, name := range []string{"config", "database", "auth", "users", "audit"} {
		if !seen[name] {
			t.Fatalf("expected import %s", name)
		}
	}
}

func TestModule_DefinitionIncludesProviders(t *testing.T) {
	mod := NewModule()
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
	mod := NewModule()
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

	middleware, err := corsProvider.Build(mapResolver{
		configmodule.TokenCORSAllowedOrigins: []string{"http://example.com"},
		configmodule.TokenCORSAllowedMethods: []string{"GET"},
		configmodule.TokenCORSAllowedHeaders: []string{"X-Custom"},
	})
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
	mod := NewModule()
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

	middleware, err := rateLimitProvider.Build(mapResolver{
		configmodule.TokenRateLimitPerSecond: 10.5,
		configmodule.TokenRateLimitBurst:     20,
	})
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
	mod := NewModule()
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

	middleware, err := timingProvider.Build(mapResolver{})
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
	mod := NewModule()
	def := mod.Definition()

	if len(def.Controllers) != 1 {
		t.Fatalf("expected 1 controller, got %d", len(def.Controllers))
	}

	if def.Controllers[0].Name != HealthControllerID {
		t.Fatalf("expected controller name %q, got %q", HealthControllerID, def.Controllers[0].Name)
	}
}

func TestModule_ControllerBuilds(t *testing.T) {
	mod := NewModule()
	def := mod.Definition()

	controller, err := def.Controllers[0].Build(mapResolver{})
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
