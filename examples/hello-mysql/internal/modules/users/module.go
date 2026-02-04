package users

import (
	"database/sql"
	"time"

	"github.com/aryeko/modkit/examples/hello-mysql/internal/modules/database"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/logging"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/sqlc"
	"github.com/aryeko/modkit/modkit/module"
)

const (
	TokenRepository   module.Token = "users.repository"
	TokenService      module.Token = "users.service"
	TokenController   module.Token = "users.controller"
	UsersControllerID              = "UsersController"
)

type Options struct {
	Database module.Module
}

type Module struct {
	opts Options
}

type UsersModule = Module

func NewModule(opts Options) module.Module {
	return &Module{opts: opts}
}

func (m Module) Definition() module.ModuleDef {
	return module.ModuleDef{
		Name:    "users",
		Imports: []module.Module{m.opts.Database},
		Providers: []module.ProviderDef{
			{
				Token: TokenRepository,
				Build: func(r module.Resolver) (any, error) {
					dbAny, err := r.Get(database.TokenDB)
					if err != nil {
						return nil, err
					}
					queries := sqlc.New(dbAny.(*sql.DB))
					return NewMySQLRepo(queries), nil
				},
			},
			{
				Token: TokenService,
				Build: func(r module.Resolver) (any, error) {
					repoAny, err := r.Get(TokenRepository)
					if err != nil {
						return nil, err
					}
					return NewService(repoAny.(Repository), logging.New()), nil
				},
			},
		},
		Controllers: []module.ControllerDef{
			{
				Name: UsersControllerID,
				Build: func(r module.Resolver) (any, error) {
					svcAny, err := r.Get(TokenService)
					if err != nil {
						return nil, err
					}
					return NewController(svcAny.(Service)), nil
				},
			},
		},
		Exports: []module.Token{TokenService},
	}
}

// User represents the API response model.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
