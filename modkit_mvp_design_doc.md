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

- `type Module interface { Definition() ModuleDef }`

#### Errors
- `ErrInvalidModuleDef` (missing name, invalid exports, etc.)

#### MVP rules enforced by kernel but described here
- `Exports` must refer to tokens provided by:
  - this module’s `Providers`, OR
  - imported module’s `Exports` (re-export allowed)

---

### 5.2 `modkit/kernel`

**Responsibility:** Build module graph, enforce import/export visibility, provide DI container (singleton), bootstrap application.

#### Types

- `type App struct {
    Root module.Module
    Modules []module.Module // flattened order (imports first)
    Container module.Resolver
    Controllers []any
  }`

- `func Bootstrap(root module.Module) (*App, error)`

#### Container semantics
- Singleton instances per token
- Lazy build by default (instances created on first resolve), with optional eager init function for validation

#### Graph semantics
- Flattened module list must be deterministic and stable
- Cycle detection with module-name path in error
- Duplicate module names are errors in MVP (later could add unique ModuleID)

#### Visibility enforcement
For any provider/controller factory running within module `M`:
- Allowed tokens are:
  - tokens provided by `M`
  - tokens exported by `M`’s direct imports, plus recursively by their imports (Nest-like)
- Attempts to resolve other tokens should yield `ErrNotVisible(token, fromModule)`

Implementation note (MVP): Enforce by creating a **module-scoped resolver** passed to provider/controller build functions, wrapping the global container.

---

### 5.3 `modkit/http`

**Responsibility:** Provide HTTP server bootstrap utilities and conventions. Does not define controllers by reflection.

#### Controller integration model
Controllers are plain Go structs with handler methods; modules expose route registration functions.

**Pattern:**
- Each feature module includes `routes.go` with `func RegisterRoutes(r chi.Router, c module.Resolver) error`.
- `RegisterRoutes` resolves controller token(s) and binds them to routes.

modkit/http will provide:
- `func NewRouter() chi.Router` with baseline middleware
- `func Serve(addr string, r http.Handler) error`
- Optional middleware helpers: request ID, logging hooks, recovery, error mapping

---

## 6. Module responsibilities & relations (within consuming app)

This section defines the *recommended* architecture in consuming apps.

### 6.1 AppModule (root)
**Responsibilities**
- Own configuration (env parsing)
- Compose imports of all feature and platform modules
- Provide global providers (logger, config)

**Exports**
- Usually none (root)

### 6.2 Platform modules

#### DatabaseModule
- Provides `*sql.DB`
- Exports DB token

#### ConfigModule
- Provides strongly typed config struct
- Exports config token

#### LoggingModule
- Provides logger instance
- Exports logger token

### 6.3 Feature modules (e.g., UsersModule)

**Responsibilities**
- Define domain/application services
- Provide repositories via adapters (MySQL repo provider depends on DB token)
- Provide controllers
- Export services intended for other modules

**Imports**
- DatabaseModule
- LoggingModule

**Exports**
- `UsersService` token (optional)

### 6.4 Relations summary
- AppModule imports platform + feature modules
- Feature modules import platform modules
- Feature modules can import other feature modules only via exported services (not internal repos)

---

## 7. Dynamic module configuration

### 7.1 Requirements
- Modules must be constructible with options (like Nest “dynamic modules”)
- Options affect:
  - which modules are imported
  - which providers are registered
  - which exports are available

### 7.2 Pattern

```
func NewUsersModule(opts UsersOptions) module.Module
```

Options must be immutable after creation.

### 7.3 Validation
Modules must validate options in `Definition()` and return a `ModuleDef` that will pass kernel validation.

---

## 8. Validation & testing strategy

### 8.1 Kernel unit tests (required)

1) **Graph order**
- Given imports A→B→C, flattened list is [C, B, A]

2) **Cycle detection**
- A imports B, B imports A → error includes path

3) **Duplicate module names**
- Two modules with Name=“users” → error

4) **Duplicate provider tokens**
- Two providers with same token in visible graph → error

5) **Visibility enforcement**
- Module A does not import module B → resolving B export from A provider must error

6) **Re-export support**
- A imports B and exports B’s token → C imports A → C can resolve that token

7) **Bootstrap controller instantiation**
- Controller build runs, can resolve allowed providers

### 8.2 Example app tests (required)

- `go test ./...` in `examples/hello-mysql`
- Integration smoke test:
  - bring up MySQL (local docker or testcontainers)
  - run migrations
  - start server
  - call `/health` and one CRUD endpoint

---

## 9. Implementation phases

### Phase 0 — Repo bootstrap (day 0)
**Deliverables**
- Repo initialized, CI runs `go test ./...`
- `docs/design/mvp.md` present
- Public package scaffolding

### Phase 1 — module package
**Deliverables**
- `modkit/module` types + docs
- Minimal compile-only tests

### Phase 2 — kernel graph + container
**Deliverables**
- `kernel.Bootstrap`
- Graph flattening + validation
- Module-scoped resolver enforcing visibility
- Full kernel unit test suite (Section 8.1)

### Phase 3 — HTTP adapter (minimal)
**Deliverables**
- `modkit/http` router + server helper
- Convention docs for `routes.go` registration

### Phase 4 — consuming app (hello-mysql)
**Deliverables**
- AppModule + DatabaseModule + UsersModule
- One working endpoint group (GET/POST)
- sqlc + migrations included
- Smoke test script / Makefile targets

### Phase 5 — Documentation completeness
**Deliverables**
- Getting started guide
- Modules guide referencing Nest module concepts
- Testing guide

---

## 10. Final output checklist (verification criteria)

This is the MOST IMPORTANT section. The MVP is complete only if all items below are true.

### 10.1 Code outputs

1) **Bootstrap works**
- A user can call `kernel.Bootstrap(app.NewModule(opts))` and get an `*kernel.App` without panics.

2) **Visibility is enforced**
- Providers/controllers cannot resolve tokens from modules that are not imported (unless re-exported)

3) **Deterministic module ordering**
- `App.Modules` order is stable and imports-first

4) **Controller instances can be resolved and mounted**
- Example app binds routes by resolving controller(s) via resolver

5) **MySQL example runs end-to-end**
- `make run` starts MySQL, applies migrations, starts server
- `make test` runs unit tests + integration smoke test

### 10.2 Documentation outputs

1) README contains:
- What modkit is
- Quickstart using the example
- Architecture overview (modules/providers/controllers/exports)
- Explicit “Inspired by NestJS modules” note with reference

2) Docs contain:
- `docs/design/mvp.md`
- `docs/guides/getting-started.md`
- `docs/guides/modules.md` mapping Nest concepts

### 10.3 CI outputs

- GitHub Actions workflow:
  - `go test ./...` for library
  - `go test ./...` for example
  - `golangci-lint` (optional but recommended)

### 10.4 Consumer validation

The consuming app must demonstrate:
- A feature module exporting a service
- Another module importing it and consuming via exported token (can be small “audit” module)
- Re-export case (optional but recommended)

---

## 11. Next steps after MVP (not implemented)

- Provider scopes (request/transient)
- Lifecycle hooks (`OnStart/OnStop`)
- CLI scaffolding (`modkit new`, `modkit gen module`)
- HTTP controller helpers (request decoding/encoding, validation)
- Observability defaults (metrics, tracing)
- OpenAPI generation / adapters

---

## Appendix A — NestJS docs references

- Modules concept, imports/providers/controllers/exports: NestJS Modules documentation.

(Agents implementing modkit should mirror the semantics, not the Nest implementation details.)

