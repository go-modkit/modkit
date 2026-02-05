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
| C1 | Controller Registry Scoping | 🟢 | [code/01-controller-registry-scoping.md](code/01-controller-registry-scoping.md) | Medium | Namespace controller keys to allow same name across modules |
| C2 | Router Group and Use | 🟢 | [code/02-router-group-use.md](code/02-router-group-use.md) | High | Add `Group()` and `Use()` to Router interface (docs describe but not implemented) |
| C3 | App Container Access | 🟢 | [code/03-app-container-access.md](code/03-app-container-access.md) | Medium | Fix docs to use `App.Get()` instead of unexported `app.Container` |
| C4 | Logger Interface Alignment | 🟢 | [code/04-logger-interface-alignment.md](code/04-logger-interface-alignment.md) | Medium | Align Logger interface with docs: `...any` args, add `Warn` method |
| C5 | NewSlogLogger Rename | 🟢 | [code/05-newsloglogger-rename.md](code/05-newsloglogger-rename.md) | Low | Rename `NewSlog` → `NewSlogLogger` to match docs |
| C6 | Graceful Shutdown | 🟢 | [code/06-graceful-shutdown.md](code/06-graceful-shutdown.md) | Medium | Implement SIGINT/SIGTERM handling in `Serve()` (docs claim but not implemented) |

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
| D9 | Interceptors | 🟢 | [docs/09-interceptors.md](docs/09-interceptors.md) | Interceptors | New guide: middleware wrappers |
| D10 | Context Helpers | 🟢 | [docs/10-context-helpers.md](docs/10-context-helpers.md) | Custom Decorators | New guide: typed context keys |

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

1. **D1-D10 complete** — All documentation guides in `docs/guides/` are complete
2. **C1-C6 complete** — All code changes implemented and merged
3. **Testing** — Each guide references examples from `examples/hello-mysql`

---

## Dependency Analysis

### Story Dependencies

| Story | Can Start Immediately | Blocks | Blocked By |
|-------|----------------------|--------|------------|
| C1 | ✅ | — | — |
| C2 | ✅ | D9 | — |
| C3 | ✅ | — | — |
| C4 | ✅ | C5 | — |
| C5 | ❌ | — | C4 |
| C6 | ✅ | — | — |
| D9 | ❌ | — | C2 |
| D10 | ✅ | — | — |

### Sequential Dependencies

```text
C4 (Logger Interface) ──► C5 (NewSlogLogger Rename)
        │
        └── Both modify modkit/logging/slog.go and logger_test.go
            C5 must be done AFTER C4 to avoid conflicts

C2 (Router Group/Use) ──► D9 (Interceptors Guide)
        │
        └── D9 documents middleware wrapper patterns using
            Group() and Use() methods from C2
```

### Parallel Work Groups

**Group A: Kernel** — isolated from http/logging
- C1: Controller Registry Scoping

**Group B: HTTP** — isolated from kernel/logging
- C2: Router Group and Use
- C6: Graceful Shutdown *(different files)*

**Group C: Logging** — sequential internally
- C4: Logger Interface Alignment *(first)*
- C5: NewSlogLogger Rename *(after C4)*

**Group D: Docs Only** — no code dependencies
- C3: App Container Access
- D10: Context Helpers

**Group E: Depends on C2**
- D9: Interceptors

---

## Execution Plan

### Optimal Agent Assignment (4 concurrent agents)

Sequential work stays within the same agent context — no cross-agent waiting required.

| Agent | Work | Notes |
|-------|------|-------|
| Agent 1 | C1, C3 | Independent stories, parallel within agent |
| Agent 2 | C2 → D9 | D9 continues immediately after C2 |
| Agent 3 | C4 → C5 | C5 continues immediately after C4; same files |
| Agent 4 | C6, D10 | Independent stories, parallel within agent |

```text
Agent 1: ─── C1 ───┬─── C3 ───
                   │
Agent 2: ─── C2 ───────────────► D9 ───
                   │
Agent 3: ─── C4 ───────────────► C5 ───
                   │
Agent 4: ─── C6 ───┬─── D10 ───
```

**Why this works:**
- Each agent runs to completion without waiting for other agents
- Sequential dependencies (C4→C5, C2→D9) stay within same agent context
- Agent has full knowledge of prior changes when continuing to dependent work
- Same files stay together (C4/C5 both touch `logging/slog.go`)

### Priority Order (single agent)

1. **C2 (Router Group/Use)** — High priority; unblocks D9
2. **C4 (Logger Interface)** — Unblocks C5
3. **C3 (Container Access)** — Docs fix only; quick win
4. **C6 (Graceful Shutdown)** — Docs promise feature that doesn't exist
5. **C1 (Controller Scoping)** — Improves multi-module apps
6. **D10 (Context Helpers)** — Standalone doc
7. **C5 (NewSlogLogger)** — Simple rename; requires C4
8. **D9 (Interceptors)** — Requires C2
