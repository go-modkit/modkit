# Post-MVP Roadmap

**Status:** ⏭️ Deferred  
**Type:** Roadmap / Future planning

---

## Purpose

This document defines the post-MVP direction for modkit with a clear architectural north star. The core will remain Go-idiomatic and minimal (Option 1). Ergonomics and scaffolding will live in separate, optional packages (`modkitx` and `modkit-cli`). Implementation scheduling is intentionally deferred until the MVP documentation is complete.

---

## Architectural North Star: Go-Idiomatic Minimal Core

The core package (`modkit`) is responsible for only the essential, stable primitives:
- Module metadata (`imports`, `providers`, `controllers`, `exports`)
- Deterministic graph construction with visibility enforcement
- Singleton provider container with explicit resolution
- Explicit controller registration without reflection

The core must avoid runtime magic, decorators, and hidden conventions. It should be small, predictable, and debuggable. API changes should be conservative and motivated by clarity or correctness, not convenience.

### In Scope for Core

- Module model and validation
- Kernel graph, visibility, container, and bootstrap
- Minimal HTTP adapter (routing + server glue)

### Out of Scope for Core

- Scaffolding or generation tooling
- Opinionated defaults (config, logging, metrics)
- Additional adapters (gRPC, jobs)
- Helper DSLs or builders

---

## Companion Package: modkitx (Ergonomics Layer)

`modkitx` is an optional helper package that reduces boilerplate without changing semantics. It should only wrap or generate the same metadata that core consumes. It must not introduce reflection, auto-wiring, or hidden registration.

### Goals

- Provide a small builder API that compiles to `module.Module` and produces `ModuleDef`
- Add common provider helpers (e.g., value provider, func provider)
- Provide optional HTTP helper middleware (logging, request IDs, standardized error helpers)

### Constraints

- All helpers must be explicit and deterministic
- The kernel behavior must remain unchanged; only ergonomics improve
- `modkitx` should be fully optional for adoption

### Quality Bar

- Unit tests that assert the builder output equals a hand-written `ModuleDef`
- Middleware helpers are opt-in and do not change behavior unless attached

---

## Companion Package: modkit-cli (Scaffolding Tooling)

`modkit-cli` is a developer productivity tool for generating skeletons and boilerplate. It must never be required at runtime. Generated code should compile cleanly and use only public APIs from `modkit` (and optionally `modkitx`).

### Goals

- Generate module scaffolds (`module.go`, `controller.go`, `routes.go`, `service.go`)
- Generate a minimal app bootstrap with an HTTP server
- Produce code that follows modkit conventions by default but is fully editable

### Constraints

- No runtime dependency on the CLI after generation
- Templates should be simple and stable before adding complex project inspection

### Quality Bar

- Golden-file tests for templates
- Generated code compiles and passes a minimal test suite

---

## Future Adapters

### gRPC Adapter

Add `modkit/grpc` as a thin adapter similar to `modkit/http`:
- Controllers implement a gRPC registration interface
- No reflection-based service discovery
- Works with protoc-generated code

### Job Runner

Add `modkit/jobs` for background job processing:
- Jobs as providers with explicit registration
- Simple in-process runner for MVP
- Optional external queue integration later

---

## Sequencing Notes

Implementation scheduling will be defined after the MVP documentation is complete. This document only fixes direction and package boundaries, not timeline or task breakdown.

Priority order (when ready):
1. Complete MVP documentation (D1-D10)
2. modkitx builder API
3. modkit-cli scaffolding
4. gRPC adapter
5. Job runner
