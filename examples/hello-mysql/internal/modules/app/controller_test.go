package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

func TestNewController(t *testing.T) {
	ctrl := NewController()
	if ctrl == nil {
		t.Fatal("expected non-nil controller")
	}
}

func TestController_RegisterRoutes(t *testing.T) {
	ctrl := NewController()
	router := modkithttp.NewRouter()
	ctrl.RegisterRoutes(modkithttp.AsRouter(router))

	// Verify route is registered by making a request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestController_HandleHealth(t *testing.T) {
	ctrl := NewController()
	router := modkithttp.NewRouter()
	ctrl.RegisterRoutes(modkithttp.AsRouter(router))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	status, ok := response["status"]
	if !ok {
		t.Fatal("expected 'status' field in response")
	}
	if status != "ok" {
		t.Errorf("expected status 'ok', got %q", status)
	}
}
