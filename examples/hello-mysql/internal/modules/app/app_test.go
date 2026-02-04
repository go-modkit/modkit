package app

import "testing"

func TestModule_DefinitionIncludesImports(t *testing.T) {
	mod := NewModule(Options{HTTPAddr: ":8080", MySQLDSN: "user:pass@tcp(localhost:3306)/app"})
	def := mod.Definition()

	if def.Name == "" {
		t.Fatalf("expected module name")
	}

	if len(def.Imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(def.Imports))
	}

	seen := map[string]bool{}
	for _, imp := range def.Imports {
		seen[imp.Definition().Name] = true
	}

	for _, name := range []string{"database", "users", "audit"} {
		if !seen[name] {
			t.Fatalf("expected import %s", name)
		}
	}
}
