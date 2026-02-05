package kernel

import (
	"errors"
	"fmt"

	"github.com/go-modkit/modkit/modkit/module"
)

// ErrExportAmbiguous is returned when a module tries to re-export a token that multiple imports provide.
var ErrExportAmbiguous = errors.New("export token is ambiguous across imports")

// ErrNilGraph is returned when BuildVisibility is called with a nil graph.
var ErrNilGraph = errors.New("graph is nil")

// RootModuleNilError is returned when Bootstrap is called with a nil root module.
type RootModuleNilError struct{}

func (e *RootModuleNilError) Error() string {
	return "root module is nil"
}

// InvalidModuleNameError is returned when a module has an empty or invalid name.
type InvalidModuleNameError struct {
	Name string
}

func (e *InvalidModuleNameError) Error() string {
	return fmt.Sprintf("invalid module name: %q", e.Name)
}

// ModuleNotPointerError is returned when a module is not passed by pointer.
type ModuleNotPointerError struct {
	Module string
}

func (e *ModuleNotPointerError) Error() string {
	return fmt.Sprintf("module must be a pointer: %q", e.Module)
}

// InvalidModuleDefError is returned when a module's Definition() returns invalid metadata.
type InvalidModuleDefError struct {
	Module string
	Reason string
}

func (e *InvalidModuleDefError) Error() string {
	return fmt.Sprintf("invalid module definition: module=%q reason=%s", e.Module, e.Reason)
}

func (e *InvalidModuleDefError) Unwrap() error {
	return module.ErrInvalidModuleDef
}

// NilImportError is returned when a module has a nil entry in its Imports slice.
type NilImportError struct {
	Module string
	Index  int
}

func (e *NilImportError) Error() string {
	return fmt.Sprintf("nil import: module=%q index=%d", e.Module, e.Index)
}

// DuplicateModuleNameError is returned when multiple modules have the same name.
type DuplicateModuleNameError struct {
	Name string
}

func (e *DuplicateModuleNameError) Error() string {
	return fmt.Sprintf("duplicate module name: %s", e.Name)
}

// ModuleCycleError is returned when a circular dependency exists in module imports.
type ModuleCycleError struct {
	Path []string
}

func (e *ModuleCycleError) Error() string {
	return fmt.Sprintf("module cycle detected: %v", e.Path)
}

// DuplicateProviderTokenError is returned when the same provider token is registered in multiple modules.
type DuplicateProviderTokenError struct {
	Token   module.Token
	Modules []string
}

func (e *DuplicateProviderTokenError) Error() string {
	if len(e.Modules) == 2 {
		return fmt.Sprintf("duplicate provider token: %q (modules %q, %q)", e.Token, e.Modules[0], e.Modules[1])
	}
	return fmt.Sprintf("duplicate provider token: %q", e.Token)
}

// DuplicateControllerNameError is returned when a module has multiple controllers with the same name.
type DuplicateControllerNameError struct {
	Module string
	Name   string
}

func (e *DuplicateControllerNameError) Error() string {
	return fmt.Sprintf("duplicate controller name in module %q: %s", e.Module, e.Name)
}

// TokenNotVisibleError is returned when a module attempts to resolve a token that isn't visible to it.
type TokenNotVisibleError struct {
	Module string
	Token  module.Token
}

func (e *TokenNotVisibleError) Error() string {
	return fmt.Sprintf("token not visible: module=%q token=%q", e.Module, e.Token)
}

// ExportNotVisibleError is returned when a module exports a token it cannot access.
type ExportNotVisibleError struct {
	Module string
	Token  module.Token
}

func (e *ExportNotVisibleError) Error() string {
	return fmt.Sprintf("export not visible: module=%q token=%q", e.Module, e.Token)
}

// ExportAmbiguousError is returned when a module re-exports a token that multiple imports provide.
type ExportAmbiguousError struct {
	Module  string
	Token   module.Token
	Imports []string
}

func (e *ExportAmbiguousError) Error() string {
	return fmt.Sprintf("export token %q in module %q is exported by multiple imports: %v", e.Token, e.Module, e.Imports)
}

func (e *ExportAmbiguousError) Unwrap() error {
	return ErrExportAmbiguous
}

// ProviderNotFoundError is returned when attempting to resolve a token that has no registered provider.
type ProviderNotFoundError struct {
	Module string
	Token  module.Token
}

func (e *ProviderNotFoundError) Error() string {
	if e.Module == "" {
		return fmt.Sprintf("provider not found: token=%q", e.Token)
	}
	return fmt.Sprintf("provider not found: module=%q token=%q", e.Module, e.Token)
}

// ProviderCycleError is returned when a circular dependency exists in provider resolution.
type ProviderCycleError struct {
	Token module.Token
}

func (e *ProviderCycleError) Error() string {
	return fmt.Sprintf("provider cycle detected: token=%q", e.Token)
}

// ProviderBuildError wraps an error that occurred while building a provider instance.
type ProviderBuildError struct {
	Module string
	Token  module.Token
	Err    error
}

func (e *ProviderBuildError) Error() string {
	return fmt.Sprintf("provider build failed: module=%q token=%q: %v", e.Module, e.Token, e.Err)
}

func (e *ProviderBuildError) Unwrap() error {
	return e.Err
}

// ControllerBuildError wraps an error that occurred while building a controller instance.
type ControllerBuildError struct {
	Module     string
	Controller string
	Err        error
}

func (e *ControllerBuildError) Error() string {
	return fmt.Sprintf("controller build failed: module=%q controller=%q: %v", e.Module, e.Controller, e.Err)
}

func (e *ControllerBuildError) Unwrap() error {
	return e.Err
}
