package seed

import (
	"context"
	"database/sql"
	"errors"
)

type userSeed struct {
	Name  string
	Email string
}

var defaultUsers = []userSeed{
	{Name: "Ada Lovelace", Email: "ada@example.com"},
	{Name: "Grace Hopper", Email: "grace@example.com"},
	{Name: "Linus Torvalds", Email: "linus@example.com"},
}

func Seed(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("seed db is required")
	}

	var count int
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users")
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, seedUser := range defaultUsers {
		if _, err := tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", seedUser.Name, seedUser.Email); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
