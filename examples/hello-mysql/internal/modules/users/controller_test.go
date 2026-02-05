package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/httpapi"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/modules/auth"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

type stubService struct {
	createFn func(ctx context.Context, input CreateUserInput) (User, error)
	listFn   func(ctx context.Context) ([]User, error)
	updateFn func(ctx context.Context, id int64, input UpdateUserInput) (User, error)
	deleteFn func(ctx context.Context, id int64) error
	getFn    func(ctx context.Context, id int64) (User, error)
}

func allowAll(next http.Handler) http.Handler {
	return next
}
func (s stubService) GetUser(ctx context.Context, id int64) (User, error) {
	if s.getFn == nil {
		return User{}, nil
	}
	return s.getFn(ctx, id)
}

func (s stubService) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	return s.createFn(ctx, input)
}

func (s stubService) ListUsers(ctx context.Context) ([]User, error) {
	return s.listFn(ctx)
}

func (s stubService) UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (User, error) {
	return s.updateFn(ctx, id, input)
}

func (s stubService) DeleteUser(ctx context.Context, id int64) error {
	return s.deleteFn(ctx, id)
}

func TestController_CreateUser(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) {
			if input.Name != "Ada" || input.Email != "ada@example.com" {
				t.Fatalf("unexpected input: %+v", input)
			}
			return User{ID: 10, Name: input.Name, Email: input.Email}, nil
		},
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc, allowAll)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	body := []byte(`{"name":"Ada","email":"ada@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	var user User
	if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if user.ID != 10 {
		t.Fatalf("expected id 10, got %d", user.ID)
	}
}

func TestController_CreateUser_Conflict(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) {
			return User{}, ErrConflict
		},
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc, allowAll)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	body := []byte(`{"name":"Ada","email":"ada@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rec.Code)
	}
	var problem httpapi.Problem
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if problem.Status != http.StatusConflict {
		t.Fatalf("expected problem status 409, got %d", problem.Status)
	}
}

func TestController_GetUser_NotFound(t *testing.T) {
	svc := stubService{
		getFn: func(ctx context.Context, id int64) (User, error) {
			return User{}, ErrNotFound
		},
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc, allowAll)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	req := httptest.NewRequest(http.MethodGet, "/users/99", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	var problem httpapi.Problem
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if problem.Status != http.StatusNotFound {
		t.Fatalf("expected problem status 404, got %d", problem.Status)
	}
}

func TestController_ListUsers(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn: func(ctx context.Context) ([]User, error) {
			return []User{{ID: 1, Name: "Ada", Email: "ada@example.com"}}, nil
		},
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc, allowAll)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var users []User
	if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(users) != 1 || users[0].ID != 1 {
		t.Fatalf("unexpected users: %+v", users)
	}
}

func TestController_UpdateUser(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) {
			if id != 4 {
				t.Fatalf("expected id 4, got %d", id)
			}
			return User{ID: id, Name: input.Name, Email: input.Email}, nil
		},
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc, allowAll)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	body := []byte(`{"name":"Bea","email":"bea@example.com"}`)
	req := httptest.NewRequest(http.MethodPut, "/users/4", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var user User
	if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if user.Email != "bea@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestController_DeleteUser(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error {
			if id != 3 {
				t.Fatalf("expected id 3, got %d", id)
			}
			return nil
		},
	}

	controller := NewController(svc, allowAll)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	req := httptest.NewRequest(http.MethodDelete, "/users/3", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestController_CreateUser_RequiresAuth(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	mw := auth.NewJWTMiddleware(auth.Config{
		Secret: "test-secret",
		Issuer: "test-issuer",
		TTL:    time.Minute,
	})

	controller := NewController(svc, mw)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	body := []byte(`{"name":"Ada","email":"ada@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
