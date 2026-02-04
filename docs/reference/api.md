# API Reference

This is a quick reference for modkit's core types. For full documentation, see [pkg.go.dev](https://pkg.go.dev/github.com/aryeko/modkit).

## Packages

| Package | Import | Purpose |
|---------|--------|---------|
| `module` | `github.com/aryeko/modkit/modkit/module` | Module metadata types |
| `kernel` | `github.com/aryeko/modkit/modkit/kernel` | Graph builder, bootstrap |
| `http` | `github.com/aryeko/modkit/modkit/http` | HTTP adapter |
| `logging` | `github.com/aryeko/modkit/modkit/logging` | Logging interface |

---

## module

### Module Interface

```go
type Module interface {
    Definition() ModuleDef
}
```

Every module must implement this interface. Modules must be passed as pointers.

### ModuleDef

```go
type ModuleDef struct {
    Name        string
    Imports     []Module
    Providers   []ProviderDef
    Controllers []ControllerDef
    Exports     []Token
}
```

| Field | Description |
|-------|-------------|
| `Name` | Unique identifier for the module |
| `Imports` | Modules this module depends on |
| `Providers` | Services/values created by this module |
| `Controllers` | HTTP controllers created by this module |
| `Exports` | Tokens visible to modules that import this one |

### ProviderDef

```go
type ProviderDef struct {
    Token Token
    Build func(Resolver) (any, error)
}
```

| Field | Description |
|-------|-------------|
| `Token` | Unique identifier for the provider |
| `Build` | Factory function called on first `Get()` |

### ControllerDef

```go
type ControllerDef struct {
    Name  string
    Build func(Resolver) (any, error)
}
```

| Field | Description |
|-------|-------------|
| `Name` | Unique identifier within the module |
| `Build` | Factory function that creates the controller |

### Token

```go
type Token string
```

String identifier for providers. Convention: `module.component` (e.g., `users.service`).

### Resolver

```go
type Resolver interface {
    Get(token Token) (any, error)
}
```

Used in `Build` functions to retrieve dependencies.

---

## kernel

### Bootstrap

```go
func Bootstrap(root Module) (*App, error)
```

Entry point for bootstrapping your application. Returns an `App` with built controllers and a container for accessing providers.

### App

```go
type App struct {
    Controllers map[string]any
    Container   Container
}
```

| Field | Description |
|-------|-------------|
| `Controllers` | Map of controller name → controller instance |
| `Container` | Access to the provider container |

### Container

```go
type Container interface {
    Get(token Token) (any, error)
}
```

Provides access to built providers. Used for manual resolution or cleanup.

### Errors

| Type | When |
|------|------|
| `RootModuleNilError` | `Bootstrap(nil)` |
| `DuplicateModuleNameError` | Two modules have the same name |
| `ModuleCycleError` | Circular module imports |
| `DuplicateProviderTokenError` | Same token registered twice |
| `ProviderNotFoundError` | `Get()` with unknown token |
| `TokenNotVisibleError` | Token not exported to requester |
| `ProviderCycleError` | Provider depends on itself |
| `ProviderBuildError` | Provider's `Build` function failed |
| `ControllerBuildError` | Controller's `Build` function failed |

---

## http

### NewRouter

```go
func NewRouter() *chi.Mux
```

Creates a new chi router with baseline middleware (request ID, recoverer).

### RegisterRoutes

```go
func RegisterRoutes(router Router, controllers map[string]any) error
```

Registers all controllers that implement `RouteRegistrar`.

### AsRouter

```go
func AsRouter(mux *chi.Mux) Router
```

Wraps a chi router to implement the `Router` interface.

### Router Interface

```go
type Router interface {
    Handle(method, pattern string, handler http.Handler)
    Group(pattern string, fn func(Router))
    Use(middleware ...func(http.Handler) http.Handler)
}
```

### RouteRegistrar Interface

```go
type RouteRegistrar interface {
    RegisterRoutes(router Router)
}
```

Controllers must implement this interface.

### Serve

```go
func Serve(addr string, handler http.Handler) error
```

Starts an HTTP server with graceful shutdown on SIGINT/SIGTERM.

---

## logging

### Logger Interface

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    With(args ...any) Logger
}
```

Generic logging interface. Use `logging.NewSlogLogger(slog.Default())` for slog integration.

### NewSlogLogger

```go
func NewSlogLogger(logger *slog.Logger) Logger
```

Wraps a `*slog.Logger` to implement the `Logger` interface.

### NewNopLogger

```go
func NewNopLogger() Logger
```

Returns a no-op logger (useful for testing).

---

## Common Patterns

### Basic Module

```go
type UsersModule struct {
    db *DatabaseModule
}

func NewUsersModule(db *DatabaseModule) *UsersModule {
    return &UsersModule{db: db}
}

func (m *UsersModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name:    "users",
        Imports: []module.Module{m.db},
        Providers: []module.ProviderDef{{
            Token: "users.service",
            Build: func(r module.Resolver) (any, error) {
                db, _ := r.Get("db.connection")
                return NewUsersService(db.(*sql.DB)), nil
            },
        }},
        Controllers: []module.ControllerDef{{
            Name: "UsersController",
            Build: func(r module.Resolver) (any, error) {
                svc, _ := r.Get("users.service")
                return NewUsersController(svc.(UsersService)), nil
            },
        }},
        Exports: []module.Token{"users.service"},
    }
}
```

### Bootstrap and Serve

```go
func main() {
    app, err := kernel.Bootstrap(&AppModule{})
    if err != nil {
        log.Fatal(err)
    }

    router := mkhttp.NewRouter()
    if err := mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers); err != nil {
        log.Fatal(err)
    }

    log.Println("Listening on :8080")
    if err := mkhttp.Serve(":8080", router); err != nil {
        log.Fatal(err)
    }
}
```
