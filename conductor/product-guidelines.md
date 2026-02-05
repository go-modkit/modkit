# Product Guidelines

## Design Philosophy

### Explicit over Implicit
`modkit` rejects "magic." We do not use autowiring, classpath scanning, or excessive runtime reflection. All module composition, provider registration, and route definitions must be explicitly declared in the code. This ensures that any developer can read the `main.go` or module definition and understand exactly how the application is wired together.

### Idiomatic Go
While inspired by NestJS, `modkit` is first and foremost a Go framework. We adapt concepts to fit the language.
- **No Decorators:** We use struct methods (like `Definition()`) and explicit registration functions instead of decorators.
- **Error Handling:** We use standard Go error returns (`val, err := ...`) rather than exception-based control flow.
- **Context:** `context.Context` is a first-class citizen and must be propagated through all request handlers and providers.

### Developer Experience (DX)
- **Deterministic Bootstrap:** The application graph is built once at startup. If it fails, it must fail with a clear, actionable error message explaining exactly which dependency is missing or circular.
- **Type Safety:** The API is designed to catch as many configuration errors as possible at compile time or early startup, rather than at runtime.

## Quality Standards

### Performance
- **Zero-Cost Abstractions:** The core dependency injection and module system usually run only at startup. Runtime overhead for request processing (HTTP routing, middleware) must be comparable to using raw `chi` or `net/http`.
- **Allocation Efficiency:** Hot paths (request lifecycle) should be optimized to minimize memory allocations.

### Dependencies
- **Minimal Core:** The `kernel`, `module`, and `di` packages should remain dependency-free or rely only on standard library equivalents.
- **Opt-in Ecosystem:** Extensions (like specific database drivers, complex loggers, or validation libraries) should be separate packages or modules, keeping the core lightweight.

### Documentation & Reliability
- **GoDoc:** All exported functions and types must have comprehensive GoDoc comments.
- **Examples:** Every major feature (e.g., a new middleware type or DI pattern) must be demonstrated in the `examples/` directory.
- **Testing:** We maintain high test coverage, particularly for the DI container and graph resolution logic, to ensure stability across complex dependency trees.
