# AI Agent Guidelines

This file provides AI-focused instructions for code generation and modifications. For human contributors, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Project Identity

modkit is a **Go framework for building modular backend services, inspired by NestJS**. It provides:
- Module system with explicit imports/exports and visibility enforcement
- No reflection, no decorators, no magic
- Explicit dependency injection via string tokens
- Chi-based HTTP adapter
- Go-idiomatic patterns (explicit errors, standard `http.Handler` middleware)

## Project Structure

```
modkit/
├── modkit/              # Core library packages
│   ├── module/          # Module metadata types
│   ├── kernel/          # Graph builder, bootstrap
│   ├── http/            # HTTP adapter (chi-based)
│   └── logging/         # Logging interface
├── examples/            # Example applications
│   ├── hello-simple/    # Minimal example (no dependencies)
│   └── hello-mysql/     # Full CRUD example with DB
├── docs/
│   ├── guides/          # User guides
│   ├── reference/       # API reference
│   └── architecture.md  # How modkit works
└── .github/internal/plans/  # Implementation tracking (internal)
```

## Development Workflow

**Before making changes:**
```bash
make fmt    # Format code (gofmt + goimports)
make lint   # Run golangci-lint (must pass)
make test   # Run all tests (must pass)
```

## Code Generation Guidelines

### Modules
- Modules must be pointers (`&AppModule{}`)
- `Definition()` must be pure/deterministic
- Module names must be unique
- Use constructor functions for modules with configuration

### Providers
- Token convention: `"module.component"` (e.g., `"users.service"`)
- Providers are lazy singletons (built on first `Get()`)
- Always check `err` from `r.Get()`
- Type assert the resolved value

### Controllers
- Must implement `RegisterRoutes(router Router)`
- Use `r.Handle(method, pattern, handler)` for routes
- Use `r.Group()` and `r.Use()` for scoped middleware

### Error Handling
- Return errors, don't panic
- Use sentinel errors for known conditions
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- No global exception handlers (explicit in handlers/middleware)

### Testing
- Unit tests alongside code (`*_test.go`)
- Table-driven tests for multiple cases
- Bootstrap real modules in integration tests
- Use testcontainers for smoke tests with external deps

## Commit Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):
```
<type>(<scope>): <subject>

[optional body]
```

Valid types: `feat`, `fix`, `docs`, `test`, `chore`, `refactor`, `perf`, `ci`

**Examples:**
- `feat(http): add graceful shutdown to Serve`
- `fix(kernel): detect provider cycles correctly`
- `docs: add lifecycle guide`

## Pull Requests

See [.github/pull_request_template.md](.github/pull_request_template.md) for the complete template.

**Key requirements:**
1. Run `make fmt && make lint && make test` before submitting
2. Check all applicable type boxes
3. Complete the validation section with command outputs
4. Document any breaking changes

## Documentation

- User guides: `docs/guides/*.md`
- API reference: `docs/reference/api.md`
- Architecture deep-dive: `docs/architecture.md`
- Keep examples in sync with docs
- Use Mermaid for diagrams where helpful

## Principles

- **Explicit over implicit**: No reflection, no magic
- **Go-idiomatic**: Prefer language patterns over framework abstractions
- **Minimal API surface**: Export only what users need
- **Clear errors**: Typed errors with helpful messages
- **Testability**: Easy to test with standard Go testing
