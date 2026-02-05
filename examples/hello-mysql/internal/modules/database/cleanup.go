package database

import (
	"context"
	"database/sql"
)

func CleanupDB(ctx context.Context, db *sql.DB) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if db == nil {
		return nil
	}
	return db.Close()
}
