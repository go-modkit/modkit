package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestNewRouter_AllowsRoute(t *testing.T) {
	router := NewRouter()
	router.Method(http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestRouterGroup_RegistersGroupedRoutes(t *testing.T) {
	router := chi.NewRouter()
	r := AsRouter(router)

	called := false
	r.Group("/api", func(sub Router) {
		sub.Handle(http.MethodGet, "/users", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			called = true
		}))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected grouped handler to be called")
	}
}

func TestRouterUse_AttachesMiddleware(t *testing.T) {
	router := chi.NewRouter()
	r := AsRouter(router)

	middlewareCalled := false
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, req)
		})
	})

	r.Handle(http.MethodGet, "/test", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !middlewareCalled {
		t.Fatal("expected middleware to be called")
	}
}

func TestRouterGroup_MiddlewareScopedToGroup(t *testing.T) {
	router := chi.NewRouter()
	r := AsRouter(router)

	groupMiddlewareCalled := false
	groupHandlerCalled := false

	r.Handle(http.MethodGet, "/public", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

	r.Group("/protected", func(sub Router) {
		sub.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				groupMiddlewareCalled = true
				next.ServeHTTP(w, req)
			})
		})
		sub.Handle(http.MethodGet, "/resource", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			groupHandlerCalled = true
		}))
	})

	reqPublic := httptest.NewRequest(http.MethodGet, "/public", http.NoBody)
	router.ServeHTTP(httptest.NewRecorder(), reqPublic)

	if groupMiddlewareCalled {
		t.Fatal("group middleware should not affect routes outside group")
	}

	reqProtected := httptest.NewRequest(http.MethodGet, "/protected/resource", http.NoBody)
	router.ServeHTTP(httptest.NewRecorder(), reqProtected)

	if !groupMiddlewareCalled || !groupHandlerCalled {
		t.Fatal("expected group middleware and handler to be called")
	}
}
