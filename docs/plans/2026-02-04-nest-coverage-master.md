# Nest Topics Coverage Master Plan

**Goal:** Track coverage of key NestJS documentation topics in modkit. For each topic, record whether it is implemented, planned (documented in a plan), or not covered. This is a maintainer-facing inventory, not a roadmap.

**Scope:** The topics listed mirror the NestJS left‑nav “Introduction” section: Introduction, Overview, First steps, Controllers, Providers, Modules, Middleware, Exception filters, Pipes, Guards, Interceptors, Custom decorators.

**Legend:**
- **Implemented** = framework behavior exists and is documented in repo.
- **Planned** = there is an existing plan doc to add or clarify coverage.
- **Not covered** = no framework support and no plan.

| Topic | Status | Local References | Differences vs Nest | Recommendation |
| --- | --- | --- | --- | --- |
| Introduction | Implemented | `README.md`, `docs/design/mvp.md` | No CLI or decorators; explicit module metadata. | Add “Why modkit” + no‑reflection callout (plan: `docs/plans/2026-02-04-intro-overview-first-steps-docs.md`). |
| Overview | Implemented | `README.md`, `docs/design/mvp.md` | Kernel graph + visibility instead of metadata scanning. | Add architecture flow/diagram (plan: `docs/plans/2026-02-04-intro-overview-first-steps-docs.md`). |
| First steps | Implemented | `docs/guides/getting-started.md`, `examples/hello-mysql/README.md` | Bootstrap via `kernel.Bootstrap`; no CLI scaffold. | Add minimal `main.go` snippet (plan: `docs/plans/2026-02-04-intro-overview-first-steps-docs.md`). |
| Controllers | Implemented + Planned docs | `modkit/http/router.go`, `modkit/http/doc.go`, `docs/design/http-adapter.md`, `docs/guides/getting-started.md`, `docs/plans/2026-02-04-controller-contract-docs.md` | Explicit `RouteRegistrar` interface; no decorators. | Execute controller contract doc plan. |
| Providers | Implemented | `modkit/module/provider.go`, `modkit/kernel/container.go`, `docs/guides/modules.md`, `docs/design/mvp.md` | String tokens; lazy singleton build; explicit `Get`. | Add provider lifecycle note (plan: `docs/plans/2026-02-04-providers-docs.md`). |
| Modules | Implemented + Planned docs | `modkit/module/module.go`, `modkit/kernel/graph.go`, `modkit/kernel/visibility.go`, `docs/guides/modules.md`, `docs/design/mvp.md`, `docs/plans/2026-02-04-document-definition-purity.md`, `docs/plans/2026-02-04-module-identity-clarification.md` | Pointer identity; no dynamic module metadata. | Execute existing module doc plans. |
| Middleware | Implemented (minimal) | `modkit/http/router.go`, `modkit/http/logging.go`, `docs/design/http-adapter.md` | Go `http.Handler` chain via `chi`. | Add middleware guide (plan: `docs/plans/2026-02-04-middleware-guide.md`). |
| Exception filters | Not covered | — | Nest filters are framework hooks; Go handles errors in handlers/middleware. | Do not add to core; add error‑handling guide (plan: `docs/plans/2026-02-04-error-handling-guide.md`). |
| Pipes | Not covered | — | Nest pipes do validation/transform; Go uses explicit decode/validate. | Do not add to core; add validation guide (plan: `docs/plans/2026-02-04-validation-guide.md`). |
| Guards | Not covered | — | Nest guards run pre‑handler; Go uses auth middleware. | Do not add to core; add auth/guards guide (plan: `docs/plans/2026-02-04-auth-guards-guide.md`). |
| Interceptors | Not covered | — | Nest interceptors wrap execution; Go uses middleware/wrappers. | Do not add to core; add interceptors guide (plan: `docs/plans/2026-02-04-interceptors-guide.md`). |
| Custom decorators | Not covered | — | Nest decorators rely on metadata; Go uses context helpers. | Do not add to core; add context helpers guide (plan: `docs/plans/2026-02-04-context-helpers-guide.md`). |

**Notes**
- “Implemented” means the framework (not just example app) supports the concept and docs exist. Example‑only patterns are not counted as coverage.
- “Planned” indicates a doc improvement plan exists; no behavioral change is implied unless explicitly stated in a plan.
