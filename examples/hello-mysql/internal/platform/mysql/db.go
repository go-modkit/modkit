package mysql

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("mysql dsn is required")
	}
	return sql.Open("mysql", dsn)
}

func ApplyMigrations(ctx context.Context, db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(files)

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(content) == 0 {
			continue
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return err
		}
	}

	return nil
}
