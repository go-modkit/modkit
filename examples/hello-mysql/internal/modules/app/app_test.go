package app

import (
	"testing"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
)

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
