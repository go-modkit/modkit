package kernel_test

import (
	"errors"
	"testing"

	"github.com/aryeko/modkit/modkit/kernel"
	"github.com/aryeko/modkit/modkit/module"
)

func TestContainerMissingTokenError(t *testing.T) {
	modA := mod("A", nil, nil, nil, nil)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	_, err = app.Container.Get(module.Token("missing"))
	if err == nil {
		t.Fatalf("expected missing token error")
	}

	var notFound *kernel.ProviderNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if notFound.Token != module.Token("missing") {
		t.Fatalf("unexpected token: %q", notFound.Token)
	}
}
