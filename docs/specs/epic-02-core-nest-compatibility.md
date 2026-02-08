# Epic 02: Core NestJS Compatibility

**Status:** Implemented (v1)
**Last Reviewed:** 2026-02-08

## Overview

This epic brings modkit to feature parity with NestJS's core module system, implementing the features that make sense in Go's idiomatic context while documenting why certain NestJS features are intentionally skipped or implemented differently.

**Goals:**
1. Implement graceful shutdown with `io.Closer` support
2. Implement module re-exporting
3. Create comprehensive NestJS compatibility documentation

## Deliverables

### 1. Graceful Shutdown

**Problem:** Applications need to cleanly shut down database connections, flush buffers, and release resources when receiving termination signals.

**NestJS approach:** 5 lifecycle hooks (`onModuleInit`, `onApplicationBootstrap`, `onModuleDestroy`, `beforeApplicationShutdown`, `onApplicationShutdown`)

**Go-idiomatic approach:** Leverage the standard `io.Closer` interface and Go's signal handling.

#### API Design

```go
// App gains a Close method
type App struct {
    Controllers map[string]any
    // internal: ordered list of closers
}

// Close shuts down the application gracefully.
// Calls Close() on all providers implementing io.Closer
// in reverse initialization order.
func (a *App) Close() error

// CloseContext is like Close but respects context cancellation.
func (a *App) CloseContext(ctx context.Context) error
```

#### Provider Cleanup

Providers that need cleanup implement `io.Closer`:

```go
type DatabaseConnection struct {
    db *sql.DB
}

func (d *DatabaseConnection) Close() error {
    return d.db.Close()
}
```

The kernel tracks which providers implement `io.Closer` and calls them in reverse order during `App.Close()`.

#### Signal Handling Helper (Optional)

```go
// Serve starts the HTTP server and handles graceful shutdown on SIGINT/SIGTERM.
// This is a convenience wrapper - users can also handle signals themselves.
func ServeWithShutdown(ctx context.Context, addr string, handler http.Handler, app *App) error
```

Or users can handle signals themselves (standard Go pattern):

```go
func main() {
    app, _ := kernel.Bootstrap(&AppModule{})
    
    // Standard Go signal handling
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    
    server := &http.Server{Addr: ":8080", Handler: router}
    
    go func() {
        <-ctx.Done()
        server.Shutdown(context.Background())
        app.Close()
    }()
    
    server.ListenAndServe()
}
```

#### Implementation Details

1. **Track initialization order:** Container records order providers are built
2. **Detect io.Closer:** When provider is built, check if it implements `io.Closer`
3. **Close in reverse order:** On `App.Close()`, iterate closers in reverse
4. **Error aggregation:** Collect all close errors, return as multi-error
5. **Idempotent:** Multiple calls to `Close()` are safe (no-op after first)

#### Acceptance Criteria

- [x] `App.Close()` method implemented
- [x] `App.CloseContext(ctx)` method for timeout support
- [x] Providers implementing `io.Closer` are called in reverse init order
- [x] Multiple close errors aggregated into single error
- [x] Close is idempotent (safe to call multiple times)
- [x] Tests for close ordering
- [x] Tests for error aggregation
- [x] Example updated to demonstrate graceful shutdown
- [x] Documentation in `docs/guides/lifecycle.md` updated

---

### 2. Module Re-exporting

**Problem:** A module may want to re-export tokens from its imports, creating a facade or aggregating multiple modules.

**NestJS example:**
```typescript
@Module({
  imports: [CommonModule],
  exports: [CommonModule],  // Re-export everything from CommonModule
})
export class CoreModule {}
```

**modkit approach:** Allow exporting tokens that come from imported modules.

#### API Design

Current behavior: Exports can only contain tokens from the module's own providers.

New behavior: Exports can also contain:
1. Tokens from own providers (current)
2. Tokens exported by imported modules (new)

```go
func (m *CoreModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name:    "core",
        Imports: []module.Module{m.common, m.config},
        Exports: []module.Token{
            "common.logger",   // Re-export from CommonModule
            "config.settings", // Re-export from ConfigModule
        },
    }
}
```

#### Alternative: Export entire module

Could also support exporting an entire module's exports:

```go
Exports: []module.Token{
    module.All(m.common),  // Re-export all exports from CommonModule
}
```

This is more complex and may not be worth it for v1. Start with explicit token re-export.

#### Implementation Details

1. **Visibility check update:** When validating exports, check if token is:
   - Defined in own providers, OR
   - Exported by an imported module
2. **Transitive visibility:** If A imports B, and B re-exports from C, A can access C's re-exported tokens
3. **No re-exporting non-exported tokens:** Can only re-export what the imported module exports

#### Acceptance Criteria

- [x] Modules can export tokens from imported modules
- [x] Validation: cannot re-export non-exported tokens (clear error)
- [x] Transitive re-exporting works (A→B→C)
- [x] Tests for re-export scenarios
- [x] Tests for invalid re-export attempts
- [x] Documentation updated

---

### 3. NestJS Compatibility Documentation

Create `docs/guides/nestjs-compatibility.md` documenting how modkit maps to NestJS concepts.

#### Document Structure

```markdown
# NestJS Compatibility Guide

This guide documents how modkit implements (or intentionally differs from) 
NestJS concepts for Go developers coming from the Node.js ecosystem.

## Feature Matrix

| NestJS Feature | modkit Status | Notes |
|----------------|---------------|-------|
| ... | ... | ... |

## Detailed Comparison

### Modules
...

### Providers
...

### Lifecycle
...
```

#### Feature Matrix Content

| Category | NestJS Feature | modkit Status | Notes |
|----------|----------------|---------------|-------|
| **Modules** |
| | Module definition | ✅ Implemented | `ModuleDef` struct vs `@Module()` decorator |
| | Imports | ✅ Implemented | Same concept |
| | Exports | ✅ Implemented | Same concept |
| | Providers | ✅ Implemented | Same concept |
| | Controllers | ✅ Implemented | Same concept |
| | Global modules | ⏭️ Skipped | Anti-pattern in Go; prefer explicit imports |
| | Dynamic modules | ⏭️ Different | Use constructor functions with options |
| | Module re-exporting | ✅ Implemented | Exporting tokens from imported modules |
| **Providers** |
| | Singleton scope | ✅ Implemented | Default and only scope |
| | Request scope | ⏭️ Skipped | Use context.Context instead |
| | Transient scope | ⏭️ Skipped | Use factory functions if needed |
| | useClass | ✅ Implemented | Via `Build` function |
| | useValue | ✅ Implemented | Via `Build` returning static value |
| | useFactory | ✅ Implemented | `Build` function IS a factory |
| | useExisting | ⏭️ Skipped | Use token aliases in Build function |
| | Async providers | ⏭️ Different | Go is sync; use goroutines if needed |
| **Lifecycle** |
| | onModuleInit | ⏭️ Skipped | Put init logic in `Build()` function |
| | onApplicationBootstrap | ⏭️ Skipped | Controllers built = app bootstrapped |
| | onModuleDestroy | ✅ Implemented | Via `io.Closer` interface |
| | beforeApplicationShutdown | ⏭️ Skipped | Covered by `io.Closer` |
| | onApplicationShutdown | ✅ Implemented | `App.Close()` method |
| | enableShutdownHooks | ⏭️ Different | Use `signal.NotifyContext` (Go stdlib) |
| **HTTP** |
| | Controllers | ✅ Implemented | `RouteRegistrar` interface |
| | Route decorators | ⏭️ Different | Explicit `RegisterRoutes()` method |
| | Middleware | ✅ Implemented | Standard `func(http.Handler) http.Handler` |
| | Guards | ⏭️ Different | Implement as middleware |
| | Interceptors | ⏭️ Different | Implement as middleware |
| | Pipes | ⏭️ Different | Validation in handler or middleware |
| | Exception filters | ⏭️ Different | Error handling middleware |
| **Other** |
| | CLI scaffolding | ✅ Implemented | `modkit` CLI ships project/module/provider/controller generators |
| | Devtools | ⏸️ Decision pending | Roadmap item remains open in PRD (P2 decision) |
| | Microservices | ❌ Not planned | Out of scope |
| | WebSockets | ❌ Not planned | Use gorilla/websocket directly |
| | GraphQL | ❌ Not planned | Use gqlgen directly |

#### Justification Sections

For each "Skipped" or "Different" feature, document:
1. What NestJS does
2. Why it doesn't fit Go
3. The Go-idiomatic alternative

Example:

```markdown
### Global Modules

**NestJS:** The `@Global()` decorator makes a module's exports available 
everywhere without explicit imports.

**modkit:** Not implemented.

**Justification:** Global modules contradict Go's explicit dependency philosophy.
In Go, if a package needs something, it imports it explicitly. This makes 
dependencies visible in code and easier to trace. modkit's core value 
proposition is visibility enforcement - adding global modules would undermine this.

**Alternative:** If you need a provider available to many modules:
1. Create the module once with a constructor function
2. Import it explicitly where needed
3. The singleton nature means all modules share the same instance

​```go
// Create once
configModule := NewConfigModule()

// Import explicitly where needed
usersModule := NewUsersModule(configModule)
ordersModule := NewOrdersModule(configModule)
​```
```

#### Acceptance Criteria

- [x] `docs/guides/nestjs-compatibility.md` created
- [x] Feature matrix with all major NestJS features
- [x] Justification for each skipped/different feature
- [x] Go-idiomatic alternatives documented
- [x] Cross-linked from README and other relevant docs

---

## Stories Breakdown

### Story 2.1: Graceful Shutdown - Core Implementation
**Points:** 3

- Implement `App.Close()` method
- Track provider initialization order in Container
- Detect `io.Closer` implementations
- Close in reverse order
- Unit tests for close ordering

### Story 2.2: Graceful Shutdown - Error Handling
**Points:** 2

- Aggregate multiple close errors
- Make Close idempotent
- Implement `App.CloseContext(ctx)` with timeout
- Tests for error scenarios

### Story 2.3: Graceful Shutdown - Example & Docs
**Points:** 2

- Update hello-mysql example with graceful shutdown
- Update `docs/guides/lifecycle.md`
- Add signal handling example

### Story 2.4: Module Re-exporting - Implementation
**Points:** 3

- Update visibility validation to allow re-exports
- Update graph builder for transitive re-exports
- Validate re-exports are actually exported by import
- Unit tests for re-export scenarios

### Story 2.5: Module Re-exporting - Docs
**Points:** 1

- Update `docs/guides/modules.md` with re-export examples
- Add error message examples for invalid re-exports

### Story 2.6: NestJS Compatibility Documentation
**Points:** 3

- Create `docs/guides/nestjs-compatibility.md`
- Write feature matrix
- Write justifications for skipped features
- Write Go-idiomatic alternatives
- Cross-link from README

---

## Total Estimate

| Story | Points |
|-------|--------|
| 2.1 Graceful Shutdown - Core | 3 |
| 2.2 Graceful Shutdown - Errors | 2 |
| 2.3 Graceful Shutdown - Docs | 2 |
| 2.4 Module Re-exporting | 3 |
| 2.5 Re-exporting Docs | 1 |
| 2.6 NestJS Compatibility Docs | 3 |
| **Total** | **14** |

---

## Dependencies

- None (builds on existing kernel/module packages)

## Risks

1. **Close ordering complexity:** If providers have circular dependencies (which should be rejected), close ordering is undefined. Mitigation: Circular deps already rejected at build time.

2. **io.Closer detection:** Interface detection in Go is implicit. Risk: Some types might unexpectedly implement io.Closer. Mitigation: Only check providers, document behavior clearly.

## Success Metrics

- hello-mysql example demonstrates clean shutdown
- Users can migrate from NestJS with clear documentation
- No breaking changes to existing API
