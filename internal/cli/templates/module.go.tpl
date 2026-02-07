package {{.Package}}

import "github.com/go-modkit/modkit/modkit/module"

// {{.Identifier}}Module is the {{.Name}} module.
type {{.Identifier}}Module struct{}

func (m *{{.Identifier}}Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name:        "{{.Name}}",
		Providers:   []module.ProviderDef{},
		Controllers: []module.ControllerDef{},
	}
}
