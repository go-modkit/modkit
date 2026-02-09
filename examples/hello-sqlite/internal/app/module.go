package app

import (
	"github.com/go-modkit/modkit/modkit/data/sqlite"
	"github.com/go-modkit/modkit/modkit/module"
)

type Module struct{}

func NewModule() module.Module {
	return &Module{}
}

func (m *Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name: "app",
		Imports: []module.Module{
			sqlite.NewModule(sqlite.Options{}),
		},
	}
}
