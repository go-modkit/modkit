package auth

import (
	"errors"
	"testing"

	configmodule "github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/config"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/database"
	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
	"github.com/go-modkit/modkit/modkit/testkit"
)

func TestModule_Bootstrap(t *testing.T) {
	mod := NewModule(Options{})
	_, err := kernel.Bootstrap(mod)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
}

func TestAuthModule_Definition(t *testing.T) {
	def := NewModule(Options{}).(*Module).Definition()

	if def.Name != "auth" {
		t.Fatalf("name = %q", def.Name)
	}
	if len(def.Controllers) != 1 {
		t.Fatalf("controllers = %d", len(def.Controllers))
	}
}

type errorResolver struct {
	token module.Token
	err   error
}

func (r errorResolver) Get(token module.Token) (any, error) {
	if token == r.token {
		return nil, r.err
	}
	return nil, nil
}

func TestAuthModule_ControllerBuildError(t *testing.T) {
	def := NewModule(Options{}).(*Module).Definition()

	_, err := def.Controllers[0].Build(errorResolver{
		token: TokenHandler,
		err:   errors.New("boom"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
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

func TestAuthAndDatabase_DefaultConfigComposition(t *testing.T) {
	root := &rootModule{imports: []module.Module{
		NewModule(Options{}),
		database.NewModule(database.Options{}),
	}}

	if _, err := kernel.Bootstrap(root); err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
}

func TestModule_TestKitOverrideConfigTokens(t *testing.T) {
	h := testkit.New(t,
		NewModule(Options{}),
		testkit.WithOverrides(
			testkit.OverrideValue(configmodule.TokenAuthUsername, "override-user"),
			testkit.OverrideValue(configmodule.TokenAuthPassword, "override-pass"),
		),
	)

	handler := testkit.Get[*Handler](t, h, TokenHandler)
	if handler.cfg.Username != "override-user" {
		t.Fatalf("username = %q", handler.cfg.Username)
	}
	if handler.cfg.Password != "override-pass" {
		t.Fatalf("password = %q", handler.cfg.Password)
	}
}
