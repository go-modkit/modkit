package kernel_test

import (
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

func TestBootstrapWithOptions_NoOptionsParity(t *testing.T) {
	token := module.Token("svc.token")

	root := mod("root", nil,
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) {
				return "value", nil
			},
		}},
		[]module.ControllerDef{{
			Name: "Controller",
			Build: func(r module.Resolver) (any, error) {
				return r.Get(token)
			},
		}},
		[]module.Token{token},
	)

	appA, err := kernel.Bootstrap(root)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	appB, err := kernel.BootstrapWithOptions(root)
	if err != nil {
		t.Fatalf("BootstrapWithOptions failed: %v", err)
	}

	va, err := appA.Get(token)
	if err != nil {
		t.Fatalf("appA.Get failed: %v", err)
	}
	vb, err := appB.Get(token)
	if err != nil {
		t.Fatalf("appB.Get failed: %v", err)
	}

	if va != vb {
		t.Fatalf("expected equal values, got %v vs %v", va, vb)
	}

	if appA.Controllers["root:Controller"] != appB.Controllers["root:Controller"] {
		t.Fatalf("expected controller parity")
	}
}

func TestBootstrapWithOptions_OverrideValue(t *testing.T) {
	token := module.Token("svc.token")

	root := mod("root", nil,
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) {
				return "real", nil
			},
		}},
		nil,
		[]module.Token{token},
	)

	app, err := kernel.BootstrapWithOptions(root,
		kernel.WithProviderOverrides(kernel.ProviderOverride{
			Token: token,
			Build: func(module.Resolver) (any, error) {
				return "fake", nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("BootstrapWithOptions failed: %v", err)
	}

	v, err := app.Get(token)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if v != "fake" {
		t.Fatalf("expected fake override value, got %v", v)
	}
}

func TestBootstrapWithOptions_OverrideBuildUsesOriginalOwnerVisibility(t *testing.T) {
	hidden := module.Token("b.hidden")
	overrideTarget := module.Token("a.target")

	modB := mod("B", nil,
		[]module.ProviderDef{{
			Token: hidden,
			Build: func(module.Resolver) (any, error) {
				return "secret", nil
			},
		}},
		nil,
		nil,
	)

	modA := mod("A", []module.Module{modB},
		[]module.ProviderDef{{
			Token: overrideTarget,
			Build: func(module.Resolver) (any, error) {
				return "original", nil
			},
		}},
		nil,
		[]module.Token{overrideTarget},
	)

	app, err := kernel.BootstrapWithOptions(modA,
		kernel.WithProviderOverrides(kernel.ProviderOverride{
			Token: overrideTarget,
			Build: func(r module.Resolver) (any, error) {
				return r.Get(hidden)
			},
		}),
	)
	if err != nil {
		t.Fatalf("BootstrapWithOptions failed: %v", err)
	}

	_, err = app.Get(overrideTarget)
	if err == nil {
		t.Fatal("expected visibility error")
	}

	var visErr *kernel.TokenNotVisibleError
	if !errors.As(err, &visErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if visErr.Module != "A" || visErr.Token != hidden {
		t.Fatalf("unexpected visibility error: %v", visErr)
	}
}

func TestBootstrapWithOptions_RejectsDuplicateOverrideToken(t *testing.T) {
	token := module.Token("svc.token")
	root := mod("root", nil,
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) { return "real", nil },
		}},
		nil,
		[]module.Token{token},
	)

	_, err := kernel.BootstrapWithOptions(root,
		kernel.WithProviderOverrides(
			kernel.ProviderOverride{Token: token, Build: func(module.Resolver) (any, error) { return "a", nil }},
			kernel.ProviderOverride{Token: token, Build: func(module.Resolver) (any, error) { return "b", nil }},
		),
	)
	if err == nil {
		t.Fatal("expected duplicate override token error")
	}

	var dupErr *kernel.DuplicateOverrideTokenError
	if !errors.As(err, &dupErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestBootstrapWithOptions_RejectsUnknownOverrideToken(t *testing.T) {
	root := mod("root", nil, nil, nil, nil)

	_, err := kernel.BootstrapWithOptions(root,
		kernel.WithProviderOverrides(kernel.ProviderOverride{
			Token: module.Token("missing"),
			Build: func(module.Resolver) (any, error) { return "x", nil },
		}),
	)
	if err == nil {
		t.Fatal("expected unknown override token error")
	}

	var notFoundErr *kernel.OverrideTokenNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestBootstrapWithOptions_RejectsOverrideTokenNotVisibleFromRoot(t *testing.T) {
	hidden := module.Token("hidden")

	modB := mod("B", nil,
		[]module.ProviderDef{{
			Token: hidden,
			Build: func(module.Resolver) (any, error) { return "secret", nil },
		}},
		nil,
		nil,
	)

	modA := mod("A", []module.Module{modB}, nil, nil, nil)

	_, err := kernel.BootstrapWithOptions(modA,
		kernel.WithProviderOverrides(kernel.ProviderOverride{
			Token: hidden,
			Build: func(module.Resolver) (any, error) { return "fake", nil },
		}),
	)
	if err == nil {
		t.Fatal("expected root visibility override error")
	}

	var visErr *kernel.OverrideTokenNotVisibleFromRootError
	if !errors.As(err, &visErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if visErr.Root != "A" || visErr.Token != hidden {
		t.Fatalf("unexpected error values: %v", visErr)
	}
}

func TestBootstrapWithOptions_RejectsOptionConflictForSameToken(t *testing.T) {
	token := module.Token("svc.token")
	root := mod("root", nil,
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) { return "real", nil },
		}},
		nil,
		[]module.Token{token},
	)

	_, err := kernel.BootstrapWithOptions(root,
		kernel.WithProviderOverrides(kernel.ProviderOverride{Token: token, Build: func(module.Resolver) (any, error) { return "a", nil }}),
		kernel.WithProviderOverrides(kernel.ProviderOverride{Token: token, Build: func(module.Resolver) (any, error) { return "b", nil }}),
	)
	if err == nil {
		t.Fatal("expected bootstrap option conflict")
	}

	var conflictErr *kernel.BootstrapOptionConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if conflictErr.Token != token {
		t.Fatalf("unexpected token: %q", conflictErr.Token)
	}
}

func TestBootstrapWithOptions_RejectsNilOption(t *testing.T) {
	root := mod("root", nil, nil, nil, nil)

	_, err := kernel.BootstrapWithOptions(root, nil)
	if err == nil {
		t.Fatal("expected nil bootstrap option error")
	}

	var nilOptErr *kernel.NilBootstrapOptionError
	if !errors.As(err, &nilOptErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestBootstrapWithOptions_RejectsNilOverrideBuild(t *testing.T) {
	token := module.Token("svc.token")
	root := mod("root", nil,
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) { return "real", nil },
		}},
		nil,
		[]module.Token{token},
	)

	_, err := kernel.BootstrapWithOptions(root,
		kernel.WithProviderOverrides(kernel.ProviderOverride{Token: token, Build: nil}),
	)
	if err == nil {
		t.Fatal("expected nil override build error")
	}

	var nilBuildErr *kernel.OverrideBuildNilError
	if !errors.As(err, &nilBuildErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}
