package {{.Package}}

import "github.com/go-modkit/modkit/modkit/module"

// {{.Name | Title | Replace " " ""}}Module is the {{.Name}} module.
type {{.Name | Title | Replace " " ""}}Module struct{}

func (m *{{.Name | Title | Replace " " ""}}Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name:        "{{.Name}}",
		Providers:   []module.ProviderDef{},
		Controllers: []module.ControllerDef{},
	}
}
