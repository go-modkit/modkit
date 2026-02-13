package smoke

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/go-modkit/modkit/examples/hello-postgres/internal/httpserver"
	"github.com/go-modkit/modkit/modkit/data/sqlmodule"
)

func TestSmoke_Postgres_ModuleBootsAndServes(t *testing.T) {
	requireDocker(t)

	ctx := context.Background()
	container, dsn := startPostgres(t, ctx)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	t.Setenv("POSTGRES_DSN", dsn)
	t.Setenv("POSTGRES_CONNECT_TIMEOUT", "2s")

	boot, handler, err := httpserver.BuildAppHandler()
	if err != nil {
		t.Fatalf("build handler failed: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	if got := bytes.TrimSpace(buf.Bytes()); len(got) == 0 {
		t.Fatalf("expected non-empty body")
	}

	dbAny, err := boot.Get(sqlmodule.TokenDB)
	if err != nil {
		t.Fatalf("resolve db: %v", err)
	}
	db, ok := dbAny.(*sql.DB)
	if !ok {
		t.Fatalf("unexpected db type: %T", dbAny)
	}

	dialectAny, err := boot.Get(sqlmodule.TokenDialect)
	if err != nil {
		t.Fatalf("resolve dialect: %v", err)
	}
	dialect, ok := dialectAny.(sqlmodule.Dialect)
	if !ok {
		t.Fatalf("unexpected dialect type: %T", dialectAny)
	}
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
