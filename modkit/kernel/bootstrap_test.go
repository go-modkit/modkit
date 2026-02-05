package kernel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
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

	if app.Controllers["A:UsesShared"] != "value" {
		t.Fatalf("unexpected controller value: %v", app.Controllers["A:UsesShared"])
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
	if dupErr.Module != "A" {
		t.Fatalf("unexpected module name: %q", dupErr.Module)
	}
}

func TestBootstrap_CollectsCleanupHooksInLIFO(t *testing.T) {
	tokenB := module.Token("test.tokenB")
	tokenA := module.Token("test.tokenA")
	calls := make([]string, 0, 2)

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: tokenB,
			Build: func(r module.Resolver) (any, error) {
				return "b", nil
			},
			Cleanup: func(ctx context.Context) error {
				calls = append(calls, "B")
				return nil
			},
		}, {
			Token: tokenA,
			Build: func(r module.Resolver) (any, error) {
				_, err := r.Get(tokenB)
				if err != nil {
					return nil, err
				}
				return "a", nil
			},
			Cleanup: func(ctx context.Context) error {
				calls = append(calls, "A")
				return nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if _, err := app.Get(tokenA); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	hooks := app.CleanupHooks()
	if len(hooks) != 2 {
		t.Fatalf("expected 2 cleanup hooks, got %d", len(hooks))
	}

	for _, hook := range hooks {
		if err := hook(context.Background()); err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 cleanup calls, got %d", len(calls))
	}
	if calls[0] != "A" || calls[1] != "B" {
		t.Fatalf("unexpected cleanup order: %v", calls)
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

	if app.Controllers["A:ControllerA"] != "controller" {
		t.Fatalf("expected controller instance to be registered")
	}
}

func TestBootstrapAllowsSameControllerNameAcrossModules(t *testing.T) {
	modB := mod("B", nil, nil,
		[]module.ControllerDef{{
			Name: "Shared",
			Build: func(r module.Resolver) (any, error) {
				return "controller-from-B", nil
			},
		}},
		nil,
	)

	modA := mod("A", []module.Module{modB}, nil,
		[]module.ControllerDef{{
			Name: "Shared",
			Build: func(r module.Resolver) (any, error) {
				return "controller-from-A", nil
			},
		}},
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(app.Controllers) != 2 {
		t.Fatalf("expected 2 controllers, got %d", len(app.Controllers))
	}

	if app.Controllers["A:Shared"] != "controller-from-A" {
		t.Errorf("controller A:Shared has wrong value: %v", app.Controllers["A:Shared"])
	}
	if app.Controllers["B:Shared"] != "controller-from-B" {
		t.Errorf("controller B:Shared has wrong value: %v", app.Controllers["B:Shared"])
	}
}

func TestBootstrapRejectsDuplicateControllerInSameModule(t *testing.T) {
	modA := mod("A", nil, nil,
		[]module.ControllerDef{
			{
				Name: "Dup",
				Build: func(r module.Resolver) (any, error) {
					return "one", nil
				},
			},
			{
				Name: "Dup",
				Build: func(r module.Resolver) (any, error) {
					return "two", nil
				},
			},
		},
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
	if dupErr.Name != "Dup" {
		t.Fatalf("unexpected controller name: %q", dupErr.Name)
	}
	if dupErr.Module != "A" {
		t.Fatalf("unexpected module name: %q", dupErr.Module)
	}
}

func TestControllerKeyFormat(t *testing.T) {
	modA := mod("users", nil, nil,
		[]module.ControllerDef{{
			Name: "Controller",
			Build: func(r module.Resolver) (any, error) {
				return nil, nil
			},
		}},
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if _, ok := app.Controllers["users:Controller"]; !ok {
		t.Fatalf("expected key 'users:Controller', got keys: %v", controllerKeys(app.Controllers))
	}
}

func controllerKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
