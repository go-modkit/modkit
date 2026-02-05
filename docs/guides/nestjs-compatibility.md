# NestJS Compatibility Guide

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
|  | Module re-exporting | 🔄 This Epic | Exporting tokens from imported modules |
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
|  | onModuleDestroy | ✅ This Epic | Via `io.Closer` interface |
|  | beforeApplicationShutdown | ⏭️ Skipped | Covered by `io.Closer` |
|  | onApplicationShutdown | ✅ This Epic | `App.Close()` method |
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
|  | CLI scaffolding | ❌ Not planned | Go boilerplate is minimal |
|  | Devtools | ❌ Not planned | Use standard Go tooling |
|  | Microservices | ❌ Not planned | Out of scope |
|  | WebSockets | ❌ Not planned | Use gorilla/websocket directly |
|  | GraphQL | ❌ Not planned | Use gqlgen directly |
