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

func TestAuthIntegration_LoginAndProtect(t *testing.T) {
	cfg := Config{
		Secret:   "test-secret",
		Issuer:   "test-issuer",
		TTL:      time.Minute,
		Username: "demo",
		Password: "s3cret",
	}

	router := modkithttp.NewRouter()
	root := modkithttp.AsRouter(router)
	handler := NewHandler(cfg)
	handler.RegisterRoutes(root)

	root.Group("/", func(r Router) {
		r.Use(NewJWTMiddleware(cfg))
		r.Handle(http.MethodGet, "/protected", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				t.Fatal("expected user in context")
			}
			if user.Email == "" {
				t.Fatal("expected email in context")
			}
			w.WriteHeader(http.StatusOK)
		}))
	})

	loginBody := []byte(`{"username":"demo","password":"s3cret"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginRec.Code)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("expected login token")
	}

	protectedReq := httptest.NewRequest(http.MethodGet, "/protected", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	protectedRec := httptest.NewRecorder()
	router.ServeHTTP(protectedRec, protectedReq)
	if protectedRec.Code != http.StatusOK {
		t.Fatalf("expected protected 200, got %d", protectedRec.Code)
	}
}

func TestAuthIntegration_Login_InvalidCredentials(t *testing.T) {
	cfg := Config{
		Secret:   "test-secret",
		Issuer:   "test-issuer",
		TTL:      time.Minute,
		Username: "demo",
		Password: "s3cret",
	}

	router := modkithttp.NewRouter()
	root := modkithttp.AsRouter(router)
	handler := NewHandler(cfg)
	handler.RegisterRoutes(root)

	loginBody := []byte(`{"username":"demo","password":"wrong"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected login 401, got %d", loginRec.Code)
	}
}
