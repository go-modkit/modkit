# Validation

modkit doesn't include a validation framework—you use explicit Go code to validate input. This guide covers common validation patterns.

## Philosophy

In NestJS, validation pipes automatically validate request bodies using decorators. In modkit (and idiomatic Go), you:

1. Decode the request body
2. Validate fields explicitly
3. Return structured errors

This keeps validation visible and testable.

## Basic Validation

### Decode and Validate

```go
func (c *UsersController) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate
    if err := req.Validate(); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(err)
        return
    }

    // Proceed with creation...
}
```

### Request Type with Validation

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (r CreateUserRequest) Validate() *ValidationErrors {
    var errs []FieldError

    if strings.TrimSpace(r.Name) == "" {
        errs = append(errs, FieldError{Field: "name", Message: "required"})
    } else if len(r.Name) > 100 {
        errs = append(errs, FieldError{Field: "name", Message: "must be 100 characters or less"})
    }

    if strings.TrimSpace(r.Email) == "" {
        errs = append(errs, FieldError{Field: "email", Message: "required"})
    } else if !isValidEmail(r.Email) {
        errs = append(errs, FieldError{Field: "email", Message: "invalid format"})
    }

    if len(errs) > 0 {
        return &ValidationErrors{Errors: errs}
    }
    return nil
}

func isValidEmail(email string) bool {
    // Simple check; use a regex or library for production
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}
```

## Validation Error Types

### Field-Level Errors

```go
type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type ValidationErrors struct {
    Errors []FieldError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
    return fmt.Sprintf("validation failed: %d errors", len(e.Errors))
}
```

Response example:

```json
{
  "errors": [
    {"field": "name", "message": "required"},
    {"field": "email", "message": "invalid format"}
  ]
}
```

### RFC 7807 Problem Details

For API consistency, use Problem Details:

```go
func writeValidationProblem(w http.ResponseWriter, errs []FieldError) {
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(http.StatusUnprocessableEntity)
    json.NewEncoder(w).Encode(map[string]any{
        "type":   "https://api.example.com/problems/validation-error",
        "title":  "Validation Failed",
        "status": 422,
        "errors": errs,
    })
}
```

## Using Validation Libraries

For complex validation, consider these Go libraries:

### go-playground/validator

```go
import "github.com/go-playground/validator/v10"

var validate = validator.New()

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,max=100"`
    Email string `json:"email" validate:"required,email"`
}

func (c *UsersController) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    if err := validate.Struct(req); err != nil {
        errs := translateValidationErrors(err.(validator.ValidationErrors))
        writeValidationProblem(w, errs)
        return
    }

    // Proceed...
}

func translateValidationErrors(ve validator.ValidationErrors) []FieldError {
    var errs []FieldError
    for _, fe := range ve {
        errs = append(errs, FieldError{
            Field:   strings.ToLower(fe.Field()),
            Message: translateTag(fe.Tag()),
        })
    }
    return errs
}

func translateTag(tag string) string {
    switch tag {
    case "required":
        return "required"
    case "email":
        return "invalid email format"
    case "max":
        return "exceeds maximum length"
    default:
        return "invalid"
    }
}
```

### ozzo-validation

```go
import validation "github.com/go-ozzo/ozzo-validation/v4"
import "github.com/go-ozzo/ozzo-validation/v4/is"

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (r CreateUserRequest) Validate() error {
    return validation.ValidateStruct(&r,
        validation.Field(&r.Name, validation.Required, validation.Length(1, 100)),
        validation.Field(&r.Email, validation.Required, is.Email),
    )
}
```

## URL Parameter Validation

Validate path and query parameters:

```go
func (c *UsersController) Get(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        http.Error(w, "invalid user ID", http.StatusBadRequest)
        return
    }

    user, err := c.service.GetByID(r.Context(), id)
    // ...
}
```

### Query Parameter Validation

```go
func (c *UsersController) List(w http.ResponseWriter, r *http.Request) {
    // Parse with defaults
    limit := 20
    if l := r.URL.Query().Get("limit"); l != "" {
        parsed, err := strconv.Atoi(l)
        if err != nil || parsed < 1 || parsed > 100 {
            http.Error(w, "limit must be 1-100", http.StatusBadRequest)
            return
        }
        limit = parsed
    }

    users, err := c.service.List(r.Context(), limit)
    // ...
}
```

## Validation as a Provider

For shared validation logic, register as a provider:

```go
type Validator struct {
    v *validator.Validate
}

func NewValidator() *Validator {
    return &Validator{v: validator.New()}
}

func (v *Validator) Struct(s any) []FieldError {
    err := v.v.Struct(s)
    if err == nil {
        return nil
    }
    return translateValidationErrors(err.(validator.ValidationErrors))
}

// Register as provider
module.ProviderDef{
    Token: "validator",
    Build: func(r module.Resolver) (any, error) {
        return NewValidator(), nil
    },
}
```

## Comparison with NestJS

| NestJS | modkit |
|--------|--------|
| `@Body()` with class-validator | `json.Decode()` + explicit validation |
| `ValidationPipe` | Validation in handler or middleware |
| Decorators (`@IsEmail()`) | Struct tags or validation functions |
| Automatic transformation | Explicit type conversion |

## Tips

- Validate early—check input before business logic
- Return all validation errors at once, not one at a time
- Use consistent error response format across endpoints
- Consider validation libraries for complex rules
- Keep validation logic testable (pure functions)
- Document expected formats in API docs/OpenAPI

## See example

- [Validation helpers package](../../examples/hello-mysql/internal/validation/)
- [RFC 7807 Problem Details writer](../../examples/hello-mysql/internal/validation/problem_details.go)
- [User input validation structs](../../examples/hello-mysql/internal/modules/users/types.go)
- [Controller validation paths](../../examples/hello-mysql/internal/modules/users/controller.go)
- [Validation-focused tests](../../examples/hello-mysql/internal/modules/users/validation_test.go)
