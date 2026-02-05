package module

// ModuleDef declares metadata for a module, including its dependencies,
// providers, controllers, and exported tokens. The name should be unique
// within the module graph.
//
//nolint:revive // Intentional API name for clarity
type ModuleDef struct {
	Name        string
	Imports     []Module
	Providers   []ProviderDef
	Controllers []ControllerDef
	Exports     []Token
}

// Module provides its definition for graph construction.
type Module interface {
	Definition() ModuleDef
}
