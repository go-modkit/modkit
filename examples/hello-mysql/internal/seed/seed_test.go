package seed

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/platform/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSeed_InsertsDefaultUsersIfEmpty(t *testing.T) {
	ctx := context.Background()
	container, dsn := startMySQL(t, ctx)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	if err := waitForMySQL(ctx, dsn); err != nil {
		t.Fatalf("mysql not ready: %v", err)
	}

	db, err := mysql.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := mysql.ApplyMigrations(ctx, db, "../../migrations"); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	if err := Seed(ctx, db); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	count := countUsers(t, db)
	if count == 0 {
		t.Fatalf("expected seeded users, got %d", count)
	}

	if err := Seed(ctx, db); err != nil {
		t.Fatalf("seed idempotent call failed: %v", err)
	}

	countAfter := countUsers(t, db)
	if countAfter != count {
		t.Fatalf("expected idempotent seed, got %d -> %d", count, countAfter)
	}
}

func countUsers(t *testing.T, db *sql.DB) int64 {
	row := db.QueryRow("SELECT COUNT(*) FROM users")
	var count int64
	if err := row.Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	return count
}

func startMySQL(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "app",
		},
		WaitingFor: wait.ForListeningPort("3306/tcp").WithStartupTimeout(2 * time.Minute),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("container start failed: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("container host failed: %v", err)
	}
	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("container port failed: %v", err)
	}

	dsn := "root:password@tcp(" + host + ":" + port.Port() + ")/app?parseTime=true&multiStatements=true"
	return container, dsn
}

func waitForMySQL(ctx context.Context, dsn string) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(30 * time.Second)
	}

	for {
		db, err := mysql.Open(dsn)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			pingErr := db.PingContext(pingCtx)
			cancel()
			_ = db.Close()
			if pingErr == nil {
				return nil
			}
		}

		if time.Now().After(deadline) {
			if err != nil {
				return err
			}
			return context.DeadlineExceeded
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}
