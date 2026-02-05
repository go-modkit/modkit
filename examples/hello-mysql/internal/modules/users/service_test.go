package users

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubRepo struct {
	createInput CreateUserInput
	createUser  User
	listUsers   []User
	updateID    int64
	updateInput UpdateUserInput
	updateUser  User
	deleteID    int64
	createErr   error
	listErr     error
	updateErr   error
	deleteErr   error
}

func (s *stubRepo) GetUser(ctx context.Context, id int64) (User, error) {
	return User{}, nil
}

func (s *stubRepo) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	s.createInput = input
	return s.createUser, s.createErr
}

func (s *stubRepo) ListUsers(ctx context.Context) ([]User, error) {
	return s.listUsers, s.listErr
}

func (s *stubRepo) UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (User, error) {
	s.updateID = id
	s.updateInput = input
	return s.updateUser, s.updateErr
}

func (s *stubRepo) DeleteUser(ctx context.Context, id int64) error {
	s.deleteID = id
	return s.deleteErr
}

func TestService_CreateUser(t *testing.T) {
	repo := &stubRepo{createUser: User{ID: 7, Name: "Ada", Email: "ada@example.com"}}
	svc := NewService(repo, nil)

	user, err := svc.CreateUser(context.Background(), CreateUserInput{Name: "Ada", Email: "ada@example.com"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.createInput.Name != "Ada" || repo.createInput.Email != "ada@example.com" {
		t.Fatalf("expected repo called with input, got %+v", repo.createInput)
	}
	if user.ID != 7 {
		t.Fatalf("expected user id 7, got %d", user.ID)
	}
}

func TestService_ListUsers(t *testing.T) {
	repo := &stubRepo{listUsers: []User{{ID: 1, Name: "Ada", Email: "ada@example.com"}}}
	svc := NewService(repo, nil)

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 || users[0].ID != 1 {
		t.Fatalf("expected one user, got %+v", users)
	}
}

func TestService_UpdateUser(t *testing.T) {
	repo := &stubRepo{updateUser: User{ID: 2, Name: "Bea", Email: "bea@example.com"}}
	svc := NewService(repo, nil)

	updated, err := svc.UpdateUser(context.Background(), 2, UpdateUserInput{Name: "Bea", Email: "bea@example.com"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updateID != 2 {
		t.Fatalf("expected update id 2, got %d", repo.updateID)
	}
	if updated.Email != "bea@example.com" {
		t.Fatalf("expected updated email, got %s", updated.Email)
	}
}

func TestService_DeleteUser(t *testing.T) {
	repo := &stubRepo{}
	svc := NewService(repo, nil)

	if err := svc.DeleteUser(context.Background(), 9); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deleteID != 9 {
		t.Fatalf("expected delete id 9, got %d", repo.deleteID)
	}
}

func TestService_LongOperation_RespectsContextCancel(t *testing.T) {
	svc := NewService(&stubRepo{}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.LongOperation(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestService_LongOperation_Completes(t *testing.T) {
	svc := NewService(&stubRepo{}, nil).(*service)
	origDelay := svc.longOperationDelay
	svc.longOperationDelay = 2 * time.Millisecond
	t.Cleanup(func() { svc.longOperationDelay = origDelay })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := svc.LongOperation(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
