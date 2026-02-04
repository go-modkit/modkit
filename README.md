# modkit

[![Go Reference](https://pkg.go.dev/badge/github.com/aryeko/modkit.svg)](https://pkg.go.dev/github.com/aryeko/modkit)
[![CI](https://github.com/aryeko/modkit/actions/workflows/ci.yml/badge.svg)](https://github.com/aryeko/modkit/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/aryeko/modkit)](https://goreportcard.com/report/github.com/aryeko/modkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**A Go framework for building modular backend services, inspired by NestJS.**

modkit brings NestJS-style module organization to Go—without reflection, decorators, or magic. Define modules with explicit imports, providers, controllers, and exports. The kernel builds a dependency graph, enforces visibility, and bootstraps your app deterministically.

## Why modkit?

modkit is a Go-idiomatic alternative to decorator-driven frameworks. It keeps wiring explicit, avoids reflection, and makes module boundaries and dependencies visible in code.

| If you want... | modkit gives you... |
|----------------|---------------------|
| NestJS-style modules in Go | `imports`, `providers`, `controllers`, `exports` |
| Explicit dependency injection | String tokens + resolver, no reflection |
| Debuggable bootstrap | Deterministic graph construction with clear errors |
| Minimal framework overhead | Thin HTTP adapter on chi, no ORM, no config magic |

### Compared to Other Go Frameworks

| If you use... | modkit is different because... |
|---------------|--------------------------------|
| google/wire   | modkit adds module boundaries + visibility enforcement |
| uber-go/fx    | No reflection, explicit Build functions |
| samber/do     | Full module system with imports/exports |
| No framework  | Structured module organization without boilerplate |

See the [full comparison](docs/guides/comparison.md) for details.

## Quick Example

```go
// Define a module
type UsersModule struct{}

func (m *UsersModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "users",
        Providers: []module.ProviderDef{{
            Token: "users.service",
            Build: func(r module.Resolver) (any, error) {
                return NewUsersService(), nil
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

// Bootstrap and serve
func main() {
    app, _ := kernel.Bootstrap(&UsersModule{})
    router := mkhttp.NewRouter()
    mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers)
    mkhttp.Serve(":8080", router)
}
```

## Installation

```bash
go get github.com/aryeko/modkit
```

Requires Go 1.25+

## Features

- **Module System** — Compose apps from self-contained modules with explicit boundaries
- **Dependency Injection** — Providers built on first access, cached as singletons
- **Visibility Enforcement** — Only exported tokens are accessible to importers
- **HTTP Adapter** — Chi-based router with explicit route registration
- **No Reflection** — Everything is explicit and type-safe
- **Deterministic Bootstrap** — Predictable initialization order with clear error messages

## Packages

| Package | Description |
|---------|-------------|
| `modkit/module` | Module metadata types (`ModuleDef`, `ProviderDef`, `Token`) |
| `modkit/kernel` | Graph builder, visibility enforcer, bootstrap |
| `modkit/http` | HTTP adapter for chi router |
| `modkit/logging` | Logging interface with slog adapter |

## Architecture

```mermaid
flowchart LR
    subgraph Input
        A[📦 Module Definitions]
    end
    
    subgraph Kernel
        B[🔗 Graph Builder]
        C[📦 Container]
    end
    
    subgraph Output
        D[🎮 Controllers]
        E[🌐 HTTP Adapter]
    end
    
    A --> B
    B --> C
    C --> D
    D --> E
    
    style A fill:#e1f5fe,stroke:#01579b,color:#01579b
    style B fill:#fff3e0,stroke:#e65100,color:#e65100
    style C fill:#fff3e0,stroke:#e65100,color:#e65100
    style D fill:#e8f5e9,stroke:#2e7d32,color:#2e7d32
    style E fill:#e8f5e9,stroke:#2e7d32,color:#2e7d32
```

See [Architecture Guide](docs/architecture.md) for details.

## Documentation

**Guides:**
- [Getting Started](docs/guides/getting-started.md) — Your first modkit app
- [Modules](docs/guides/modules.md) — Module composition and visibility
- [Providers](docs/guides/providers.md) — Dependency injection and lifecycle
- [Controllers](docs/guides/controllers.md) — HTTP handlers and routing
- [Middleware](docs/guides/middleware.md) — Request/response middleware
- [Error Handling](docs/guides/error-handling.md) — Error patterns and Problem Details
- [Validation](docs/guides/validation.md) — Input validation patterns
- [Authentication](docs/guides/authentication.md) — Auth middleware and guards
- [Testing](docs/guides/testing.md) — Testing patterns
- [Comparison](docs/guides/comparison.md) — vs Wire, Fx, and others

**Reference:**
- [API Reference](docs/reference/api.md) — Types and functions
- [Architecture](docs/architecture.md) — How modkit works under the hood
- [FAQ](docs/faq.md) — Common questions

**Examples:**
- [hello-simple](examples/hello-simple/) — Minimal example, no Docker
- [hello-mysql](examples/hello-mysql/) — Full CRUD API with MySQL

## How It Compares to NestJS

| Concept | NestJS | modkit |
|---------|--------|--------|
| Module definition | `@Module()` decorator | `ModuleDef` struct |
| Dependency injection | Constructor injection via metadata | Explicit `r.Get(token)` |
| Route binding | `@Get()`, `@Post()` decorators | `RegisterRoutes(router)` method |
| Middleware | `NestMiddleware` interface | `func(http.Handler) http.Handler` |
| Guards/Pipes/Interceptors | Framework abstractions | Standard Go middleware |

## Status

modkit is in **early development**. APIs may change before v0.1.0.

After v0.1.0, changes will follow semantic versioning.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). We welcome issues, discussions, and PRs.

## License

MIT — see [LICENSE](LICENSE)
