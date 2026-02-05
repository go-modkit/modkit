package kernel_test

import (
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

func TestBuildGraphRejectsNilRoot(t *testing.T) {
	_, err := kernel.BuildGraph(nil)
	if err == nil {
		t.Fatalf("expected error for nil root")
	}

	var rootErr *kernel.RootModuleNilError
	if !errors.As(err, &rootErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestBuildGraphRejectsTypedNilRoot(t *testing.T) {
	var root *testModule

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for typed nil root")
	}

	var rootErr *kernel.RootModuleNilError
	if !errors.As(err, &rootErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestBuildGraphRejectsEmptyModuleName(t *testing.T) {
	root := mod("", nil, nil, nil, nil)

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for empty module name")
	}

	var defErr *kernel.InvalidModuleDefError
	if !errors.As(err, &defErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if !errors.Is(err, module.ErrInvalidModuleDef) {
		t.Fatalf("expected ErrInvalidModuleDef")
	}
}

func TestBuildGraphRejectsNilImport(t *testing.T) {
	root := mod("A", []module.Module{nil}, nil, nil, nil)

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for nil import")
	}

	var importErr *kernel.NilImportError
	if !errors.As(err, &importErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if importErr.Module != "A" {
		t.Fatalf("unexpected module: %q", importErr.Module)
	}
	if importErr.Index != 0 {
		t.Fatalf("unexpected index: %d", importErr.Index)
	}
}

func TestBuildGraphRejectsTypedNilImport(t *testing.T) {
	var imp *testModule
	root := mod("A", []module.Module{imp}, nil, nil, nil)

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for typed nil import")
	}

	var importErr *kernel.NilImportError
	if !errors.As(err, &importErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if importErr.Module != "A" {
		t.Fatalf("unexpected module: %q", importErr.Module)
	}
	if importErr.Index != 0 {
		t.Fatalf("unexpected index: %d", importErr.Index)
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

func TestBuildGraphRejectsValueRootModule(t *testing.T) {
	root := valueModule{def: module.ModuleDef{Name: "A"}}

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for value module root")
	}

	var ptrErr *kernel.ModuleNotPointerError
	if !errors.As(err, &ptrErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if ptrErr.Module != "A" {
		t.Fatalf("unexpected module: %q", ptrErr.Module)
	}
}

func TestBuildGraphRejectsValueImportModule(t *testing.T) {
	imp := valueModule{def: module.ModuleDef{Name: "B"}}
	root := mod("A", []module.Module{imp}, nil, nil, nil)

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for value module import")
	}

	var ptrErr *kernel.ModuleNotPointerError
	if !errors.As(err, &ptrErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if ptrErr.Module != "B" {
		t.Fatalf("unexpected module: %q", ptrErr.Module)
	}
}

func TestBuildGraphRejectsProviderWithEmptyToken(t *testing.T) {
	root := mod("A", nil, []module.ProviderDef{{}}, nil, nil)

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for provider with empty token")
	}

	var defErr *kernel.InvalidModuleDefError
	if !errors.As(err, &defErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if !errors.Is(err, module.ErrInvalidModuleDef) {
		t.Fatalf("expected ErrInvalidModuleDef")
	}
}

func TestBuildGraphRejectsProviderWithNilBuild(t *testing.T) {
	root := mod("A", nil, []module.ProviderDef{{Token: module.Token("t")}}, nil, nil)

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for provider with nil build")
	}

	var defErr *kernel.InvalidModuleDefError
	if !errors.As(err, &defErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if !errors.Is(err, module.ErrInvalidModuleDef) {
		t.Fatalf("expected ErrInvalidModuleDef")
	}
}

func TestBuildGraphRejectsControllerWithEmptyName(t *testing.T) {
	root := mod("A", nil, nil, []module.ControllerDef{{}}, nil)

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for controller with empty name")
	}

	var defErr *kernel.InvalidModuleDefError
	if !errors.As(err, &defErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if !errors.Is(err, module.ErrInvalidModuleDef) {
		t.Fatalf("expected ErrInvalidModuleDef")
	}
}

func TestBuildGraphRejectsControllerWithNilBuild(t *testing.T) {
	root := mod("A", nil, nil, []module.ControllerDef{{Name: "C"}}, nil)

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for controller with nil build")
	}

	var defErr *kernel.InvalidModuleDefError
	if !errors.As(err, &defErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if !errors.Is(err, module.ErrInvalidModuleDef) {
		t.Fatalf("expected ErrInvalidModuleDef")
	}
}

func TestBuildGraphRejectsEmptyExportToken(t *testing.T) {
	root := mod("A", nil, nil, nil, []module.Token{""})

	_, err := kernel.BuildGraph(root)
	if err == nil {
		t.Fatalf("expected error for empty export token")
	}

	var defErr *kernel.InvalidModuleDefError
	if !errors.As(err, &defErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if !errors.Is(err, module.ErrInvalidModuleDef) {
		t.Fatalf("expected ErrInvalidModuleDef")
	}
}

func TestBuildGraphRejectsCycles(t *testing.T) {
	modA := mod("A", nil, nil, nil, nil)
	modB := mod("B", []module.Module{modA}, nil, nil, nil)

	root := modA.(*modHelper)
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
