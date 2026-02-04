package users

import (
	"context"
	"log/slog"

	modkitlogging "github.com/aryeko/modkit/modkit/logging"
)

type Service interface {
	GetUser(ctx context.Context, id int64) (User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (User, error)
	DeleteUser(ctx context.Context, id int64) error
}

type service struct {
	repo   Repository
	logger modkitlogging.Logger
}

func NewService(repo Repository, logger modkitlogging.Logger) Service {
	if logger == nil {
		logger = modkitlogging.Nop()
	}
	logger = logger.With(slog.String("scope", "users"))
	return &service{repo: repo, logger: logger}
}

func (s *service) GetUser(ctx context.Context, id int64) (User, error) {
	s.logger.Debug("get user", slog.Int64("id", id))
	return s.repo.GetUser(ctx, id)
}

func (s *service) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	s.logger.Debug("create user", slog.String("email", input.Email))
	return s.repo.CreateUser(ctx, input)
}

func (s *service) ListUsers(ctx context.Context) ([]User, error) {
	s.logger.Debug("list users")
	return s.repo.ListUsers(ctx)
}

func (s *service) UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (User, error) {
	s.logger.Debug("update user", slog.Int64("id", id))
	return s.repo.UpdateUser(ctx, id, input)
}

func (s *service) DeleteUser(ctx context.Context, id int64) error {
	s.logger.Debug("delete user", slog.Int64("id", id))
	return s.repo.DeleteUser(ctx, id)
}
