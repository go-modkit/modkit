package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/app"
)

func TestBuildHandler_DocsRoute(t *testing.T) {
	h, err := BuildHandler(app.Options{HTTPAddr: ":8080", MySQLDSN: "root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true"})
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
