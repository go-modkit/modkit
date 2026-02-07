package httpserver

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

	h, err := BuildHandler()
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
	boot, handler, err := BuildAppHandler()
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

	boot, handler, err := BuildAppHandler()
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
	t.Setenv("CORS_ALLOWED_METHODS", "GET")

	h, err := BuildHandler()
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

func TestBuildHandler_DocsRedirect(t *testing.T) {
	h, err := BuildHandler()
	if err != nil {
		t.Fatalf("build handler: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected status 301, got %d", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/docs/index.html" {
		t.Fatalf("expected redirect to /docs/index.html, got %q", got)
	}
}
