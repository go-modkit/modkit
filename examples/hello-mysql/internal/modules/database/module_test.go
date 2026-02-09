package database

import (
	"context"
	"errors"
	"testing"

	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
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
	if len(def.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(def.Providers))
	}
	if len(def.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(def.Imports))
	}

	var dbProvider, dialectProvider *module.ProviderDef
	for i := range def.Providers {
		p := &def.Providers[i]
		switch p.Token {
		case TokenDB:
			dbProvider = p
		case TokenDialect:
			dialectProvider = p
		}
	}
	if dbProvider == nil {
		t.Fatalf("expected provider %q", TokenDB)
	}
	if dialectProvider == nil {
		t.Fatalf("expected provider %q", TokenDialect)
	}
	if dbProvider.Cleanup == nil {
		t.Fatal("expected cleanup hook")
	}

	dialect, err := dialectProvider.Build(resolverMap{})
	if err != nil {
		t.Fatalf("expected dialect build to succeed, got %v", err)
	}
	if dialect != sqlmodule.DialectMySQL {
		t.Fatalf("expected dialect %q, got %v", sqlmodule.DialectMySQL, dialect)
	}

	exports := map[module.Token]bool{}
	for _, token := range def.Exports {
		exports[token] = true
	}
	if !exports[TokenDB] {
		t.Fatalf("expected export %q", TokenDB)
	}
	if !exports[TokenDialect] {
		t.Fatalf("expected export %q", TokenDialect)
	}
}

func TestDatabaseModule_TokenDB_CompatibilityWithSQLContract(t *testing.T) {
	if TokenDB != sqlmodule.TokenDB {
		t.Fatalf("TokenDB = %q, want %q", TokenDB, sqlmodule.TokenDB)
	}
	if TokenDB != module.Token("database.db") {
		t.Fatalf("TokenDB = %q, want %q", TokenDB, module.Token("database.db"))
	}
}

func TestDatabaseModule_ProviderBuildError(t *testing.T) {
	mod := NewModule(Options{Config: configmodule.NewModule(configmodule.Options{})})
	def := mod.(*Module).Definition()

	var provider module.ProviderDef
	found := false
	for _, p := range def.Providers {
		if p.Token == TokenDB {
			provider = p
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected provider %q", TokenDB)
	}

	_, err := provider.Build(resolverMap{configmodule.TokenMySQLDSN: ""})
	if err == nil {
		t.Fatal("expected error for empty DSN")
	}
	if err.Error() != "mysql dsn is required" {
		t.Fatalf("expected 'mysql dsn is required' error, got %q", err.Error())
	}
}
