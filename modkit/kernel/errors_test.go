package kernel

import (
	"errors"
	"testing"
)

func TestKernelErrorStrings(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"NilGraph", ErrNilGraph},
		{"NilApp", ErrNilApp},
		{"GraphNodeNotFound", ErrGraphNodeNotFound},
		{"UnsupportedGraphFormat", &UnsupportedGraphFormatError{Format: GraphFormat("json")}},
		{"GraphNodeNotFoundTyped", &GraphNodeNotFoundError{Node: "missing"}},
		{"RootModuleNil", &RootModuleNilError{}},
		{"InvalidModuleName", &InvalidModuleNameError{Name: "mod"}},
		{"ModuleNotPointer", &ModuleNotPointerError{Module: "mod"}},
		{"InvalidModuleDef", &InvalidModuleDefError{Module: "mod", Reason: "bad"}},
		{"NilImport", &NilImportError{Module: "mod", Index: 1}},
		{"DuplicateModuleName", &DuplicateModuleNameError{Name: "mod"}},
		{"ModuleCycle", &ModuleCycleError{Path: []string{"a", "b"}}},
		{"DuplicateProviderToken", &DuplicateProviderTokenError{Token: "t", Modules: []string{"a", "b"}}},
		{"DuplicateControllerName", &DuplicateControllerNameError{Module: "mod", Name: "ctrl"}},
		{"TokenNotVisible", &TokenNotVisibleError{Module: "mod", Token: "t"}},
		{"ExportNotVisible", &ExportNotVisibleError{Module: "mod", Token: "t"}},
		{"ExportAmbiguous", &ExportAmbiguousError{Module: "mod", Token: "t", Imports: []string{"a", "b"}}},
		{"ProviderNotFound", &ProviderNotFoundError{Module: "mod", Token: "t"}},
		{"ProviderCycle", &ProviderCycleError{Token: "t"}},
		{"ProviderBuild", &ProviderBuildError{Module: "mod", Token: "t", Err: errors.New("boom")}},
		{"ControllerBuild", &ControllerBuildError{Module: "mod", Controller: "c", Err: errors.New("boom")}},
		{"OverrideTokenNotFound", &OverrideTokenNotFoundError{Token: "t"}},
		{"OverrideTokenNotVisibleFromRoot", &OverrideTokenNotVisibleFromRootError{Root: "root", Token: "t"}},
		{"DuplicateOverrideToken", &DuplicateOverrideTokenError{Token: "t"}},
		{"BootstrapOptionConflict", &BootstrapOptionConflictError{Token: "t", Options: []string{"a", "b"}}},
		{"NilBootstrapOption", &NilBootstrapOptionError{Index: 0}},
		{"OverrideBuildNil", &OverrideBuildNilError{Token: "t"}},
	}
	for _, tc := range tests {
		if tc.err == nil {
			t.Fatalf("%s produced nil error", tc.name)
		}
		if tc.err.Error() == "" {
			t.Fatalf("%s produced empty error string", tc.name)
		}
	}
}

func TestErrorWraps(t *testing.T) {
	inner := errors.New("inner")
	err := &ProviderBuildError{Module: "m", Token: "t", Err: inner}
	if !errors.Is(err.Unwrap(), inner) {
		t.Fatalf("expected unwrap to return inner error, got %v", err.Unwrap())
	}
	err2 := &ControllerBuildError{Module: "m", Controller: "c", Err: inner}
	if !errors.Is(err2.Unwrap(), inner) {
		t.Fatalf("expected unwrap to return inner error, got %v", err2.Unwrap())
	}
	ambiguous := &ExportAmbiguousError{Module: "m", Token: "t", Imports: []string{"a", "b"}}
	if !errors.Is(ambiguous, ErrExportAmbiguous) {
		t.Fatalf("expected ExportAmbiguousError to unwrap to ErrExportAmbiguous")
	}
	nodeMissing := &GraphNodeNotFoundError{Node: "x"}
	if !errors.Is(nodeMissing, ErrGraphNodeNotFound) {
		t.Fatalf("expected GraphNodeNotFoundError to unwrap to ErrGraphNodeNotFound")
	}
}
