package sqlite

import (
	"context"
	"database/sql"
)

// CleanupDB closes a DB handle if present.
func CleanupDB(ctx context.Context, db *sql.DB) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if db == nil {
		return nil
	}
	return db.Close()
}
