package module

// ModuleDef declares metadata for a module.
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
