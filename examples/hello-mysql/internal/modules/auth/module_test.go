package auth

import (
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
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
