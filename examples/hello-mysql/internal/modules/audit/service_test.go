package audit

import (
	"context"
	"testing"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/users"
)

type stubUsersService struct {
	user users.User
}

func (s stubUsersService) GetUser(ctx context.Context, id int64) (users.User, error) {
	return s.user, nil
}

func (s stubUsersService) CreateUser(ctx context.Context, input users.CreateUserInput) (users.User, error) {
	return users.User{}, nil
}

func (s stubUsersService) ListUsers(ctx context.Context) ([]users.User, error) {
	return nil, nil
}

func (s stubUsersService) UpdateUser(ctx context.Context, id int64, input users.UpdateUserInput) (users.User, error) {
	return users.User{}, nil
}

func (s stubUsersService) DeleteUser(ctx context.Context, id int64) error {
	return nil
}

func (s stubUsersService) LongOperation(ctx context.Context) error {
	return nil
}

func TestAuditService_FormatsEntry(t *testing.T) {
	svc := NewService(stubUsersService{user: users.User{ID: 3, Name: "Jo", Email: "jo@example.com"}})

	entry, err := svc.AuditUserLookup(context.Background(), 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry == "" {
		t.Fatalf("expected non-empty entry")
	}
}
