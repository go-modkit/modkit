package users

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/database"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/logging"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/sqlc"
	"github.com/go-modkit/modkit/modkit/module"
)

const (
	TokenRepository   module.Token = "users.repository"
	TokenService      module.Token = "users.service"
	TokenController   module.Token = "users.controller"
	UsersControllerID              = "UsersController"
)

type Options struct {
	Database module.Module
	Auth     module.Module
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
		Imports: []module.Module{m.opts.Database, m.opts.Auth},
		Providers: []module.ProviderDef{
			{
				Token: TokenRepository,
				Build: func(r module.Resolver) (any, error) {
					db, err := module.Get[*sql.DB](r, database.TokenDB)
					if err != nil {
						return nil, err
					}
					queries := sqlc.New(db)
					return NewMySQLRepo(queries), nil
				},
			},
			{
				Token: TokenService,
				Build: func(r module.Resolver) (any, error) {
					repo, err := module.Get[Repository](r, TokenRepository)
					if err != nil {
						return nil, err
					}
					return NewService(repo, logging.New()), nil
				},
			},
		},
		Controllers: []module.ControllerDef{
			{
				Name: UsersControllerID,
				Build: func(r module.Resolver) (any, error) {
					svc, err := module.Get[Service](r, TokenService)
					if err != nil {
						return nil, err
					}
					authMiddleware, err := module.Get[func(http.Handler) http.Handler](r, auth.TokenMiddleware)
					if err != nil {
						return nil, err
					}
					return NewController(svc, authMiddleware), nil
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
