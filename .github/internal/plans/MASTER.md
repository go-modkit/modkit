# modkit Documentation & Feature Plan

## Overview

This document tracks all planned improvements to modkit, organized by type and priority.
Each item links to its detailed implementation plan.

## Status Legend

- 🔴 Not started
- 🟡 In progress
- 🟢 Complete
- ⏭️ Deferred (post-MVP)

---

## Code Changes

| # | Topic | Status | Plan | Priority | Summary |
|---|-------|--------|------|----------|---------|
| C1 | Controller Registry Scoping | 🔴 | [code/01-controller-registry-scoping.md](code/01-controller-registry-scoping.md) | Medium | Namespace controller keys to allow same name across modules |
| C2 | Router Group and Use | 🔴 | [code/02-router-group-use.md](code/02-router-group-use.md) | High | Add `Group()` and `Use()` to Router interface (docs describe but not implemented) |
| C3 | App Container Access | 🔴 | [code/03-app-container-access.md](code/03-app-container-access.md) | Medium | Fix docs to use `App.Get()` instead of unexported `app.Container` |
| C4 | Logger Interface Alignment | 🔴 | [code/04-logger-interface-alignment.md](code/04-logger-interface-alignment.md) | Medium | Align Logger interface with docs: `...any` args, add `Warn` method |
| C5 | NewSlogLogger Rename | 🔴 | [code/05-newsloglogger-rename.md](code/05-newsloglogger-rename.md) | Low | Rename `NewSlog` → `NewSlogLogger` to match docs |
| C6 | Graceful Shutdown | 🔴 | [code/06-graceful-shutdown.md](code/06-graceful-shutdown.md) | Medium | Implement SIGINT/SIGTERM handling in `Serve()` (docs claim but not implemented) |

---

## SDLC / Tooling

| # | Topic | Status | Plan | Summary |
|---|-------|--------|------|---------|
| S1 | Commit Validation | 🟢 | [sdlc/01-commit-validation.md](sdlc/01-commit-validation.md) | Lefthook + Go commitlint |
| S2 | Changelog Automation | ⏭️ | — | Auto-generate CHANGELOG.md from commits |
| S3 | Release Workflow | ⏭️ | — | GitHub Actions with goreleaser |
| S4 | Pre-commit Hooks | ⏭️ | — | Run fmt/lint before commit |
| S5 | Test Coverage | ⏭️ | — | Coverage reporting in CI |

---

## Documentation Improvements

Ordered by logical implementation sequence. Complete earlier items before later ones.

| # | Topic | Status | Plan | NestJS Equivalent | Approach |
|---|-------|--------|------|-------------------|----------|
| D1 | Introduction & Overview | 🟢 | [docs/01-intro-overview.md](docs/01-intro-overview.md) | Introduction, Overview, First Steps | Add "Why modkit", architecture flow, bootstrap snippet |
| D2 | Modules | 🟢 | [docs/02-modules.md](docs/02-modules.md) | Modules | Clarify pointer identity, Definition() purity |
| D3 | Providers | 🟢 | [docs/03-providers.md](docs/03-providers.md) | Providers | Document lazy singleton lifecycle, cycle errors |
| D4 | Controllers | 🟢 | [docs/04-controllers.md](docs/04-controllers.md) | Controllers | Document RouteRegistrar contract |
| D5 | Middleware | 🟢 | [docs/05-middleware.md](docs/05-middleware.md) | Middleware | New guide: Go http.Handler patterns |
| D6 | Error Handling | 🟢 | [docs/06-error-handling.md](docs/06-error-handling.md) | Exception Filters | New guide: handler errors + middleware |
| D7 | Validation | 🟢 | [docs/07-validation.md](docs/07-validation.md) | Pipes | New guide: explicit decode/validate |
| D8 | Auth & Guards | 🟢 | [docs/08-auth-guards.md](docs/08-auth-guards.md) | Guards | New guide: auth middleware + context |
| D9 | Interceptors | 🔴 | [docs/09-interceptors.md](docs/09-interceptors.md) | Interceptors | New guide: middleware wrappers |
| D10 | Context Helpers | 🔴 | [docs/10-context-helpers.md](docs/10-context-helpers.md) | Custom Decorators | New guide: typed context keys |

---

## Post-MVP Roadmap

| Topic | Plan | Summary |
|-------|------|---------|
| modkitx | [docs/99-post-mvp-roadmap.md](docs/99-post-mvp-roadmap.md) | Optional ergonomics layer (builders, helpers) |
| modkit-cli | [docs/99-post-mvp-roadmap.md](docs/99-post-mvp-roadmap.md) | Scaffolding tool (not runtime) |
| gRPC Adapter | [docs/99-post-mvp-roadmap.md](docs/99-post-mvp-roadmap.md) | Future adapter package |

---

## NestJS Topics Intentionally Not Covered

These are handled by Go-idiomatic patterns documented in guides, not framework abstractions:

| NestJS Topic | modkit Approach | See Guide |
|--------------|-----------------|-----------|
| Exception Filters | Return errors + error middleware | D6: Error Handling |
| Pipes | Explicit json.Decode + validate | D7: Validation |
| Guards | Auth middleware | D8: Auth & Guards |
| Interceptors | Middleware wrappers | D9: Interceptors |
| Custom Decorators | Context helpers | D10: Context Helpers |
| Global Modules | Not supported (breaks explicit visibility) | — |
| Dynamic Modules | Options pattern in constructors | D2: Modules |

---

## Implementation Notes

1. **D1-D8 complete** — Existing guides in `docs/guides/` cover all planned content
2. **D9-D10** — Each guide is standalone; can be done in any order
3. **Testing** — Each guide should reference examples from `examples/hello-mysql`

### Code Change Priority Order

1. **C2 (Router Group/Use)** — High priority; multiple guides depend on this API
2. **C3 (Container Access)** — Docs fix only; quick win
3. **C4 (Logger Interface)** — Aligns implementation with documented API
4. **C6 (Graceful Shutdown)** — Docs promise feature that doesn't exist
5. **C1 (Controller Scoping)** — Improves multi-module apps
6. **C5 (NewSlogLogger)** — Simple rename; low priority
