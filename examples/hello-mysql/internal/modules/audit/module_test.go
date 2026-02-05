package audit

import (
	"context"
	"errors"
	"testing"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/users"
	"github.com/go-modkit/modkit/modkit/module"
)

type stubResolver struct {
	values map[module.Token]any
	errors map[module.Token]error
}

func (r stubResolver) Get(token module.Token) (any, error) {
	if err := r.errors[token]; err != nil {
		return nil, err
	}
	if val, ok := r.values[token]; ok {
		return val, nil
	}
	return nil, nil
}

type stubUserService struct{}

func (stubUserService) GetUser(ctx context.Context, id int64) (users.User, error) {
	return users.User{}, nil
}
func (stubUserService) CreateUser(ctx context.Context, input users.CreateUserInput) (users.User, error) {
	return users.User{}, nil
}
func (stubUserService) ListUsers(ctx context.Context) ([]users.User, error) {
	return nil, nil
}
func (stubUserService) UpdateUser(ctx context.Context, id int64, input users.UpdateUserInput) (users.User, error) {
	return users.User{}, nil
}
func (stubUserService) DeleteUser(ctx context.Context, id int64) error {
	return nil
}
func (stubUserService) LongOperation(ctx context.Context) error {
	return nil
}

func TestAuditModule_Definition_WiresUsersImport(t *testing.T) {
	usersMod := &users.Module{}
	mod := NewModule(Options{Users: usersMod})
	def := mod.(*Module).Definition()
	if def.Name != "audit" {
		t.Fatalf("expected name audit, got %q", def.Name)
	}
	if len(def.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(def.Imports))
	}
	if def.Imports[0].Definition().Name != "users" {
		t.Fatalf("expected users import, got %q", def.Imports[0].Definition().Name)
	}
}

func TestAuditModule_ProviderBuildInvokesUsersService(t *testing.T) {
	mod := NewModule(Options{Users: &users.Module{}})
	def := mod.(*Module).Definition()
	provider := def.Providers[0]
	resolver := stubResolver{
		values: map[module.Token]any{
			users.TokenService: stubUserService{},
		},
	}
	res, err := provider.Build(resolver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(Service); !ok {
		t.Fatalf("expected Service, got %T", res)
	}
}

func TestAuditModule_ProviderBuildError(t *testing.T) {
	mod := NewModule(Options{Users: &users.Module{}})
	def := mod.(*Module).Definition()
	provider := def.Providers[0]

	_, err := provider.Build(stubResolver{
		errors: map[module.Token]error{
			users.TokenService: errors.New("users service not found"),
		},
	})
	if err == nil {
		t.Fatal("expected error for missing users service")
	}
	if err.Error() != "users service not found" {
		t.Fatalf("expected 'users service not found' error, got %q", err.Error())
	}
}
