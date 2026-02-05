# Frequently Asked Questions

## General

### What is modkit?

modkit is a Go framework for building modular backend services. It brings NestJS-style module organization to Go—without reflection, decorators, or magic.

### Why "modkit"?

**Mod**ular tool**kit** for Go.

### Is modkit production-ready?

modkit is in **early development**. APIs may change before v0.1.0. Use it for prototypes, side projects, or evaluation, but expect potential breaking changes.

### What Go version is required?

Go 1.22 or later.

---

## Design Philosophy

### Why no reflection?

Reflection in Go:
- Makes code harder to debug
- Obscures the call graph
- Can lead to runtime surprises
- Doesn't work well with static analysis tools

modkit uses explicit `Build` functions and string tokens. Everything is visible in code.

### Why string tokens instead of types?

String tokens are simple, explicit, and work without reflection. The trade-off is manual type casting when you call `Get()`:

```go
svc, _ := r.Get("users.service")
userService := svc.(UsersService)
```

This is intentional—it keeps the framework small and makes dependencies visible.

### Why singletons only?

modkit only supports singleton scope (one instance per provider). This keeps the model simple and predictable. If you need request-scoped values, pass them through `context.Context`.

### Why modules instead of flat DI?

Modules provide:
- **Boundaries:** Clear separation between features
- **Visibility:** Control what's exposed to other modules
- **Organization:** Natural structure for larger codebases
- **Testability:** Replace entire modules in tests

---

## Comparison

### How does modkit compare to other frameworks?

modkit compares with google/wire, uber-go/fx, samber/do, manual DI, and NestJS. For a detailed comparison, see the [Comparison Guide](guides/comparison.md).

### Should I use modkit or just wire dependencies manually?

For small services, manual DI in `main()` is fine. modkit helps when:
- You have multiple feature modules
- You want visibility enforcement between modules
- You're building a larger service with a team

See [Comparison Guide](guides/comparison.md#no-framework-manual-di) for a detailed analysis.

---

## Modules

### Do modules need to be pointers?

Yes. Modules must be passed as pointers to ensure stable identity when shared across imports:

```go
// Correct
app, _ := kernel.Bootstrap(&AppModule{})

// Wrong - will not work correctly
app, _ := kernel.Bootstrap(AppModule{})
```

### Can I have circular module imports?

No. modkit rejects circular imports with `ModuleCycleError`. Refactor to break the cycle, often by extracting shared dependencies into a separate module.

### What happens if two modules have the same name?

`DuplicateModuleNameError`. Module names must be unique across the graph.

### Can a module import the same dependency twice?

Yes, as long as it's the same pointer instance. modkit deduplicates by pointer identity.

---

## Providers

### When are providers built?

Providers are built lazily—on first `Get()` call, not at bootstrap. This means:
- Unused providers are never built
- Build order depends on resolution order
- Circular dependencies are detected at build time

### Can I have multiple providers with the same token?

No. `DuplicateProviderTokenError`. Each token must be unique within a module.

### How do I provide different implementations for testing?

Create a test module that provides mock implementations:

```go
type TestUsersModule struct{}

func (m *TestUsersModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "users",
        Providers: []module.ProviderDef{{
            Token: "users.service",
            Build: func(r module.Resolver) (any, error) {
                return &MockUsersService{}, nil
            },
        }},
        Exports: []module.Token{"users.service"},
    }
}
```

---

## Controllers

### What interface must controllers implement?

`RouteRegistrar`:

```go
type RouteRegistrar interface {
    RegisterRoutes(router Router)
}
```

### Can I use a different router?

modkit includes a chi-based HTTP adapter, but controllers are just structs with a `RegisterRoutes` method. You can adapt to any router.

### How do I add middleware to specific routes?

Use `Group` and `Use`:

```go
func (c *Controller) RegisterRoutes(r mkhttp.Router) {
    r.Group("/admin", func(r mkhttp.Router) {
        r.Use(adminAuthMiddleware)
        r.Handle(http.MethodGet, "/users", handler)
    })
}
```

---

## HTTP

### What router does modkit use?

[chi](https://github.com/go-chi/chi) v5.

### Does modkit support gRPC?

Not yet. A gRPC adapter is planned for post-MVP.

### How do I handle graceful shutdown?

`mkhttp.Serve` handles SIGINT/SIGTERM automatically. For custom shutdown logic:

```go
server := &http.Server{Addr: ":8080", Handler: router}

go func() {
    <-ctx.Done()
    server.Shutdown(context.Background())
}()

server.ListenAndServe()
```

---

## Errors

### What errors can Bootstrap return?

See [Error Types](reference/api.md#errors) in the API reference.

### How do I return JSON errors from handlers?

Use a helper function or RFC 7807 Problem Details. See [Error Handling Guide](guides/error-handling.md).

---

## Contributing

### How do I contribute?

See [CONTRIBUTING.md](../CONTRIBUTING.md). Start with issues labeled `good first issue`.

### Where do I report bugs?

Open an issue on GitHub using the bug report template.

### Where do I suggest features?

Open an issue on GitHub using the feature request template.
