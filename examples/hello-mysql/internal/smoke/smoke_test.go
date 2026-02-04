package smoke

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aryeko/modkit/examples/hello-mysql/internal/httpserver"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/modules/app"
	"github.com/aryeko/modkit/examples/hello-mysql/internal/platform/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSmoke_HealthAndUsers(t *testing.T) {
	ctx := context.Background()
	container, dsn := startMySQL(t, ctx)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	if err := migrate(ctx, dsn); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	if err := seedUser(ctx, dsn); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	handler, err := httpserver.BuildHandler(app.Options{HTTPAddr: ":8080", MySQLDSN: dsn})
	if err != nil {
		t.Fatalf("build handler failed: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/users/1")
	if err != nil {
		t.Fatalf("users request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if body.ID != 1 || body.Name == "" || body.Email == "" {
		t.Fatalf("unexpected body: %+v", body)
	}
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

	dsn := fmt.Sprintf("root:password@tcp(%s:%s)/app?parseTime=true&multiStatements=true", host, port.Port())
	return container, dsn
}

func migrate(ctx context.Context, dsn string) error {
	if err := waitForMySQL(ctx, dsn); err != nil {
		return err
	}

	db, err := mysql.Open(dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	return mysql.ApplyMigrations(ctx, db, "../../migrations")
}

func seedUser(ctx context.Context, dsn string) error {
	if err := waitForMySQL(ctx, dsn); err != nil {
		return err
	}

	db, err := mysql.Open(dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "INSERT INTO users (id, name, email) VALUES (1, 'Ada', 'ada@example.com')")
	return err
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

var _ *sql.DB
