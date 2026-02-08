# API Reference

This is a quick reference for modkit's core types. For full documentation, see [pkg.go.dev](https://pkg.go.dev/github.com/go-modkit/modkit).

## Packages

| Package | Import | Purpose |
|---------|--------|---------|
| `module` | `github.com/go-modkit/modkit/modkit/module` | Module metadata types |
| `config` | `github.com/go-modkit/modkit/modkit/config` | Typed config loading helpers |
| `kernel` | `github.com/go-modkit/modkit/modkit/kernel` | Graph builder, bootstrap |
| `http` | `github.com/go-modkit/modkit/modkit/http` | HTTP adapter |
| `logging` | `github.com/go-modkit/modkit/modkit/logging` | Logging interface |
| `testkit` | `github.com/go-modkit/modkit/modkit/testkit` | Testing harness and overrides |

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

### Get[T] (Generic Helper)

```go
func Get[T any](r Resolver, token Token) (T, error)
```

Type-safe wrapper around `Resolver.Get`. Returns an error if resolution fails or if the type doesn't match `T`.

### App.Get

```go
func (a *App) Get(token Token) (any, error)
```

Resolves a token from the root module scope. Note that `module.Get[T]` can be used with an `App` instance because `App` implements the `Resolver` interface.

### App.Resolver

```go
func (a *App) Resolver() Resolver
```

Returns a root-scoped resolver that enforces module visibility.

### BootstrapWithOptions

```go
func BootstrapWithOptions(root module.Module, opts ...BootstrapOption) (*App, error)
```

Bootstraps with explicit options. In v1, `WithProviderOverrides` is the mutation option for tests.

```go
type ProviderOverride struct {
    Token   module.Token
    Build   func(module.Resolver) (any, error)
    Cleanup func(context.Context) error
}

func WithProviderOverrides(overrides ...ProviderOverride) BootstrapOption
```

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
| `DuplicateOverrideTokenError` | Override list contains duplicate token |
| `OverrideTokenNotFoundError` | Override targets missing provider token |
| `OverrideTokenNotVisibleFromRootError` | Override token not visible from root |
| `BootstrapOptionConflictError` | Multiple options mutate same token |

---

## http

---

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

## testkit

### New / NewE

```go
func New(tb testkit.TB, root module.Module, opts ...testkit.Option) *testkit.Harness
func NewE(tb testkit.TB, root module.Module, opts ...testkit.Option) (*testkit.Harness, error)
```

Bootstraps a test harness. `New` fails the test on bootstrap error. `NewE` returns the error.

### Harness lifecycle

```go
func (h *Harness) App() *kernel.App
func (h *Harness) Close() error
func (h *Harness) CloseContext(ctx context.Context) error
```

`Close` runs provider cleanup hooks first and then app closers.

### Overrides

```go
func WithOverrides(overrides ...Override) Option
func OverrideValue(token module.Token, value any) Override
func OverrideBuild(token module.Token, build func(module.Resolver) (any, error)) Override
func WithoutAutoClose() Option
```

Applies token-level provider overrides while preserving graph and visibility semantics.

### Typed Helpers

```go
func Get[T any](tb testkit.TB, h *testkit.Harness, token module.Token) T
func GetE[T any](h *testkit.Harness, token module.Token) (T, error)
func Controller[T any](tb testkit.TB, h *testkit.Harness, moduleName, controllerName string) T
func ControllerE[T any](h *testkit.Harness, moduleName, controllerName string) (T, error)
```

Typed wrappers for provider and controller retrieval in tests.

### Key errors

| Type | When |
|------|------|
| `ControllerNotFoundError` | Controller key was not found in harness app |
| `TypeAssertionError` | Typed helper could not assert expected type |
| `HarnessCloseError` | Hook close and/or app close returned errors |

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
                db, err := module.Get[*sql.DB](r, "db.connection")
                if err != nil {
                    return nil, err
                }
                return NewUsersService(db), nil
            },
        }},
        Controllers: []module.ControllerDef{{
            Name: "UsersController",
            Build: func(r module.Resolver) (any, error) {
                svc, err := module.Get[UsersService](r, "users.service")
                if err != nil {
                    return nil, err
                }
                return NewUsersController(svc), nil
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
