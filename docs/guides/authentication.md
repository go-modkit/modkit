# Authentication

modkit uses standard Go middleware for authentication—no framework-specific guards or decorators. This guide covers common authentication patterns.

## Philosophy

In NestJS, guards are decorators that protect routes. In modkit (and idiomatic Go), authentication is middleware that:

1. Extracts credentials from the request
2. Validates the credentials
3. Attaches user info to the request context
4. Allows or denies access

## Basic Auth Middleware

### Bearer Token Authentication

```go
type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(validateToken func(string) (*User, error)) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            user, err := validateToken(token)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), UserContextKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func extractBearerToken(r *http.Request) string {
    auth := r.Header.Get("Authorization")
    if !strings.HasPrefix(auth, "Bearer ") {
        return ""
    }
    return strings.TrimPrefix(auth, "Bearer ")
}
```

### Retrieving User in Handlers

```go
func (c *UsersController) GetProfile(w http.ResponseWriter, r *http.Request) {
    user := r.Context().Value(UserContextKey).(*User)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// Type-safe helper
func UserFromContext(ctx context.Context) (*User, bool) {
    user, ok := ctx.Value(UserContextKey).(*User)
    return user, ok
}
```

## JWT Authentication

### JWT Middleware

```go
import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    UserID int    `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func JWTMiddleware(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tokenStr := extractBearerToken(r)
            if tokenStr == "" {
                http.Error(w, "missing token", http.StatusUnauthorized)
                return
            }

            token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
                if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("unexpected signing method")
                }
                return secret, nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            claims := token.Claims.(*Claims)
            ctx := context.WithValue(r.Context(), UserContextKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Generating Tokens

```go
func (s *AuthService) GenerateToken(user *User) (string, error) {
    claims := Claims{
        UserID: user.ID,
        Email:  user.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secret)
}
```

## API Key Authentication

```go
func APIKeyMiddleware(validKeys map[string]bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := r.Header.Get("X-API-Key")
            if key == "" {
                key = r.URL.Query().Get("api_key")
            }

            if !validKeys[key] {
                http.Error(w, "invalid API key", http.StatusUnauthorized)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Role-Based Authorization

### Role Middleware

```go
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    allowed := make(map[string]bool)
    for _, r := range roles {
        allowed[r] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user, ok := UserFromContext(r.Context())
            if !ok {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            if !allowed[user.Role] {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Applying to Routes

```go
func (c *AdminController) RegisterRoutes(r mkhttp.Router) {
    r.Group("/admin", func(r mkhttp.Router) {
        r.Use(RequireRole("admin", "superadmin"))
        
        r.Handle(http.MethodGet, "/users", http.HandlerFunc(c.ListUsers))
        r.Handle(http.MethodDelete, "/users/{id}", http.HandlerFunc(c.DeleteUser))
    })
}
```

## Auth Middleware as Provider

For middleware that needs dependencies, register as a provider:

```go
func (m *AuthModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "auth",
        Providers: []module.ProviderDef{
            {
                Token: "auth.middleware",
                Build: func(r module.Resolver) (any, error) {
                    svc, err := r.Get(TokenAuthService)
                    if err != nil {
                        return nil, err
                    }
                    return JWTMiddleware(svc.(*AuthService).Secret()), nil
                },
            },
        },
        Exports: []module.Token{"auth.middleware"},
    }
}
```

Usage at startup:

```go
app, _ := kernel.Bootstrap(&AppModule{})

authMW, _ := app.Get("auth.middleware")
router.Use(authMW.(func(http.Handler) http.Handler))
```

## Optional Authentication

For routes where auth is optional:

```go
func OptionalAuth(validateToken func(string) (*User, error)) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)
            if token != "" {
                if user, err := validateToken(token); err == nil {
                    ctx := context.WithValue(r.Context(), UserContextKey, user)
                    r = r.WithContext(ctx)
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Public + Protected Routes

Mix public and protected routes in a controller:

```go
func (c *UsersController) RegisterRoutes(r mkhttp.Router) {
    // Public routes
    r.Handle(http.MethodPost, "/login", http.HandlerFunc(c.Login))
    r.Handle(http.MethodPost, "/register", http.HandlerFunc(c.Register))

    // Protected routes
    r.Group("/users", func(r mkhttp.Router) {
        r.Use(c.authMiddleware)
        
        r.Handle(http.MethodGet, "/me", http.HandlerFunc(c.GetProfile))
        r.Handle(http.MethodPut, "/me", http.HandlerFunc(c.UpdateProfile))
    })
}
```

## Testing Authenticated Routes

```go
func TestGetProfile_Authenticated(t *testing.T) {
    user := &User{ID: 1, Email: "test@example.com"}
    
    req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
    ctx := context.WithValue(req.Context(), UserContextKey, user)
    req = req.WithContext(ctx)
    
    rec := httptest.NewRecorder()
    controller.GetProfile(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}

func TestGetProfile_Unauthenticated(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
    rec := httptest.NewRecorder()
    
    // Apply middleware + handler
    handler := authMiddleware(http.HandlerFunc(controller.GetProfile))
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got %d", rec.Code)
    }
}
```

## Comparison with NestJS

| NestJS | modkit |
|--------|--------|
| `@UseGuards(AuthGuard)` | `r.Use(authMiddleware)` |
| `@Roles('admin')` | `r.Use(RequireRole("admin"))` |
| `@Request() req` | `r.Context().Value(UserContextKey)` |
| Passport.js strategies | Direct implementation or libraries |
| `JwtModule` | JWT library + middleware |

## Tips

- Use context for passing user info—it's the Go-idiomatic approach
- Create type-safe context helpers (`UserFromContext`)
- Keep auth middleware thin—delegate to services
- Use different middleware for different auth strategies
- Return 401 for "who are you?" and 403 for "you can't do that"
- Test both authenticated and unauthenticated scenarios
