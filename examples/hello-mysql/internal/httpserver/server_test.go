package httpserver

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/app"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

func TestBuildHandler_LogsRequest(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = origStdout
		_ = r.Close()
	}()

	h, err := BuildHandler(testAppOptions())
	if err != nil {
		_ = w.Close()
		t.Fatalf("build handler: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		_ = w.Close()
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	_ = w.Close()
	output, _ := io.ReadAll(r)
	if !bytes.Contains(output, []byte("http request")) {
		t.Fatalf("expected log output, got %s", string(output))
	}
}

func TestBuildAppHandler_ReturnsAppAndHandler(t *testing.T) {
	boot, handler, err := BuildAppHandler(testAppOptions())
	if err != nil {
		t.Fatalf("build app handler: %v", err)
	}
	if boot == nil {
		t.Fatal("expected app, got nil")
	}
	if len(boot.Controllers) == 0 {
		t.Fatal("expected controllers to be registered")
	}
	if handler == nil {
		t.Fatal("expected handler, got nil")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestBuildAppHandler_ReturnsBootOnRouteError(t *testing.T) {
	origRegister := registerRoutes
	registerRoutes = func(_ modkithttp.Router, _ map[string]any) error {
		return errors.New("routes failed")
	}
	defer func() { registerRoutes = origRegister }()

	boot, handler, err := BuildAppHandler(testAppOptions())
	if err == nil {
		t.Fatal("expected error")
	}
	if boot == nil {
		t.Fatal("expected boot to be returned on error")
	}
	if handler != nil {
		t.Fatal("expected nil handler on error")
	}
}

func TestBuildHandler_UsesMiddlewareProviders(t *testing.T) {
	opts := testAppOptions()
	opts.CORSAllowedMethods = []string{"GET"}

	h, err := BuildHandler(opts)
	if err != nil {
		t.Fatalf("build handler: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected Access-Control-Allow-Origin header, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatalf("expected Access-Control-Allow-Methods header, got empty")
	}
}

func testAppOptions() app.Options {
	return app.Options{
		HTTPAddr: ":8080",
		MySQLDSN: "root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true",
		Auth: auth.Config{
			Secret:   "dev-secret-change-me",
			Issuer:   "hello-mysql",
			TTL:      time.Hour,
			Username: "demo",
			Password: "demo",
		},
		CORSAllowedOrigins: []string{"http://localhost:3000"},
		CORSAllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		RateLimitPerSecond: 5,
		RateLimitBurst:     10,
	}
}
