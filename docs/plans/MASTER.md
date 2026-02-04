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

| # | Topic | Status | Plan | Summary |
|---|-------|--------|------|---------|
| C1 | Controller Registry Scoping | 🔴 | [code/01-controller-registry-scoping.md](code/01-controller-registry-scoping.md) | Namespace controller keys to allow same name across modules |

---

## SDLC / Tooling

| # | Topic | Status | Plan | Summary |
|---|-------|--------|------|---------|
| S1 | Commit Validation | 🔴 | [sdlc/01-commit-validation.md](sdlc/01-commit-validation.md) | Lefthook + Go commitlint |
| S2 | Changelog Automation | ⏭️ | — | Auto-generate CHANGELOG.md from commits |
| S3 | Release Workflow | ⏭️ | — | GitHub Actions with goreleaser |
| S4 | Pre-commit Hooks | ⏭️ | — | Run fmt/lint before commit |
| S5 | Test Coverage | ⏭️ | — | Coverage reporting in CI |

---

## Documentation Improvements

Ordered by logical implementation sequence. Complete earlier items before later ones.

| # | Topic | Status | Plan | NestJS Equivalent | Approach |
|---|-------|--------|------|-------------------|----------|
| D1 | Introduction & Overview | 🔴 | [docs/01-intro-overview.md](docs/01-intro-overview.md) | Introduction, Overview, First Steps | Add "Why modkit", architecture flow, bootstrap snippet |
| D2 | Modules | 🔴 | [docs/02-modules.md](docs/02-modules.md) | Modules | Clarify pointer identity, Definition() purity |
| D3 | Providers | 🔴 | [docs/03-providers.md](docs/03-providers.md) | Providers | Document lazy singleton lifecycle, cycle errors |
| D4 | Controllers | 🔴 | [docs/04-controllers.md](docs/04-controllers.md) | Controllers | Document RouteRegistrar contract |
| D5 | Middleware | 🔴 | [docs/05-middleware.md](docs/05-middleware.md) | Middleware | New guide: Go http.Handler patterns |
| D6 | Error Handling | 🔴 | [docs/06-error-handling.md](docs/06-error-handling.md) | Exception Filters | New guide: handler errors + middleware |
| D7 | Validation | 🔴 | [docs/07-validation.md](docs/07-validation.md) | Pipes | New guide: explicit decode/validate |
| D8 | Auth & Guards | 🔴 | [docs/08-auth-guards.md](docs/08-auth-guards.md) | Guards | New guide: auth middleware + context |
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

1. **Order matters** — D1-D4 document existing behavior; complete before D5-D10 (new guides)
2. **Code change C1** — Can be done independently; improves multi-module apps
3. **D5-D10** — Each guide is standalone; can be done in any order after D1-D4
4. **Testing** — Each guide should reference examples from `examples/hello-mysql`
