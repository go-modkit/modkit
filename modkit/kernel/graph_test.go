package kernel_test

import (
	"errors"
	"testing"

	"github.com/aryeko/modkit/modkit/kernel"
	"github.com/aryeko/modkit/modkit/module"
)

type testModule struct {
	def module.ModuleDef
}

func (m *testModule) Definition() module.ModuleDef {
	return m.def
}

func mod(
	name string,
	imports []module.Module,
	providers []module.ProviderDef,
	controllers []module.ControllerDef,
	exports []module.Token,
) module.Module {
	return &testModule{
		def: module.ModuleDef{
			Name:        name,
			Imports:     imports,
			Providers:   providers,
			Controllers: controllers,
			Exports:     exports,
		},
	}
}

func TestBuildGraphImportsFirst(t *testing.T) {
	modD := mod("D", nil, nil, nil, nil)
	modB := mod("B", []module.Module{modD}, nil, nil, nil)
	modC := mod("C", nil, nil, nil, nil)
	modA := mod("A", []module.Module{modB, modC}, nil, nil, nil)

	g, err := kernel.BuildGraph(modA)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	got := make([]string, 0, len(g.Modules))
	for _, node := range g.Modules {
		got = append(got, node.Name)
	}

	want := []string{"D", "B", "C", "A"}
	if len(got) != len(want) {
		t.Fatalf("unexpected module count: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected order at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestBuildGraphAllowsSharedImports(t *testing.T) {
	shared := mod("Shared", nil, nil, nil, nil)
	modB := mod("B", []module.Module{shared}, nil, nil, nil)
	modC := mod("C", []module.Module{shared}, nil, nil, nil)
	modA := mod("A", []module.Module{modB, modC}, nil, nil, nil)

	_, err := kernel.BuildGraph(modA)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
}

func TestBuildGraphRejectsDuplicateModuleNames(t *testing.T) {
	modB1 := mod("B", nil, nil, nil, nil)
	modB2 := mod("B", nil, nil, nil, nil)
	modA := mod("A", []module.Module{modB1, modB2}, nil, nil, nil)

	_, err := kernel.BuildGraph(modA)
	if err == nil {
		t.Fatalf("expected error for duplicate module names")
	}

	var dupErr *kernel.DuplicateModuleNameError
	if !errors.As(err, &dupErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if dupErr.Name != "B" {
		t.Fatalf("unexpected duplicate name: %q", dupErr.Name)
	}
}

func TestBuildGraphRejectsCycles(t *testing.T) {
	modA := mod("A", nil, nil, nil, nil)
	modB := mod("B", []module.Module{modA}, nil, nil, nil)

	root := modA.(*testModule)
	root.def.Imports = []module.Module{modB}

	_, err := kernel.BuildGraph(modA)
	if err == nil {
		t.Fatalf("expected cycle error")
	}

	var cycleErr *kernel.ModuleCycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if len(cycleErr.Path) == 0 {
		t.Fatalf("expected cycle path")
	}
}
