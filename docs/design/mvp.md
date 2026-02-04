# modkit — MVP Design Document

**Status:** Draft (MVP)

**Goal:** Define a Go-idiomatic backend service framework with a NestJS-inspired *module architecture* (imports/providers/controllers/exports), starting simple (HTTP + MySQL). The MVP must be implementable by coding agents with minimal ambiguity and must produce verifiable outputs.

> NestJS reference: The module metadata model (imports, providers, controllers, exports) is inspired by NestJS modules. See NestJS Modules docs. (Reference only; implementation is Go-idiomatic.)

---

## 1. Product definition

### 1.1 What modkit is
modkit is a modular application kernel and thin web adapter for building backend services in Go with:
- **Nest-like modular composition**: `imports`, `providers`, `controllers`, `exports`
- **No reflection-based auto-wiring**: explicit tokens and resolver
- **Single responsibility boundaries**: modules provide composition; business logic remains plain Go
- **Adapters-first**: HTTP/MySQL are first adapters; architecture supports adding gRPC/jobs later

### 1.2 What modkit is not (MVP)
- Not a full Nest feature parity project
- No decorators, metadata scanning, or auto-route binding
- No scopes other than singleton
- No plugin marketplace or dynamic runtime loading
- No ORM framework

### 1.3 MVP scope
- Module system with strict Nest-like semantics
- Deterministic build: module graph → provider container → controller instances
- HTTP adapter using `net/http` + `chi`
- MySQL integration via `database/sql` + sqlc (in consuming app)
- CLI is optional for MVP; if included, it only scaffolds a module skeleton

---

## 2. Non-functional requirements

### 2.1 Go-idiomatic principles
- Explicit construction and configuration
- Interfaces only at boundaries; avoid “interface everywhere”
- Prefer code generation over runtime magic (sqlc, oapi-codegen later)
- Minimal dependencies; vendor only what’s needed

### 2.2 Determinism & debuggability
- Bootstrap must be deterministic and testable
- Errors must include module/provider token context
- Panics are acceptable only for truly programmer errors in internal code; public API should return errors

### 2.3 Stability promises
- MVP public APIs in `modkit/*` are stable within minor versions after v0.1.0
- Anything under `internal/` is not stable

---

## 3. Core architecture

### 3.1 NestJS mapping

| NestJS concept | modkit MVP concept |
|---|---|
| `@Module({ imports, providers, controllers, exports })` | `module.ModuleDef` returned by `module.Module.Definition()` |
| Provider token = class | `module.Token` string (type-safe generics may be added later) |
| DI container | `kernel.Container` implementing `module.Resolver` |
| Module import/export visibility | Enforced by `kernel.Graph` and `kernel.Visibility` |
| `NestFactory.create(AppModule)` | `kernel.Bootstrap(rootModule)` |

### 3.2 Layering

- **module package**: declarations only (metadata, tokens, provider/controller descriptors)
- **kernel package**: graph building, container, visibility enforcement, bootstrap
- **http package**: adapter from controller instances to router; shared middleware helpers
- **examples/consuming app**: validates how modkit is consumed; includes MySQL/sqlc integration

### 3.3 MVP runtime model

1) User constructs root module (`AppModule`) via `NewAppModule(options)`
2) `kernel.Bootstrap(root)`:
   - Flattens module graph (imports first)
   - Validates graph (cycles, duplicate module names/IDs)
   - Builds visibility map per module based on exports
   - Registers provider factories in container
   - Instantiates controllers (as singletons)
   - Returns `App` containing container + module graph + controller registry
3) HTTP adapter mounts routes by calling module route registration functions (no reflection)

---

## 4. Repository structure

### 4.1 GitHub repo
- **Repo name:** `modkit`
- **Go module path:** `github.com/<org>/modkit`

### 4.2 Directory layout

```
modkit/
  go.mod
  README.md
  LICENSE
  CONTRIBUTING.md
  CODE_OF_CONDUCT.md
  SECURITY.md
  docs/
    design/
      mvp.md                 (this document)
      module-model.md        (module semantics + examples)
      kernel.md              (graph/container/visibility)
      http-adapter.md        (routing + middleware)
    guides/
      getting-started.md
      modules.md
      testing.md
    adr/
      0001-tokens-over-reflection.md
      0002-singleton-only-mvp.md
  modkit/
    module/
      module.go
      token.go
      provider.go
      controller.go
      errors.go
    kernel/
      bootstrap.go
      graph.go
      visibility.go
      container.go
      errors.go
    http/
      server.go
      router.go
      middleware.go
      errors.go
  examples/
    hello-mysql/
      go.mod               (or use workspace)
      cmd/api/main.go
      internal/
        modules/
          app/
          module.go
          database/
          module.go
          users/
            module.go
            controller.go
            routes.go
            service.go
            repository.go
            repo_mysql.go
        platform/
          config/
          config.go
          mysql/
            db.go
          logging/
            log.go
      sql/
        queries.sql
        sqlc.yaml
      migrations/
      Makefile
      README.md
  .github/
    workflows/
      ci.yml
```

Notes:
- Public library code lives under `modkit/` (import path `github.com/<org>/modkit/modkit/...`).
- Example app is a separate consumer to validate usage end-to-end.

---

## 5. Public API specification (MVP)

### 5.1 `modkit/module`

**Responsibility:** Define module metadata model inspired by NestJS modules.

#### Types

- `type Token string`
- `type Resolver interface { Get(Token) (any, error) }`

- `type ProviderDef struct {
    Token Token
    Build func(r Resolver) (any, error)
  }`

- `type ControllerDef struct {
    Name string
    Build func(r Resolver) (any, error)
  }`

- `type ModuleDef struct {
    Name string
    Imports []Module
    Providers []ProviderDef
    Controllers []ControllerDef
    Exports []Token
  }`

- `type Module interface {
    Definition() ModuleDef
  }`

#### Semantics

- Module names must be unique in a graph (enforced by kernel).
- Providers are resolved by token in the container.
- Controllers are instantiated after providers.
- Exports define which tokens are visible to importers.

### 5.2 `modkit/kernel`

**Responsibility:** Build module graph, enforce visibility, and bootstrap application.

#### Types

- `type App struct {
    Graph *Graph
    Container *Container
    Controllers map[string]any
  }`

- `func Bootstrap(root module.Module) (*App, error)`

#### Semantics

- Validates module graph:
  - no cycles
  - no duplicate module names
- Builds visibility map per module.
- Builds container of provider factories.
- Instantiates controllers once (singletons).

### 5.3 `modkit/http`

**Responsibility:** Adapt controllers to HTTP routing.

#### Types

- `type Router interface {
    Handle(method string, pattern string, handler http.Handler)
  }`

- `func RegisterRoutes(router Router, controllers map[string]any) error`

#### Semantics

- Controller registration is explicit (no reflection).
- Controllers expose route registration functions.

---

## 6. Validation criteria (MVP)

### 6.1 Core unit tests

- Module graph builds as expected
- Visibility rules enforced across imports/exports
- Provider resolution handles missing tokens with clear errors

### 6.2 Example app validation

- `examples/hello-mysql` compiles and runs
- HTTP routes wired and return expected responses
- MySQL integration verified with sqlc-generated code

---

## 7. Roadmap (Post-MVP)

- Optional CLI for module scaffolding
- Optional gRPC adapter
- Optional job runner/cron integration
- Optional config system
