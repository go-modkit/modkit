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
			Build: func(r module.Resolver) (any, error) {
				return "first", nil
			},
		}, {
			Token: second,
			Build: func(r module.Resolver) (any, error) {
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

	_, _ = app.Get(second)
	_, _ = app.Get(first)

	order := app.container.providerBuildOrder()
	if len(order) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(order))
	}
	if order[0] != second || order[1] != first {
		t.Fatalf("unexpected order: %v", order)
	}
}
