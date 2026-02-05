package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareRecoverer_HandlesPanic(t *testing.T) {
	router := NewRouter()
	router.Method(http.MethodGet, "/boom", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/boom", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestMiddlewareStopsChainWhenHandlingResponse(t *testing.T) {
	router := NewRouter()
	handled := false

	router.Use(func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})
	router.Method(http.MethodGet, "/teapot", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/teapot", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
	if handled {
		t.Fatalf("expected handler not to run")
	}
}
