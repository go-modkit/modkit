package module_test

import (
	"testing"

	"github.com/go-modkit/modkit/modkit/module"
)

// Compile-only assertions for exported types and errors.
func TestExports(_ *testing.T) {
	var _ module.Token
	var _ module.Resolver
	var _ module.ProviderDef
	var _ module.ControllerDef
	var _ module.ModuleDef
	var _ module.Module
	_ = module.ErrInvalidModuleDef
}
