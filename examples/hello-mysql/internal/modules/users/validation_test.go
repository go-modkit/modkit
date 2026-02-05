package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/validation"
	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

func TestController_CreateUser_Validation(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	body := []byte(`{"name":"","email":""}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	var problem validation.ProblemDetails
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if problem.Status != http.StatusBadRequest {
		t.Fatalf("expected problem status 400, got %d", problem.Status)
	}
	if len(problem.InvalidParams) != 2 {
		t.Fatalf("expected 2 invalid params, got %d", len(problem.InvalidParams))
	}
}

func TestController_UpdateUser_Validation(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	body := []byte(`{"name":"","email":""}`)
	req := httptest.NewRequest(http.MethodPut, "/users/5", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	var problem validation.ProblemDetails
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if problem.Status != http.StatusBadRequest {
		t.Fatalf("expected problem status 400, got %d", problem.Status)
	}
	if len(problem.InvalidParams) != 2 {
		t.Fatalf("expected 2 invalid params, got %d", len(problem.InvalidParams))
	}
}

func TestController_CreateUser_InvalidJSONBody(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	var problem validation.ProblemDetails
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(problem.InvalidParams) != 1 || problem.InvalidParams[0].Name != "body" {
		t.Fatalf("expected invalidParams to include body, got %+v", problem.InvalidParams)
	}
}

func TestController_UpdateUser_InvalidID(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	req := httptest.NewRequest(http.MethodPut, "/users/not-a-number", bytes.NewReader([]byte(`{"name":"Ada","email":"ada@example.com"}`)))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	var problem validation.ProblemDetails
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(problem.InvalidParams) != 1 || problem.InvalidParams[0].Name != "id" {
		t.Fatalf("expected invalidParams to include id, got %+v", problem.InvalidParams)
	}
}

func TestController_UpdateUser_InvalidJSONBody(t *testing.T) {
	svc := stubService{
		createFn: func(ctx context.Context, input CreateUserInput) (User, error) { return User{}, nil },
		listFn:   func(ctx context.Context) ([]User, error) { return nil, nil },
		updateFn: func(ctx context.Context, id int64, input UpdateUserInput) (User, error) { return User{}, nil },
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}

	controller := NewController(svc)
	router := modkithttp.NewRouter()
	controller.RegisterRoutes(modkithttp.AsRouter(router))

	req := httptest.NewRequest(http.MethodPut, "/users/5", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	var problem validation.ProblemDetails
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(problem.InvalidParams) != 1 || problem.InvalidParams[0].Name != "body" {
		t.Fatalf("expected invalidParams to include body, got %+v", problem.InvalidParams)
	}
}
