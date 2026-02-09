package smoke

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-modkit/modkit/examples/hello-sqlite/internal/app"
	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
	"github.com/go-modkit/modkit/modkit/testkit"
)

func TestSmoke_SQLite_FileBacked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	t.Setenv("SQLITE_PATH", path)
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "2s")

	h := testkit.New(t, app.NewModule())
	db := testkit.Get[*sql.DB](t, h, sqlmodule.TokenDB)
	dialect := testkit.Get[sqlmodule.Dialect](t, h, sqlmodule.TokenDialect)
	if dialect != sqlmodule.DialectSQLite {
		t.Fatalf("unexpected dialect: %q", dialect)
	}

	roundTripSQLite(t, db)
}

func TestSmoke_SQLite_InMemory(t *testing.T) {
	t.Setenv("SQLITE_PATH", "file:memdb1?mode=memory&cache=shared")
	t.Setenv("SQLITE_CONNECT_TIMEOUT", "2s")

	h := testkit.New(t, app.NewModule())
	db := testkit.Get[*sql.DB](t, h, sqlmodule.TokenDB)

	roundTripSQLite(t, db)
}

func roundTripSQLite(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (1, 'Ada')`); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var name string
	if err := db.QueryRowContext(ctx, `SELECT name FROM users WHERE id = 1`).Scan(&name); err != nil {
		t.Fatalf("select: %v", err)
	}
	if name != "Ada" {
		t.Fatalf("unexpected name: %q", name)
	}
}
