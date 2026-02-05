package kernel

import (
	"testing"

	"github.com/go-modkit/modkit/modkit/module"
)

type modHelperInternal struct {
	def module.ModuleDef
}

func (m *modHelperInternal) Definition() module.ModuleDef {
	return m.def
}

func modInternal(
	name string,
	imports []module.Module,
	providers []module.ProviderDef,
	controllers []module.ControllerDef,
	exports []module.Token,
) module.Module {
	return &modHelperInternal{
		def: module.ModuleDef{
			Name:        name,
			Imports:     imports,
			Providers:   providers,
			Controllers: controllers,
			Exports:     exports,
		},
	}
}

func TestContainerRecordsProviderBuildOrder(t *testing.T) {
	first := module.Token("provider.first")
	second := module.Token("provider.second")

	modA := modInternal("A", nil,
		[]module.ProviderDef{{
			Token: first,
			Build: func(_ module.Resolver) (any, error) {
				return "first", nil
			},
		}, {
			Token: second,
			Build: func(_ module.Resolver) (any, error) {
				return "second", nil
			},
		}},
		nil,
		nil,
	)

	app, err := Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if _, err := app.Get(second); err != nil {
		t.Fatalf("Get second failed: %v", err)
	}
	if _, err := app.Get(first); err != nil {
		t.Fatalf("Get first failed: %v", err)
	}

	order := app.container.providerBuildOrder()
	if len(order) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(order))
	}
	if order[0] != second || order[1] != first {
		t.Fatalf("unexpected order: %v", order)
	}
}

func TestContainerRecordsClosersInBuildOrder(t *testing.T) {
	closerA := module.Token("closer.a")
	closerB := module.Token("closer.b")

	modA := modInternal("A", nil,
		[]module.ProviderDef{{
			Token: closerA,
			Build: func(_ module.Resolver) (any, error) {
				return &testCloser{name: "a"}, nil
			},
		}, {
			Token: closerB,
			Build: func(_ module.Resolver) (any, error) {
				return &testCloser{name: "b"}, nil
			},
		}},
		nil,
		nil,
	)

	app, err := Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	_, _ = app.Get(closerA)
	_, _ = app.Get(closerB)

	closers := app.container.closersInBuildOrder()
	if len(closers) != 2 {
		t.Fatalf("expected 2 closers, got %d", len(closers))
	}

	first, ok := closers[0].(*testCloser)
	if !ok {
		t.Fatalf("expected *testCloser, got %T", closers[0])
	}
	second, ok := closers[1].(*testCloser)
	if !ok {
		t.Fatalf("expected *testCloser, got %T", closers[1])
	}
	if first.Name() != "a" || second.Name() != "b" {
		t.Fatalf("unexpected order: %v", closers)
	}
}

type testCloser struct {
	name string
}

func (c *testCloser) Close() error {
	return nil
}

func (c *testCloser) Name() string {
	return c.name
}
