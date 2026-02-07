# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-07T10:40:45Z
**Commit:** 85a5d8b
**Branch:** chore/init-deep-agents

## OVERVIEW
modkit is a Go framework for modular backend services (NestJS-inspired, Go-idiomatic): explicit boundaries, explicit DI tokens, deterministic bootstrap, no runtime magic.

## STRUCTURE
```text
modkit/
|- modkit/                 # Core framework packages
|  |- module/              # Public module metadata API
|  |- kernel/              # Graph, visibility, container, bootstrap
|  |- http/                # Router/server/middleware adapter (chi)
|  `- logging/             # Logging contract + adapters
|- examples/               # Separate Go modules using modkit
|  |- hello-simple/        # Smallest possible end-to-end app
|  `- hello-mysql/         # Full app: auth/users/db/migrations/sqlc/swagger
`- docs/                   # Guides, reference, architecture, specs
```

## WHERE TO LOOK
| Task | Location | Notes |
|---|---|---|
| Module metadata types | `modkit/module` | `ModuleDef`, `ProviderDef`, tokens, validation |
| Bootstrap and DI behavior | `modkit/kernel` | Import graph, visibility rules, lazy singleton container |
| HTTP route registration | `modkit/http` | Controllers register via explicit router contract |
| Logging interfaces | `modkit/logging` | `slog` adapter + nop logger |
| Minimal runnable example | `examples/hello-simple/main.go` | Smallest canonical bootstrap flow |
| Production-like sample app | `examples/hello-mysql` | Separate module, compose, migrations, sqlc, swagger |
| User docs | `docs/guides` | How-to guidance by topic |

## CODE MAP
LSP unavailable here; use `rg` + tests as code-map proxies.

High-signal packages by density:
- `modkit/kernel` (14 files, mostly behavior tests): graph build, visibility, container lifecycle, typed errors.
- `modkit/http` (11 files): router/server/middleware/logging integration.
- `examples/hello-mysql/internal/modules/{auth,users}` (32 files): sample app domain patterns.

## CONVENTIONS
- Modules are pointers; names are unique; `Definition()` is deterministic.
- Provider tokens use `"module.component"`; providers are lazy singletons.
- Resolve dependencies via `r.Get()` and check `err` before type assertion.
- Controllers implement `RegisterRoutes(router Router)` and use explicit route methods.
- Error handling is explicit: return/wrap errors, no panic-driven control flow.
- Keep tests near implementation (`*_test.go`), table-driven where useful.

## ANTI-PATTERNS (THIS PROJECT)
- Do not introduce reflection/decorator magic in user-facing APIs.
- Do not make `Definition()` stateful (e.g., counters/mutable side effects).
- Do not bypass module visibility via cross-module direct wiring.
- Do not globally exclude tests from lint rules (`.golangci.yml` guards this).
- Do not hand-edit generated artifacts (`*.sql.go`, swagger-generated docs files).

## UNIQUE STYLES
- Multi-module workspace (`go.work`) includes root + both examples.
- Root `make test` iterates every `go.mod` module with `-race`.
- CI enforces Go `1.25.7` plus lint/vuln/coverage gates.
- Commit and PR titles follow Conventional Commits.

## PR WORKFLOW POLICY
- If work is driven by a GitHub issue hierarchy (parent/story + sub-issues), PR bodies must include one `Resolves #<number>` line for each implemented issue.
- If work is not issue-driven, omit `Resolves` (or mark N/A if the active PR template requires the section).
- If `.github/pull_request_template.md` exists, PR descriptions must follow that template structure.
- Template checklist must be actively reconciled (checked/unchecked/N/A with reason) before review request.
- A checklist box may be checked only when same-session verification evidence exists.
- Before PR create/update, run `make fmt && make lint && make test` (or documented project equivalent) and reflect outcomes in the checklist.
- If a check is not run, leave it unchecked or mark N/A with an explicit reason.
- After any new commit on the PR branch, rerun affected checks and re-reconcile the checklist.

## COMMANDS
```bash
make fmt
make lint
make vuln
make test
make test-coverage
```

## NOTES
- Local guidance: `modkit/AGENTS.md`, `examples/AGENTS.md`, `examples/hello-mysql/AGENTS.md`, `examples/hello-mysql/internal/modules/AGENTS.md`.
