package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aryeko/modkit/examples/hello-mysql/internal/sqlc"
	"github.com/go-sql-driver/mysql"
)

type mysqlRepo struct {
	queries *sqlc.Queries
}

func NewMySQLRepo(queries *sqlc.Queries) Repository {
	return &mysqlRepo{queries: queries}
}

func (r *mysqlRepo) GetUser(ctx context.Context, id int64) (User, error) {
	row, err := r.queries.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return User{ID: row.ID, Name: row.Name, Email: row.Email, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func (r *mysqlRepo) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	result, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{Name: input.Name, Email: input.Email})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return User{}, ErrConflict
		}
		return User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return r.GetUser(ctx, id)
}

func (r *mysqlRepo) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, User{ID: row.ID, Name: row.Name, Email: row.Email, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return users, nil
}

func (r *mysqlRepo) UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (User, error) {
	result, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{ID: id, Name: input.Name, Email: input.Email})
	if err != nil {
		return User{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return User{}, err
	}
	if affected == 0 {
		return User{}, ErrNotFound
	}
	return r.GetUser(ctx, id)
}

func (r *mysqlRepo) DeleteUser(ctx context.Context, id int64) error {
	result, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
