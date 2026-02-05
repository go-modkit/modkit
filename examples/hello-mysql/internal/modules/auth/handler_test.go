package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

func setupAuthHandler() (Config, *Handler, http.Handler) {
	cfg := Config{
		Secret:   "test-secret",
		Issuer:   "test-issuer",
		TTL:      time.Minute,
		Username: "demo",
		Password: "s3cret",
	}

	handler := NewHandler(cfg)
	router := modkithttp.NewRouter()
	handler.RegisterRoutes(modkithttp.AsRouter(router))

	return cfg, handler, router
}

func TestHandler_Login_BadJSON(t *testing.T) {
	_, _, router := setupAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte(`{"username":`)))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandler_Login_WrongCreds(t *testing.T) {
	_, _, router := setupAuthHandler()

	body := []byte(`{"username":"demo","password":"nope"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	cfg, _, router := setupAuthHandler()

	body := []byte(`{"username":"demo","password":"s3cret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Token == "" {
		t.Fatalf("expected token in response")
	}
	if _, err := parseToken(payload.Token, cfg, time.Now()); err != nil {
		t.Fatalf("parse token: %v", err)
	}
}
