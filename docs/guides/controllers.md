# Controllers

Controllers handle HTTP requests and define your application's API surface. They're the bridge between the HTTP layer and your business logic (providers).

## What is a Controller?

A controller in modkit:

- Receives HTTP requests via registered routes
- Delegates work to providers (services, repositories)
- Returns HTTP responses
- Implements the `RouteRegistrar` interface

## RouteRegistrar Interface

Every controller must implement this interface:

```go
type RouteRegistrar interface {
    RegisterRoutes(router Router)
}
```

The `Router` interface provides:

```go
type Router interface {
    Handle(method, pattern string, handler http.Handler)
    Group(pattern string, fn func(Router))
    Use(middleware ...func(http.Handler) http.Handler)
}
```

## Defining Controllers

### Basic Controller

```go
type UsersController struct {
    service UsersService
}

func NewUsersController(service UsersService) *UsersController {
    return &UsersController{service: service}
}

func (c *UsersController) RegisterRoutes(r mkhttp.Router) {
    r.Handle(http.MethodGet, "/users", http.HandlerFunc(c.List))
    r.Handle(http.MethodPost, "/users", http.HandlerFunc(c.Create))
    r.Handle(http.MethodGet, "/users/{id}", http.HandlerFunc(c.Get))
    r.Handle(http.MethodPut, "/users/{id}", http.HandlerFunc(c.Update))
    r.Handle(http.MethodDelete, "/users/{id}", http.HandlerFunc(c.Delete))
}

func (c *UsersController) List(w http.ResponseWriter, r *http.Request) {
    users, err := c.service.List(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(users)
}
```

### Registering in a Module

```go
func (m *UsersModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name:    "users",
        Imports: []module.Module{m.db},
        Providers: []module.ProviderDef{{
            Token: TokenUsersService,
            Build: buildUsersService,
        }},
        Controllers: []module.ControllerDef{{
            Name: "UsersController",
            Build: func(r module.Resolver) (any, error) {
                svc, err := module.Get[UsersService](r, TokenUsersService)
                if err != nil {
                    return nil, err
                }
                return NewUsersController(svc), nil
            },
        }},
    }
}
```

## ControllerDef Fields

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Unique identifier for the controller within the module |
| `Build` | `func(Resolver) (any, error)` | Factory function that creates the controller |

## Route Patterns

modkit uses chi router patterns under the hood:

```go
// Static routes
r.Handle(http.MethodGet, "/health", handler)

// URL parameters
r.Handle(http.MethodGet, "/users/{id}", handler)

// Nested paths
r.Handle(http.MethodGet, "/users/{userID}/posts/{postID}", handler)
```

Extract parameters using chi:

```go
import "github.com/go-chi/chi/v5"

func (c *UsersController) Get(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    // ...
}
```

## Route Grouping

Use `Group` for common prefixes or middleware:

```go
func (c *UsersController) RegisterRoutes(r mkhttp.Router) {
    r.Group("/users", func(r mkhttp.Router) {
        r.Handle(http.MethodGet, "/", http.HandlerFunc(c.List))
        r.Handle(http.MethodPost, "/", http.HandlerFunc(c.Create))
        r.Handle(http.MethodGet, "/{id}", http.HandlerFunc(c.Get))
    })
}
```

## Controller-Level Middleware

Apply middleware to all routes in a controller:

```go
func (c *UsersController) RegisterRoutes(r mkhttp.Router) {
    r.Group("/users", func(r mkhttp.Router) {
        // Apply auth middleware to all /users routes
        r.Use(authMiddleware)
        
        r.Handle(http.MethodGet, "/", http.HandlerFunc(c.List))
        r.Handle(http.MethodPost, "/", http.HandlerFunc(c.Create))
    })
}
```

## Handler Patterns

### JSON Response

```go
func (c *UsersController) Get(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    user, err := c.service.GetByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "user not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

### JSON Request Body

```go
func (c *UsersController) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    
    user, err := c.service.Create(r.Context(), req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

### Error Responses with Problem Details

For RFC 7807 compliant errors:

```go
func (c *UsersController) Create(w http.ResponseWriter, r *http.Request) {
    // ...
    if err := c.service.Create(r.Context(), user); err != nil {
        if errors.Is(err, ErrDuplicateEmail) {
            w.Header().Set("Content-Type", "application/problem+json")
            w.WriteHeader(http.StatusConflict)
            json.NewEncoder(w).Encode(map[string]any{
                "type":   "https://example.com/problems/duplicate-email",
                "title":  "Email Already Exists",
                "status": 409,
                "detail": "A user with this email address already exists.",
            })
            return
        }
        // ...
    }
}
```

## Testing Controllers

Test handlers directly with `httptest`:

```go
func TestUsersController_List(t *testing.T) {
    // Mock service
    svc := &mockUsersService{
        listFn: func(ctx context.Context) ([]User, error) {
            return []User{{ID: 1, Name: "Ada"}}, nil
        },
    }
    controller := NewUsersController(svc)

    // Create request
    req := httptest.NewRequest(http.MethodGet, "/users", nil)
    rec := httptest.NewRecorder()

    // Call handler
    controller.List(rec, req)

    // Assert
    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
    
    var users []User
    json.NewDecoder(rec.Body).Decode(&users)
    if len(users) != 1 {
        t.Fatalf("expected 1 user, got %d", len(users))
    }
}
```

## Complete Example

```go
package users

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    mkhttp "github.com/go-modkit/modkit/modkit/http"
)

type Controller struct {
    service Service
}

func NewController(service Service) *Controller {
    return &Controller{service: service}
}

func (c *Controller) RegisterRoutes(r mkhttp.Router) {
    r.Group("/users", func(r mkhttp.Router) {
        r.Handle(http.MethodGet, "/", http.HandlerFunc(c.list))
        r.Handle(http.MethodPost, "/", http.HandlerFunc(c.create))
        r.Handle(http.MethodGet, "/{id}", http.HandlerFunc(c.get))
        r.Handle(http.MethodPut, "/{id}", http.HandlerFunc(c.update))
        r.Handle(http.MethodDelete, "/{id}", http.HandlerFunc(c.delete))
    })
}

func (c *Controller) list(w http.ResponseWriter, r *http.Request) {
    users, err := c.service.List(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func (c *Controller) get(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    user, err := c.service.GetByID(r.Context(), id)
    if err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// ... create, update, delete handlers
```

## Tips

- Keep controllers thin—business logic belongs in services
- Use consistent error response formats across controllers
- Group related routes with common prefixes
- Apply authentication/authorization via middleware, not in handlers
- Return appropriate HTTP status codes
- Set `Content-Type` headers explicitly
