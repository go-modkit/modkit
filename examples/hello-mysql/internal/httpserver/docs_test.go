package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildHandler_DocsRoute(t *testing.T) {
	h, err := BuildHandler()
	if err != nil {
		t.Fatalf("build handler: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs/index.html", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
