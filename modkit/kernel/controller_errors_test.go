package kernel_test

import (
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

func TestBootstrap_ReturnsControllerBuildError(t *testing.T) {
	buildErr := errors.New("controller boom")
	controller := module.ControllerDef{
		Name: "BadController",
		Build: func(module.Resolver) (any, error) {
			return nil, buildErr
		},
	}
	mod := mod("App", nil, nil, []module.ControllerDef{controller}, nil)

	_, err := kernel.Bootstrap(mod)
	if err == nil {
		t.Fatal("expected error")
	}

	var errType *kernel.ControllerBuildError
	if !errors.As(err, &errType) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if errType.Module != "App" || errType.Controller != "BadController" {
		t.Fatalf("unexpected error fields")
	}
	if !errors.Is(err, buildErr) {
		t.Fatalf("expected wrapped error")
	}
}

func TestBootstrap_ReturnsControllerBuildErrorOnMissingDependency(t *testing.T) {
	missing := module.Token("missing")
	controller := module.ControllerDef{
		Name: "BadController",
		Build: func(r module.Resolver) (any, error) {
			_, err := r.Get(missing)
			if err != nil {
				return nil, err
			}
			return "ok", nil
		},
	}
	mod := mod("App", nil, nil, []module.ControllerDef{controller}, nil)

	_, err := kernel.Bootstrap(mod)
	if err == nil {
		t.Fatal("expected error")
	}

	var errType *kernel.ControllerBuildError
	if !errors.As(err, &errType) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if errType.Module != "App" || errType.Controller != "BadController" {
		t.Fatalf("unexpected error fields")
	}

	var notVisible *kernel.TokenNotVisibleError
	if !errors.As(err, &notVisible) {
		t.Fatalf("expected TokenNotVisibleError, got %T", err)
	}
	if notVisible.Module != "App" || notVisible.Token != missing {
		t.Fatalf("unexpected nested error fields")
	}
}
