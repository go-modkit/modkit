package database

import (
	"context"
	"errors"
	"testing"
)

func TestModuleDefinition_ProviderCleanupHook_CanceledContext(t *testing.T) {
	def := Module{}.Definition()
	if len(def.Providers) == 0 {
		t.Fatal("expected at least one provider")
	}
	var cleanup func(ctx context.Context) error
	for _, provider := range def.Providers {
		if provider.Token == TokenDB {
			cleanup = provider.Cleanup
			break
		}
	}
	if cleanup == nil {
		t.Fatal("expected provider cleanup hook")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := cleanup(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDatabaseModule_Definition_ProvidesDB(t *testing.T) {
	mod := NewModule(Options{DSN: "dsn"})
	def := mod.(*Module).Definition()
	if def.Name != "database" {
		t.Fatalf("expected name database, got %q", def.Name)
	}
	if len(def.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(def.Providers))
	}
	if def.Providers[0].Token != TokenDB {
		t.Fatalf("expected TokenDB, got %q", def.Providers[0].Token)
	}
	if def.Providers[0].Cleanup == nil {
		t.Fatal("expected cleanup hook")
	}
}

func TestDatabaseModule_ProviderBuildError(t *testing.T) {
	mod := NewModule(Options{DSN: ""})
	def := mod.(*Module).Definition()
	provider := def.Providers[0]

	// Use a stub resolver - the error will come from mysql.Open with empty DSN
	_, err := provider.Build(nil)
	if err == nil {
		t.Fatal("expected error for empty DSN")
	}
	if err.Error() != "mysql dsn is required" {
		t.Fatalf("expected 'mysql dsn is required' error, got %q", err.Error())
	}
}
