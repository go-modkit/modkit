# Product Requirement Document: modkit Core Framework

**Status:** Active
**Version:** 0.1.0
**Date:** 2026-02-07
**Author:** Sisyphus (AI Agent)
**Last Reviewed:** 2026-02-08

---

## 1. Executive Summary

**modkit** is a Go framework designed to solve the "Big Ball of Mud" problem in long-lived, complex backend applications. It introduces **strict modular boundaries** and **explicit dependency injection** to the Go ecosystem, inspired by the structural discipline of NestJS but implemented with Go's philosophy of "no magic."

Unlike existing libraries that focus solely on *wiring* (Fx, Wire, Dig), modkit focuses on *architecture*. It enforces visibility rules between modules, ensuring that as a codebase grows, its dependency graph remains acyclic, understandable, and maintainable.

## 2. Problem Statement

### The "Go Monolith" Trajectory
1.  **Start:** A simple `main.go` and a `handler` package. Clean and fast.
2.  **Growth:** Handlers multiply. Business logic moves to `service` packages. Database code moves to `repo`.
3.  **Complexity:** Dependencies are passed manually in `main()`. The `main` function becomes 500+ lines of wiring code.
4.  **Decay:** To avoid import cycles, developers create a `shared` or `common` package. Everything depends on everything. The graph becomes a tangled web. Refactoring is risky.
5.  **Result:** A fragile monolith where changing one service breaks unrelated features.

### Existing Solutions & Gaps
*   **Manual Wiring:** Great for small apps, unmaintainable for large teams.
*   **Google Wire:** excellent compile-time safety, but lacks runtime module concepts or visibility enforcement.
*   **Uber Fx:** Powerful runtime DI, but defaults to a flat global namespace. It solves wiring, not architecture.
*   **NestJS (Node):** Solves the architecture problem perfectly but relies on heavy reflection and decorators, which are non-idiomatic and slow in Go.

## 3. Goals & Non-Goals

### Core Goals
*   **Enforce Boundaries:** Make "private by default" the standard. Module A cannot use Module B's internals unless explicitly exported.
*   **Zero Magic:** No runtime reflection scanning, no struct tag magic, no global state. All wiring is code.
*   **Deterministic Bootstrap:** The application starts the same way every time. The dependency graph is computed before instantiation.
*   **Go Idiomatic:** Use standard Go patterns (structs, interfaces, simple functions) instead of imitating OOP classes or decorators.
*   **Observability:** The framework should know *exactly* what the dependency graph looks like, enabling visualization and debugging.

### Non-Goals
*   **Replacing the Stdlib:** We wrap `net/http` (via Chi) but do not replace standard types.
*   **ORM Integration:** We are agnostic to data access (Gorm, Sqlc, Ent, raw SQL).
*   **Code Generation:** While we might add CLI helpers later, the core framework must work with plain Go code.

## 4. Technical Architecture

### 4.1. The Module System
The fundamental unit of architecture is the `Module`. A module is a struct that implements `Definition()`.

```go
type Definition struct {
    Name        string
    Imports     []Module
    Providers   []ProviderDef
    Controllers []ControllerDef
    Exports     []Token
}
```

*   **Encapsulation:** Providers defined in a module are **private** by default.
*   **Explicit Imports:** A module must explicitly import another module to use its exported providers.
*   **Explicit Exports:** A module must explicitly export a provider token to make it available to importers.

### 4.2. The Kernel & Dependency Injection
The `Kernel` is the engine that processes the module graph.

*   **Graph Builder:** Recursively walks the module tree, detecting cycles and validating metadata.
*   **Container:** A lazy-loading, singleton-based DI container.
*   **Resolvers:** Each module gets a scoped `Resolver`. When Module A asks for `TokenX`, the kernel checks:
    1.  Is `TokenX` defined internally in A? -> Return it.
    2.  Is `TokenX` exported by an imported module? -> Return it.
    3.  Else -> Error (Visibility Violation).

### 4.3. HTTP Layer
*   **Router Agnostic (Conceptually):** Currently implemented on `go-chi` for speed and standard compliance.
*   **Controllers:** Simple structs that implement `RegisterRoutes(r Router)`.
*   **Registration:** Routes are registered explicitly, not via struct tags.

## 5. Developer Experience (DX)

### 5.1. Creating a Feature
To add a "Users" feature, a developer:
1.  Creates a `users` package.
2.  Defines a `UsersModule` struct.
3.  Implements `UsersService` (business logic) and `UsersController` (HTTP).
4.  Wires them in `Definition()`:
    *   `Providers`: `Use(NewUsersService)`
    *   `Controllers`: `Use(NewUsersController)`
    *   `Exports`: `Export("users.service")` if other modules need it.

### 5.2. Error Handling
Errors should be:
*   **Compile-time where possible:** (e.g., using type-safe constructors).
*   **Initialization-time:** The app crashes immediately on boot if dependencies are missing or cycles exist.
*   **Descriptive:** "Module 'Orders' cannot import 'Users' because 'Users' is not in the imports list."

## 6. Comparison Matrix

| Feature | modkit | Uber Fx | Google Wire | Manual |
| :--- | :--- | :--- | :--- | :--- |
| **Architecture** | Modular (NestJS-style) | Flat / Grouped | Graph-based | Ad-hoc |
| **Visibility** | **Strict (Private/Public)** | None (Global) | None | N/A |
| **Wiring** | Explicit Config | Reflection / Global | Code Gen | Manual |
| **Performance** | Low overhead | Reflection hit | Zero overhead | Zero overhead |
| **Safety** | Runtime (Boot) | Runtime (Boot) | Compile-time | Compile-time |

## 7. Roadmap

### Phase 1: Core Stability (Current)
*   [x] Module definition & graph resolution
*   [x] Visibility enforcement
*   [x] Basic HTTP adapter (Chi)
*   [x] Graceful shutdown APIs (`App.Close` / `App.CloseContext`) with closer ordering and error aggregation.
*   [x] Generic Helpers: Reduce type-casting noise with `module.Get[T]` helper.

### Phase 2: Ecosystem & Tooling (Next)
*   [x] **modkit-cli**: Scaffold new projects, modules, and providers (`modkit new module users`). Delivered with auto-registration + CI smoke checks + release artifacts.
*   [ ] **TestKit**: Utilities for testing modules in isolation (mocking providers easily).
*   [x] **Config Module**: Standard pattern for loading env vars into the container. See `modkit/config` and `docs/specs/design-config-module.md`.

### Phase 3: Advanced Features
*   [ ] **Graph Visualization**: Dump the dependency graph to Mermaid/Dot format.
*   [ ] **Devtools**: Decision pending (currently treated as de-scoped in Nest compatibility docs).

### Remaining Work (Prioritized)

1. **P0 - TestKit (Phase 2)**
   - Deliver a focused test utility package for module-isolation tests with provider override/mocking support.
2. **P1 - Spec/roadmap synchronization**
   - Reconcile epic/spec checklists and statuses with shipped code so planning docs match reality.
3. **P2 - Graph visualization (Phase 3)**
   - Provide graph export output (Mermaid/DOT) for architecture introspection.
4. **P2 - Devtools direction decision (Phase 3)**
   - Either define a minimal built-in endpoint scope or formally de-scope from PRD to match current guidance.

### Synchronization Summary (2026-02-08)

P1 roadmap/spec synchronization reconciled shipped-state mismatches across roadmap docs and guide surfaces:

1. `docs/specs/epic-01-examples-enhancement.md`: checklist state updated to reflect delivered hello-mysql capabilities; unresolved route-group follow-ups kept open.
2. `docs/specs/epic-02-core-nest-compatibility.md`: graceful shutdown, re-export, and compatibility-doc acceptance criteria reconciled to implemented state with evidence-backed checks.
3. `docs/guides/nestjs-compatibility.md`: matrix wording aligned (CLI = implemented, lifecycle/re-export = implemented, devtools = decision pending).
4. `docs/specs/design-release-versioning-sdlc-cli.md`: moved from draft intent to implemented-state checklist for released pipeline artifacts.
5. `README.md`: guide index reconciled with shipped docs by adding the configuration guide link.

## 8. Success Metrics

1.  **Adoption**: Used in at least 3 production services within the organization.
2.  **Maintenance**: New developers can locate feature logic in <5 minutes without asking for help.
3.  **Stability**: Zero "import cycle" refactors required after adopting modkit.
4.  **Performance**: Boot time overhead <50ms for a graph of 100+ modules.
