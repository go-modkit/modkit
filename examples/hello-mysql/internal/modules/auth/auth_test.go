package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJWTMiddleware_SetsUserContext(t *testing.T) {
	cfg := Config{
		Secret: "test-secret",
		Issuer: "test-issuer",
		TTL:    time.Minute,
	}
	token, err := IssueToken(cfg, User{ID: "demo", Email: "demo@example.com"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	mw := NewJWTMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatal("expected user in context")
		}
		if user.Email != "demo@example.com" {
			t.Fatalf("unexpected user: %+v", user)
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestJWTMiddleware_RejectsWrongIssuer(t *testing.T) {
	cfg := Config{
		Secret: "test-secret",
		Issuer: "test-issuer",
		TTL:    time.Minute,
	}
	otherCfg := Config{
		Secret: "test-secret",
		Issuer: "other-issuer",
		TTL:    time.Minute,
	}

	token, err := IssueToken(otherCfg, User{ID: "demo", Email: "demo@example.com"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	mw := NewJWTMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("expected middleware to reject token with wrong issuer")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Fatalf("WWW-Authenticate = %q", got)
	}
}
