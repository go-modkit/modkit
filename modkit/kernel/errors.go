package kernel

import (
	"fmt"

	"github.com/aryeko/modkit/modkit/module"
)

type RootModuleNilError struct{}

func (e *RootModuleNilError) Error() string {
	return "root module is nil"
}

type InvalidModuleNameError struct {
	Name string
}

func (e *InvalidModuleNameError) Error() string {
	return "invalid module name"
}

type DuplicateModuleNameError struct {
	Name string
}

func (e *DuplicateModuleNameError) Error() string {
	return fmt.Sprintf("duplicate module name: %s", e.Name)
}

type ModuleCycleError struct {
	Path []string
}

func (e *ModuleCycleError) Error() string {
	return fmt.Sprintf("module cycle detected: %v", e.Path)
}

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

type DuplicateControllerNameError struct {
	Name string
}

func (e *DuplicateControllerNameError) Error() string {
	return fmt.Sprintf("duplicate controller name: %s", e.Name)
}

type TokenNotVisibleError struct {
	Module string
	Token  module.Token
}

func (e *TokenNotVisibleError) Error() string {
	return fmt.Sprintf("token not visible: module=%q token=%q", e.Module, e.Token)
}

type ExportNotVisibleError struct {
	Module string
	Token  module.Token
}

func (e *ExportNotVisibleError) Error() string {
	return fmt.Sprintf("export not visible: module=%q token=%q", e.Module, e.Token)
}

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

type ProviderCycleError struct {
	Token module.Token
}

func (e *ProviderCycleError) Error() string {
	return fmt.Sprintf("provider cycle detected: token=%q", e.Token)
}

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
