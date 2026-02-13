package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

// CleanupDB closes a DB handle if present.
func CleanupDB(ctx context.Context, db *sql.DB) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cleanup db: %w", err)
	}
	if db == nil {
		return nil
	}
	if err := db.Close(); err != nil {
		return fmt.Errorf("cleanup db: %w", err)
	}
	return nil
}
