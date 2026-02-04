# Nest Topics Coverage Master Plan

**Goal:** Track coverage of key NestJS documentation topics in modkit. For each topic, record whether it is implemented, planned (documented in a plan), or not covered. This is a maintainer-facing inventory, not a roadmap.

**Scope:** The topics listed mirror the NestJS left‑nav “Introduction” section: Introduction, Overview, First steps, Controllers, Providers, Modules, Middleware, Exception filters, Pipes, Guards, Interceptors, Custom decorators.

**Legend:**
- **Implemented** = framework behavior exists and is documented in repo.
- **Planned** = there is an existing plan doc to add or clarify coverage.
- **Not covered** = no framework support and no plan.

| Topic | Status | Local References | Notes (Go‑idiomatic mapping) |
| --- | --- | --- | --- |
| Introduction | Implemented | `README.md`, `docs/design/mvp.md` | Defines modkit purpose and constraints; emphasizes explicit wiring. |
| Overview | Implemented | `README.md`, `docs/design/mvp.md` | Architecture overview: module → kernel → http adapter. |
| First steps | Implemented | `docs/guides/getting-started.md`, `examples/hello-mysql/README.md` | Bootstrap and minimal HTTP serve path. |
| Controllers | Implemented + Planned docs | `modkit/http/router.go`, `modkit/http/doc.go`, `docs/design/http-adapter.md`, `docs/guides/getting-started.md`, `docs/plans/2026-02-04-controller-contract-docs.md` | Go controllers register routes explicitly via interface; no decorators. |
| Providers | Implemented | `modkit/module/provider.go`, `modkit/kernel/container.go`, `docs/guides/modules.md`, `docs/design/mvp.md` | Singleton, lazy construction; explicit tokens. |
| Modules | Implemented + Planned docs | `modkit/module/module.go`, `modkit/kernel/graph.go`, `modkit/kernel/visibility.go`, `docs/guides/modules.md`, `docs/design/mvp.md`, `docs/plans/2026-02-04-document-definition-purity.md`, `docs/plans/2026-02-04-module-identity-clarification.md` | Imports/exports control visibility; pointers required for identity. |
| Middleware | Implemented (minimal) | `modkit/http/router.go`, `modkit/http/logging.go`, `docs/design/http-adapter.md` | Go middleware via `chi` and standard `http.Handler` chain. |
| Exception filters | Not covered | — | Use handler‑level errors and middleware patterns in Go; no framework feature. |
| Pipes | Not covered | — | Use explicit decode/validate steps in handlers. |
| Guards | Not covered | — | Use middleware or handler wrappers for auth. |
| Interceptors | Not covered | — | Use middleware for logging/metrics/response shaping. |
| Custom decorators | Not covered | — | Use typed context keys and helper functions. |

**Notes**
- “Implemented” means the framework (not just example app) supports the concept and docs exist. Example‑only patterns are not counted as coverage.
- “Planned” indicates a doc improvement plan exists; no behavioral change is implied unless explicitly stated in a plan.
