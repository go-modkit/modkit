package kernel_test

import (
	"errors"
	"testing"

	"github.com/aryeko/modkit/modkit/kernel"
	"github.com/aryeko/modkit/modkit/module"
)

func TestBootstrapEnforcesVisibility(t *testing.T) {
	secretToken := module.Token("secret")

	modB := mod("B", nil,
		[]module.ProviderDef{{
			Token: secretToken,
			Build: func(r module.Resolver) (any, error) {
				return "shh", nil
			},
		}},
		nil,
		nil,
	)

	modA := mod("A", []module.Module{modB}, nil,
		[]module.ControllerDef{{
			Name: "NeedsSecret",
			Build: func(r module.Resolver) (any, error) {
				_, err := r.Get(secretToken)
				return nil, err
			},
		}},
		nil,
	)

	_, err := kernel.Bootstrap(modA)
	if err == nil {
		t.Fatalf("expected visibility error")
	}

	var visErr *kernel.TokenNotVisibleError
	if !errors.As(err, &visErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if visErr.Module != "A" {
		t.Fatalf("unexpected module: %q", visErr.Module)
	}
	if visErr.Token != secretToken {
		t.Fatalf("unexpected token: %q", visErr.Token)
	}
}

func TestBootstrapRejectsInvalidExport(t *testing.T) {
	missing := module.Token("missing")

	modA := mod("A", nil, nil, nil, []module.Token{missing})

	_, err := kernel.Bootstrap(modA)
	if err == nil {
		t.Fatalf("expected export validation error")
	}

	var exportErr *kernel.ExportNotVisibleError
	if !errors.As(err, &exportErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if exportErr.Module != "A" {
		t.Fatalf("unexpected module: %q", exportErr.Module)
	}
	if exportErr.Token != missing {
		t.Fatalf("unexpected token: %q", exportErr.Token)
	}
}

func TestBootstrapAllowsReExportedTokens(t *testing.T) {
	shared := module.Token("shared")

	modC := mod("C", nil,
		[]module.ProviderDef{{
			Token: shared,
			Build: func(r module.Resolver) (any, error) {
				return "value", nil
			},
		}},
		nil,
		[]module.Token{shared},
	)

	modB := mod("B", []module.Module{modC}, nil, nil, []module.Token{shared})

	modA := mod("A", []module.Module{modB}, nil,
		[]module.ControllerDef{{
			Name: "UsesShared",
			Build: func(r module.Resolver) (any, error) {
				return r.Get(shared)
			},
		}},
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if app.Controllers["UsesShared"] != "value" {
		t.Fatalf("unexpected controller value: %v", app.Controllers["UsesShared"])
	}
}

func TestBootstrapRejectsDuplicateProviderTokens(t *testing.T) {
	shared := module.Token("shared")

	modB := mod("B", nil,
		[]module.ProviderDef{{
			Token: shared,
			Build: func(r module.Resolver) (any, error) { return "b", nil },
		}},
		nil,
		nil,
	)

	modA := mod("A", []module.Module{modB},
		[]module.ProviderDef{{
			Token: shared,
			Build: func(r module.Resolver) (any, error) { return "a", nil },
		}},
		nil,
		nil,
	)

	_, err := kernel.Bootstrap(modA)
	if err == nil {
		t.Fatalf("expected duplicate provider error")
	}

	var dupErr *kernel.DuplicateProviderTokenError
	if !errors.As(err, &dupErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if dupErr.Token != shared {
		t.Fatalf("unexpected token: %q", dupErr.Token)
	}
}

func TestBootstrapRejectsDuplicateControllerNames(t *testing.T) {
	modA := mod("A", nil, nil,
		[]module.ControllerDef{{
			Name: "ControllerA",
			Build: func(r module.Resolver) (any, error) {
				return "one", nil
			},
		}, {
			Name: "ControllerA",
			Build: func(r module.Resolver) (any, error) {
				return "two", nil
			},
		}},
		nil,
	)

	_, err := kernel.Bootstrap(modA)
	if err == nil {
		t.Fatalf("expected duplicate controller error")
	}

	var dupErr *kernel.DuplicateControllerNameError
	if !errors.As(err, &dupErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if dupErr.Name != "ControllerA" {
		t.Fatalf("unexpected controller name: %q", dupErr.Name)
	}
}

func TestBootstrapRegistersControllers(t *testing.T) {
	modA := mod("A", nil, nil,
		[]module.ControllerDef{{
			Name: "ControllerA",
			Build: func(r module.Resolver) (any, error) {
				return "controller", nil
			},
		}},
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if app.Controllers["ControllerA"] != "controller" {
		t.Fatalf("expected controller instance to be registered")
	}
}
