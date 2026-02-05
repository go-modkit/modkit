package kernel_test

import "github.com/go-modkit/modkit/modkit/module"

type testModule struct {
	def module.ModuleDef
}

func (m *testModule) Definition() module.ModuleDef {
	return m.def
}

type valueModule struct {
	def module.ModuleDef
}

//nolint:gocritic // Intentionally uses value receiver to test module validation
func (m valueModule) Definition() module.ModuleDef {
	return m.def
}

type modHelper struct {
	def module.ModuleDef
}

func (m *modHelper) Definition() module.ModuleDef {
	return m.def
}

func mod(
	name string,
	imports []module.Module,
	providers []module.ProviderDef,
	controllers []module.ControllerDef,
	exports []module.Token,
) module.Module {
	return &modHelper{
		def: module.ModuleDef{
			Name:        name,
			Imports:     imports,
			Providers:   providers,
			Controllers: controllers,
			Exports:     exports,
		},
	}
}
