package config

import (
	"errors"
	"reflect"
	"testing"
	"time"

	mkconfig "github.com/go-modkit/modkit/modkit/config"
	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

type mapSource map[string]string

func (m mapSource) Lookup(key string) (string, bool) {
	v, ok := m[key]
	return v, ok
}

type rootModule struct {
	imports []module.Module
}

func (m *rootModule) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name:    "root",
		Imports: m.imports,
	}
}

func TestDefinitionExportsAllTokens(t *testing.T) {
	mod := NewModule(Options{})
	def := mod.Definition()

	if def.Name != "config" {
		t.Fatalf("unexpected module name: %q", def.Name)
	}
	if len(def.Imports) != 1 {
		t.Fatalf("expected single import, got %d", len(def.Imports))
	}
	if len(def.Exports) != len(exportedTokens) {
		t.Fatalf("unexpected exports count: got %d want %d", len(def.Exports), len(exportedTokens))
	}
}

func TestResolvesDefaults(t *testing.T) {
	app, err := kernel.Bootstrap(&rootModule{imports: []module.Module{NewModule(Options{Source: mapSource{}})}})
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	httpAddr, err := module.Get[string](app, TokenHTTPAddr)
	if err != nil {
		t.Fatalf("get http addr: %v", err)
	}
	if httpAddr != ":8080" {
		t.Fatalf("unexpected HTTP_ADDR default: %q", httpAddr)
	}

	jwtTTL, err := module.Get[time.Duration](app, TokenJWTTTL)
	if err != nil {
		t.Fatalf("get jwt ttl: %v", err)
	}
	if jwtTTL != time.Hour {
		t.Fatalf("unexpected JWT_TTL default: %v", jwtTTL)
	}

	origins, err := module.Get[[]string](app, TokenCORSAllowedOrigins)
	if err != nil {
		t.Fatalf("get cors origins: %v", err)
	}
	if !reflect.DeepEqual(origins, []string{"http://localhost:3000"}) {
		t.Fatalf("unexpected CORS origins default: %v", origins)
	}
}

func TestResolvesSourceOverrides(t *testing.T) {
	src := mapSource{
		"HTTP_ADDR":             ":9090",
		"MYSQL_DSN":             "user:pass@tcp(host:3307)/db",
		"JWT_SECRET":            "custom-secret",
		"JWT_ISSUER":            "custom-issuer",
		"JWT_TTL":               "2h",
		"AUTH_USERNAME":         "alice",
		"AUTH_PASSWORD":         "pw",
		"CORS_ALLOWED_ORIGINS":  "https://a.example,https://b.example",
		"CORS_ALLOWED_METHODS":  "GET,POST",
		"CORS_ALLOWED_HEADERS":  "Content-Type,Authorization",
		"RATE_LIMIT_PER_SECOND": "7.5",
		"RATE_LIMIT_BURST":      "25",
	}

	app, err := kernel.Bootstrap(&rootModule{imports: []module.Module{NewModule(Options{Source: src})}})
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	httpAddr, err := module.Get[string](app, TokenHTTPAddr)
	if err != nil || httpAddr != ":9090" {
		t.Fatalf("unexpected HTTP_ADDR: %q err=%v", httpAddr, err)
	}

	jwtTTL, err := module.Get[time.Duration](app, TokenJWTTTL)
	if err != nil || jwtTTL != 2*time.Hour {
		t.Fatalf("unexpected JWT_TTL: %v err=%v", jwtTTL, err)
	}

	rateLimit, err := module.Get[float64](app, TokenRateLimitPerSecond)
	if err != nil || rateLimit != 7.5 {
		t.Fatalf("unexpected RATE_LIMIT_PER_SECOND: %v err=%v", rateLimit, err)
	}

	burst, err := module.Get[int](app, TokenRateLimitBurst)
	if err != nil || burst != 25 {
		t.Fatalf("unexpected RATE_LIMIT_BURST: %d err=%v", burst, err)
	}
}

func TestRejectsNonPositiveJWTTTL(t *testing.T) {
	src := mapSource{"JWT_TTL": "0s"}

	app, err := kernel.Bootstrap(&rootModule{imports: []module.Module{NewModule(Options{Source: src})}})
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	_, err = module.Get[time.Duration](app, TokenJWTTTL)
	if err == nil {
		t.Fatalf("expected JWT_TTL parse error")
	}

	var parseErr *mkconfig.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("expected ParseError, got %T", err)
	}
}
