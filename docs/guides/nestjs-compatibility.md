# NestJS Compatibility Guide

**Last Reviewed:** 2026-02-08

This guide maps NestJS concepts to modkit equivalents (or intentional differences) to help Go developers understand what carries over from the NestJS model and what changes in a Go-idiomatic framework.

## Feature Matrix

| Category | NestJS Feature | modkit Status | Notes |
|----------|----------------|---------------|-------|
| **Modules** |  |  |  |
|  | Module definition | ✅ Implemented | `ModuleDef` struct vs `@Module()` decorator |
|  | Imports | ✅ Implemented | Same concept |
|  | Exports | ✅ Implemented | Same concept |
|  | Providers | ✅ Implemented | Same concept |
|  | Controllers | ✅ Implemented | Same concept |
|  | Global modules | ⏭️ Skipped | Anti-pattern in Go; prefer explicit imports |
|  | Dynamic modules | ⏭️ Different | Use constructor functions with options |
|  | Module re-exporting | ✅ Implemented | Exporting tokens from imported modules |
| **Providers** |  |  |  |
|  | Singleton scope | ✅ Implemented | Default and only scope |
|  | Request scope | ⏭️ Skipped | Use context.Context instead |
|  | Transient scope | ⏭️ Skipped | Use factory functions if needed |
|  | useClass | ✅ Implemented | Via `Build` function |
|  | useValue | ✅ Implemented | Via `Build` returning static value |
|  | useFactory | ✅ Implemented | `Build` function IS a factory |
|  | useExisting | ⏭️ Skipped | Use token aliases in Build function |
|  | Async providers | ⏭️ Different | Go is sync; use goroutines if needed |
| **Lifecycle** |  |  |  |
|  | onModuleInit | ⏭️ Skipped | Put init logic in `Build()` function |
|  | onApplicationBootstrap | ⏭️ Skipped | Controllers built = app bootstrapped |
|  | onModuleDestroy | ✅ Implemented | Via `io.Closer` interface |
|  | beforeApplicationShutdown | ⏭️ Skipped | Covered by `io.Closer` |
|  | onApplicationShutdown | ✅ Implemented | `App.Close()` method |
|  | enableShutdownHooks | ⏭️ Different | Use `signal.NotifyContext` (Go stdlib) |
| **HTTP** |  |  |  |
|  | Controllers | ✅ Implemented | `RouteRegistrar` interface |
|  | Route decorators | ⏭️ Different | Explicit `RegisterRoutes()` method |
|  | Middleware | ✅ Implemented | Standard `func(http.Handler) http.Handler` |
|  | Guards | ⏭️ Different | Implement as middleware |
|  | Interceptors | ⏭️ Different | Implement as middleware |
|  | Pipes | ⏭️ Different | Validation in handler or middleware |
|  | Exception filters | ⏭️ Different | Error handling middleware |
| **Other** |  |  |  |
|  | CLI scaffolding | ✅ Implemented | `modkit` CLI ships scaffolding commands for apps/modules/providers/controllers |
|  | Devtools | ⏸️ Decision pending | Listed as a P2 decision in the PRD roadmap |
|  | Microservices | ❌ Not planned | Out of scope |
|  | WebSockets | ❌ Not planned | Use gorilla/websocket directly |
|  | GraphQL | ❌ Not planned | Use gqlgen directly |

## Justifications and Alternatives

### Global Modules

**NestJS:** The `@Global()` decorator makes a module's exports available everywhere without explicit imports.

**modkit:** Skipped.

**Justification:** Global modules hide dependencies and weaken module boundaries. In Go, dependencies are explicit at the package and module level, which keeps systems easier to reason about.

**Alternative:** Construct a shared module once and import it explicitly where needed.

```go
configModule := NewConfigModule()

usersModule := NewUsersModule(configModule)
ordersModule := NewOrdersModule(configModule)
```

### Dynamic Modules

**NestJS:** `DynamicModule` lets you compute providers/exports at runtime via `register()` methods.

**modkit:** Different.

**Justification:** Go favors explicit constructors over runtime decorators. Constructor functions are testable, type-safe, and keep configuration visible.

**Alternative:** Use a constructor function that returns a module configured with options.

```go
type CacheOptions struct {
    TTL time.Duration
}

func NewCacheModule(opts CacheOptions) module.Module {
    return &CacheModule{opts: opts}
}
```

### Request Scope

**NestJS:** Providers can be request-scoped so each HTTP request gets its own instance.

**modkit:** Skipped.

**Justification:** Go already has explicit request scoping via `context.Context`. Request data should flow through context, not DI containers.

**Alternative:** Store per-request values in context and read them in handlers or middleware.

```go
type ctxKey string

func withRequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), ctxKey("request_id"), uuid.NewString())
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Transient Scope

**NestJS:** Transient providers create a new instance every time they are injected.

**modkit:** Skipped.

**Justification:** Go code can construct short-lived values directly, which is simpler and more transparent than a container-managed transient scope.

**Alternative:** Use factory functions where you need a fresh instance.

```go
func NewValidator() *Validator {
    return &Validator{now: time.Now}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    v := NewValidator()
    v.Validate(r)
}
```

### useExisting

**NestJS:** `useExisting` creates an alias to another provider token.

**modkit:** Skipped.

**Justification:** Explicit wiring is clearer than hidden aliases. In Go, you can return the existing dependency directly.

**Alternative:** Use a `Build` function that fetches and returns the existing provider.

```go
module.ProviderDef{
    Token: "users.reader",
    // Note: requires fmt for error formatting.
    Build: func(r module.Resolver) (any, error) {
        svc, err := module.Get[*UsersService](r, "users.service")
        if err != nil {
            return nil, err
        }
        return svc, nil
    },
}
```

### Async Providers

**NestJS:** Providers can be async via `useFactory` returning a promise.

**modkit:** Different.

**Justification:** Go initialization is synchronous. If you need concurrency, you launch goroutines explicitly and either return immediately with a readiness signal or block until ready before returning.

**Alternative:** Start background work in a goroutine and return a ready object.

```go
type Cache struct {
    ready chan struct{}
}

func NewCache() *Cache {
    c := &Cache{ready: make(chan struct{})}
    go func() {
        // warm cache
        close(c.ready)
    }()
    return c
}
```

### Lifecycle Hooks

**NestJS:** Multiple lifecycle hooks (`onModuleInit`, `onApplicationBootstrap`, `onModuleDestroy`, etc.).

**modkit:** Different.

**Justification:** Go favors explicit initialization and cleanup via constructors and `io.Closer`. Signal handling is a standard library concern.

**Alternative:** Put startup in `Build` and cleanup in `Close`, and wire shutdown with `signal.NotifyContext`.

```go
type DB struct{ *sql.DB }

func (d *DB) Close() error { return d.DB.Close() }

ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

go func() {
    <-ctx.Done()
    if err := app.Close(); err != nil {
        log.Printf("app close: %v", err)
    }
}()
```

### Route Decorators

**NestJS:** Decorators like `@Get()` and `@Post()` declare routes on controllers.

**modkit:** Different.

**Justification:** Go avoids decorators and reflection. Explicit route registration keeps handlers discoverable and testable.

**Alternative:** Implement `RegisterRoutes` and call `Handle` directly.

```go
func (c *UsersController) RegisterRoutes(r mkhttp.Router) {
    r.Handle("GET", "/users", c.List)
    r.Handle("POST", "/users", c.Create)
}
```

### Guards, Interceptors, Pipes, Exception Filters

**NestJS:** Cross-cutting concerns implemented via framework-specific abstractions.

**modkit:** Different.

**Justification:** Go uses standard middleware chains. This keeps behavior explicit and composable without framework-specific layers.

**Alternative:** Compose middleware for auth, validation, and error handling.

```go
router := mkhttp.NewRouter()
router.Use(RequireAuth)
router.Use(ValidateJSON)
router.Use(RecoverErrors)
```

### CLI Scaffolding

**NestJS:** CLI generates projects, modules, and scaffolding.

**modkit:** Implemented.

**Justification:** `modkit` now ships a dedicated CLI (`cmd/modkit`) with scaffold commands and release artifacts. This improves onboarding while keeping framework runtime behavior explicit and Go-idiomatic.

**Alternative:** You can still use plain Go tooling and manual wiring; the CLI is optional convenience, not required runtime magic.

```go
//go:generate go run ./internal/tools/wire
```

### Devtools

**NestJS:** Framework-specific devtools for inspection and hot reload.

**modkit:** Decision pending.

**Justification:** The PRD tracks devtools as a P2 decision. Current guidance remains to rely on standard Go tooling until a concrete built-in devtools scope is accepted.

**Alternative:** Use standard tooling like `pprof` and `delve`.

```go
import _ "net/http/pprof"

go http.ListenAndServe("localhost:6060", nil)
```

### Microservices

**NestJS:** Built-in microservices package with transport abstractions.

**modkit:** Not planned.

**Justification:** Go already has strong, explicit libraries for RPC and messaging. Keeping it out of modkit avoids locking users into one transport.

**Alternative:** Use gRPC or NATS directly.

```go
grpcServer := grpc.NewServer()
pb.RegisterUsersServer(grpcServer, usersSvc)
```

### WebSockets

**NestJS:** WebSocket gateway abstraction.

**modkit:** Not planned.

**Justification:** Go's ecosystem already provides stable WebSocket libraries with explicit control.

**Alternative:** Use `gorilla/websocket` directly.

```go
upgrader := websocket.Upgrader{}
conn, _ := upgrader.Upgrade(w, r, nil)
defer conn.Close()
```

### GraphQL

**NestJS:** GraphQL module and decorators.

**modkit:** Not planned.

**Justification:** Go GraphQL stacks are best served by specialized libraries with code generation.

**Alternative:** Use `gqlgen` and mount the handler in a controller.

```go
srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: r}))
router.Handle("POST", "/graphql", srv)
```
