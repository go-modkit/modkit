package smoke

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/go-modkit/modkit/examples/hello-postgres/internal/app"
	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
	"github.com/go-modkit/modkit/modkit/testkit"
)

func TestSmoke_Postgres_ModuleBootsAndQueries(t *testing.T) {
	requireDocker(t)

	ctx := context.Background()
	container, dsn := startPostgres(t, ctx)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	t.Setenv("POSTGRES_DSN", dsn)
	t.Setenv("POSTGRES_CONNECT_TIMEOUT", "2s")

	h := testkit.New(t, app.NewModule())

	db := testkit.Get[*sql.DB](t, h, sqlmodule.TokenDB)
	dialect := testkit.Get[sqlmodule.Dialect](t, h, sqlmodule.TokenDialect)
	if dialect != sqlmodule.DialectPostgres {
		t.Fatalf("unexpected dialect: %q", dialect)
	}

	qctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var one int
	if err := db.QueryRowContext(qctx, "SELECT 1").Scan(&one); err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if one != 1 {
		t.Fatalf("unexpected result: %d", one)
	}
}

func requireDocker(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker binary not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Skip("docker info timed out")
		}
		t.Skipf("docker unavailable: %v", err)
	}
}

func startPostgres(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "app",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
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
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("container port failed: %v", err)
	}

	dsn := fmt.Sprintf("postgres://postgres:password@%s:%s/app?sslmode=disable", host, port.Port())
	return container, dsn
}
