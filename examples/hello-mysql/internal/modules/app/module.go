package app

import (
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/audit"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/database"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/users"
	"github.com/go-modkit/modkit/modkit/module"
)

const HealthControllerID = "HealthController"

type Options struct {
	HTTPAddr string
	MySQLDSN string
	Auth     auth.Config
}

type Module struct {
	opts Options
}

type AppModule = Module

func NewModule(opts Options) module.Module {
	return &Module{opts: opts}
}

func (m *Module) Definition() module.ModuleDef {
	dbModule := database.NewModule(database.Options{DSN: m.opts.MySQLDSN})
	authModule := auth.NewModule(auth.Options{Config: m.opts.Auth})
	usersModule := users.NewModule(users.Options{Database: dbModule, Auth: authModule})
	auditModule := audit.NewModule(audit.Options{Users: usersModule})

	return module.ModuleDef{
		Name:    "app",
		Imports: []module.Module{dbModule, authModule, usersModule, auditModule},
		Controllers: []module.ControllerDef{
			{
				Name: HealthControllerID,
				Build: func(r module.Resolver) (any, error) {
					return NewController(), nil
				},
			},
		},
	}
}
