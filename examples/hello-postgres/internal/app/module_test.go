package app

import (
	"testing"

	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
)

func TestModuleDefinition(t *testing.T) {
	def := NewModule().Definition()

	if def.Name != "app" {
		t.Fatalf("expected name=app, got %q", def.Name)
	}
	if len(def.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(def.Imports))
	}
	if len(def.Exports) != 2 {
		t.Fatalf("expected 2 exports, got %d", len(def.Exports))
	}

	foundDB := false
	foundDialect := false
	for _, token := range def.Exports {
		switch token {
		case sqlmodule.TokenDB:
			foundDB = true
		case sqlmodule.TokenDialect:
			foundDialect = true
		}
	}
	if !foundDB || !foundDialect {
		t.Fatalf("expected db and dialect exports")
	}
}
