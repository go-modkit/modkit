# Error Handling

modkit follows Go's explicit error handling philosophy. There are no exception filters or global error interceptors—you handle errors directly in handlers and middleware.

## Error Handling Philosophy

In modkit:

- Errors are values, not exceptions
- Handlers return errors via HTTP responses
- Middleware can catch and transform errors
- Use RFC 7807 Problem Details for structured API errors

## Handler-Level Errors

Handle errors explicitly in each handler:

```go
func (c *UsersController) Get(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    user, err := c.service.GetByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "user not found", http.StatusNotFound)
            return
        }
        // Log unexpected errors
        log.Printf("error getting user: %v", err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

## Defining Application Errors

Use sentinel errors for known error conditions:

```go
package users

import "errors"

var (
    ErrNotFound       = errors.New("user not found")
    ErrDuplicateEmail = errors.New("email already exists")
    ErrInvalidInput   = errors.New("invalid input")
)
```

Or use custom error types for richer context:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// Check error type
var valErr *ValidationError
if errors.As(err, &valErr) {
    // Handle validation error
}
```

## RFC 7807 Problem Details

For API-friendly error responses, use RFC 7807 Problem Details:

```go
type ProblemDetail struct {
    Type     string `json:"type"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail,omitempty"`
    Instance string `json:"instance,omitempty"`
}

func writeProblem(w http.ResponseWriter, p ProblemDetail) {
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(p.Status)
    json.NewEncoder(w).Encode(p)
}
```

Usage:

```go
func (c *UsersController) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeProblem(w, ProblemDetail{
            Type:   "https://api.example.com/problems/invalid-json",
            Title:  "Invalid JSON",
            Status: http.StatusBadRequest,
            Detail: "The request body could not be parsed as JSON.",
        })
        return
    }
    
    if err := c.service.Create(r.Context(), req); err != nil {
        if errors.Is(err, ErrDuplicateEmail) {
            writeProblem(w, ProblemDetail{
                Type:   "https://api.example.com/problems/duplicate-email",
                Title:  "Email Already Exists",
                Status: http.StatusConflict,
                Detail: fmt.Sprintf("A user with email %q already exists.", req.Email),
            })
            return
        }
        // ...
    }
}
```

## Error Response Helper

Create a helper for consistent error responses:

```go
package httpapi

type ErrorResponse struct {
    Error   string            `json:"error"`
    Code    string            `json:"code,omitempty"`
    Details map[string]string `json:"details,omitempty"`
}

func WriteError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func WriteErrorWithCode(w http.ResponseWriter, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{Error: message, Code: code})
}
```

## Error Middleware

Catch panics and unhandled errors with recovery middleware:

```go
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    logger.Error("panic recovered",
                        "error", err,
                        "path", r.URL.Path,
                        "stack", string(debug.Stack()),
                    )
                    
                    writeProblem(w, ProblemDetail{
                        Type:   "https://api.example.com/problems/internal-error",
                        Title:  "Internal Server Error",
                        Status: http.StatusInternalServerError,
                    })
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

## Error Wrapping Pattern

For errors that need to bubble up with context:

```go
// In repository
func (r *MySQLUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    user, err := r.queries.GetUser(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("query user %d: %w", id, err)
    }
    return &user, nil
}

// In service
func (s *UsersService) GetByID(ctx context.Context, id int) (*User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    return user, nil
}

// In handler - check for specific errors
func (c *UsersController) Get(w http.ResponseWriter, r *http.Request) {
    user, err := c.service.GetByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            // Handle not found
        }
        // Handle other errors
    }
}
```

## Validation Errors

Handle validation with structured error responses:

```go
type ValidationErrors struct {
    Errors []FieldError `json:"errors"`
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (c *UsersController) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        WriteError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    
    // Validate
    var errs []FieldError
    if req.Name == "" {
        errs = append(errs, FieldError{Field: "name", Message: "required"})
    }
    if !isValidEmail(req.Email) {
        errs = append(errs, FieldError{Field: "email", Message: "invalid format"})
    }
    
    if len(errs) > 0 {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(ValidationErrors{Errors: errs})
        return
    }
    
    // Proceed with creation...
}
```

## Kernel Bootstrap Errors

modkit returns typed errors during bootstrap:

| Error | Cause |
|-------|-------|
| `RootModuleNilError` | Bootstrap called with nil |
| `DuplicateModuleNameError` | Two modules share a name |
| `ModuleCycleError` | Import cycle detected |
| `DuplicateProviderTokenError` | Token registered twice |
| `ProviderNotFoundError` | `Get()` for unknown token |
| `TokenNotVisibleError` | Token not exported to requester |
| `ProviderCycleError` | Provider depends on itself |
| `ProviderBuildError` | Provider's Build function failed |
| `ControllerBuildError` | Controller's Build function failed |

Handle these at startup:

```go
app, err := kernel.Bootstrap(&AppModule{})
if err != nil {
    var cycleErr *kernel.ModuleCycleError
    if errors.As(err, &cycleErr) {
        log.Fatalf("module cycle: %v", cycleErr.Cycle)
    }
    log.Fatalf("bootstrap failed: %v", err)
}
```

## Comparison with NestJS

| NestJS | modkit |
|--------|--------|
| `throw new NotFoundException()` | `return nil, ErrNotFound` |
| `@Catch()` exception filter | Recovery middleware |
| `HttpException` hierarchy | Sentinel errors + Problem Details |
| Global exception filter | Global middleware |
| `ValidationPipe` | Explicit validation in handlers |

## Tips

- Return errors, don't panic (except for truly unrecoverable situations)
- Use sentinel errors (`ErrNotFound`) for expected conditions
- Wrap errors with context as they bubble up
- Respond with consistent JSON structure
- Log unexpected errors with stack traces
- Use Problem Details for public APIs
- Handle validation errors with field-level feedback
