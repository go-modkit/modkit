package app

import "testing"

func TestModuleDefinition(t *testing.T) {
	def := NewModule().Definition()

	if def.Name != "app" {
		t.Fatalf("expected name=app, got %q", def.Name)
	}
	if len(def.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(def.Imports))
	}
}
