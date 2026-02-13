package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

type errorResponseWriter struct {
	header http.Header
	status int
}

func (w *errorResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w *errorResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *errorResponseWriter) Write(_ []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return 0, errors.New("write failed")
}

func TestHealthEncodeErrorReturnsServerError(t *testing.T) {
	w := &errorResponseWriter{}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	controller := &HealthController{}
	controller.health(w, req)

	if w.status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.status)
	}
}

func TestBuildHandlerServesHealth(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "test")
	t.Setenv("POSTGRES_CONNECT_TIMEOUT", "0")

	handler, err := BuildHandler()
	if err != nil {
		t.Fatalf("build handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", payload["status"])
	}
}

func TestBuildAppHandlerReturnsRegisterError(t *testing.T) {
	oldRegister := registerRoutes
	registerRoutes = func(_ modkithttp.Router, _ map[string]any) error {
		return errors.New("register failed")
	}
	t.Cleanup(func() {
		registerRoutes = oldRegister
	})

	_, _, err := BuildAppHandler()
	if err == nil {
		t.Fatal("expected error")
	}
}
