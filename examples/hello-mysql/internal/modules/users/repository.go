package users

import "context"

type Repository interface {
	GetUser(ctx context.Context, id int64) (User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (User, error)
	DeleteUser(ctx context.Context, id int64) error
}
