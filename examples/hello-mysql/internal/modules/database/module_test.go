package database

import (
	"context"
	"testing"
)

func TestModuleDefinition_ProviderCleanupHook(t *testing.T) {
	def := Module{}.Definition()
	if len(def.Providers) == 0 {
		t.Fatal("expected at least one provider")
	}
	cleanup := def.Providers[0].Cleanup
	if cleanup == nil {
		t.Fatal("expected provider cleanup hook")
	}
	if err := cleanup(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
