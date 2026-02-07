package database

import (
	"context"
	"errors"
	"testing"

	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/modkit/module"
)

type resolverMap map[module.Token]any

func (m resolverMap) Get(token module.Token) (any, error) {
	v, ok := m[token]
	if !ok {
		return nil, errors.New("missing token")
	}
	return v, nil
}

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
	cfgModule := configmodule.NewModule(configmodule.Options{})
	mod := NewModule(Options{Config: cfgModule})
	def := mod.(*Module).Definition()
	if def.Name != "database" {
		t.Fatalf("expected name database, got %q", def.Name)
	}
	if len(def.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(def.Providers))
	}
	if len(def.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(def.Imports))
	}
	if def.Providers[0].Token != TokenDB {
		t.Fatalf("expected TokenDB, got %q", def.Providers[0].Token)
	}
	if def.Providers[0].Cleanup == nil {
		t.Fatal("expected cleanup hook")
	}
}

func TestDatabaseModule_ProviderBuildError(t *testing.T) {
	mod := NewModule(Options{Config: configmodule.NewModule(configmodule.Options{})})
	def := mod.(*Module).Definition()
	provider := def.Providers[0]

	_, err := provider.Build(resolverMap{configmodule.TokenMySQLDSN: ""})
	if err == nil {
		t.Fatal("expected error for empty DSN")
	}
	if err.Error() != "mysql dsn is required" {
		t.Fatalf("expected 'mysql dsn is required' error, got %q", err.Error())
	}
}
