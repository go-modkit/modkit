package users

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/database"
	"github.com/go-modkit/modkit/modkit/module"
)

func TestUsersModule_Definition_WiresAuth(t *testing.T) {
	mod := NewModule(Options{Database: &database.Module{}, Auth: auth.NewModule(auth.Options{})})
	def := mod.(*Module).Definition()

	if def.Name != "users" {
		t.Fatalf("name = %q", def.Name)
	}
	if len(def.Imports) != 2 {
		t.Fatalf("imports = %d", len(def.Imports))
	}
}

type stubResolver struct {
	values map[module.Token]any
	errors map[module.Token]error
}

func (r stubResolver) Get(token module.Token) (any, error) {
	if err := r.errors[token]; err != nil {
		return nil, err
	}
	if value, ok := r.values[token]; ok {
		return value, nil
	}
	return nil, nil
}

type serviceStub struct{}

func (serviceStub) GetUser(ctx context.Context, id int64) (User, error) {
	return User{}, nil
}

func (serviceStub) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	return User{}, nil
}

func (serviceStub) ListUsers(ctx context.Context) ([]User, error) {
	return nil, nil
}

func (serviceStub) UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (User, error) {
	return User{}, nil
}

func (serviceStub) DeleteUser(ctx context.Context, id int64) error {
	return nil
}

func (serviceStub) LongOperation(ctx context.Context) error {
	return nil
}

func TestUsersModule_ControllerBuildErrors(t *testing.T) {
	mod := NewModule(Options{Database: &database.Module{}, Auth: auth.NewModule(auth.Options{})})
	def := mod.(*Module).Definition()
	controller := def.Controllers[0]

	_, err := controller.Build(stubResolver{
		errors: map[module.Token]error{
			TokenService: errors.New("missing service"),
		},
	})
	if err == nil {
		t.Fatal("expected error for missing service")
	}

	_, err = controller.Build(stubResolver{
		values: map[module.Token]any{
			TokenService: serviceStub{},
		},
		errors: map[module.Token]error{
			auth.TokenMiddleware: errors.New("missing middleware"),
		},
	})
	if err == nil {
		t.Fatal("expected error for missing middleware")
	}

	_, err = controller.Build(stubResolver{
		values: map[module.Token]any{
			TokenService:         serviceStub{},
			auth.TokenMiddleware: func(next http.Handler) http.Handler { return next },
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
