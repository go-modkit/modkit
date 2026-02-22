# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make fmt              # gofmt + goimports
make lint             # golangci-lint
make vuln             # govulncheck
make test             # run all module tests with -race
make test-coverage    # generate coverage report in .coverage/
make test-patch-coverage  # patch-only coverage vs origin/main
make tools            # install all pinned dev tools
make setup-hooks      # install dev tools + git hooks (lefthook + commitlint)
make cli-smoke-build      # build the modkit CLI binary
make cli-smoke-scaffold   # smoke test CLI scaffold flow
```

Run a single package's tests:
```bash
cd modkit/kernel && go test -race ./...
```

Run tests for all workspace modules (including examples):
```bash
# make test does this automatically by iterating all go.mod files
make test
```

## Architecture

modkit is a multi-module Go workspace (`go.work`). The root module is `github.com/go-modkit/modkit`. Examples are separate modules under `examples/`.

```
modkit/
├── modkit/           # Core framework packages (the library)
│   ├── module/       # Public API: ModuleDef, ProviderDef, Token, Resolver
│   ├── kernel/       # Graph builder, visibility enforcer, lazy singleton container, bootstrap
│   ├── http/         # Chi-based HTTP adapter: Router, RegisterRoutes, Serve, middleware
│   ├── logging/      # Logger interface + slog adapter + nop
│   ├── config/       # Typed env config module helpers
│   ├── data/         # DB providers: sqlmodule (shared contract), postgres, sqlite sub-packages
│   └── testkit/      # Test harness: bootstrap helpers, provider overrides
├── examples/         # Runnable apps (separate go.mod each)
│   ├── hello-simple/ # Minimal bootstrap, no Docker
│   ├── hello-mysql/  # Full CRUD: auth, users, middleware, sqlc, swagger
│   ├── hello-postgres/
│   └── hello-sqlite/
├── cmd/modkit/       # CLI tool (cobra): `modkit new app <name>`
└── docs/             # Guides and architecture reference
```

### Bootstrap Flow

`kernel.Bootstrap(rootModule)` does four things in order:

1. **Build graph** — flattens the import tree depth-first, rejects cycles and duplicate names
2. **Build visibility** — computes which tokens each module can access (own providers + exported tokens from imports)
3. **Create container** — registers provider factory functions (not built yet)
4. **Build controllers** — calls each `ControllerDef.Build`, which triggers lazy provider construction

Providers are lazy singletons: built on first `Get()`, cached for subsequent calls.

### Key Interfaces

```go
// A module must be a pointer and implement:
type Module interface {
    Definition() ModuleDef
}

// Controllers must implement to register HTTP routes:
type RouteRegistrar interface {
    RegisterRoutes(router Router)
}

// Resolve a typed dependency inside a Build function:
svc, err := module.Get[T](r, "token.name")
```

### Token Naming Convention

Tokens follow `"module.component"` — e.g., `"users.service"`, `"db.connection"`. Shared SQL contract tokens live in `sqlmodule`: `sqlmodule.TokenDB`, `sqlmodule.TokenDialect`.

## Conventions

- Modules must be **pointers** — pointer identity determines shared import deduplication
- `Definition()` must be **pure/deterministic** — no mutable side effects or counters
- Resolve deps via `r.Get()` inside `Build` functions; always check `err` before use
- Controllers call explicit `r.Handle(method, path, handler)` — no reflection, no decorators
- Error handling: return/wrap errors; no panic-driven control flow
- Tests: `*_test.go` alongside implementation, table-driven where useful

## Anti-Patterns

- Do not add reflection or decorator magic to user-facing APIs
- Do not make `Definition()` stateful
- Do not wire across module boundaries without going through the export/import mechanism
- Do not hand-edit generated files (`*.sql.go`, swagger-generated docs)

## Branch and PR Workflow

- `main` is integration-only — never commit directly on it
- Create work on a worktree branch: `git worktree add .worktrees/<task> -b feat/<task> main`
- Before opening a PR: `make fmt && make lint && make vuln && make test && make test-coverage`
- If the repo has a PR template (`.github/pull_request_template.md`), follow it exactly and actively check/uncheck each item
- Include `Resolves #<number>` lines for any GitHub issues the PR addresses
- Commit messages follow Conventional Commits: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`, `perf:`, `ci:` — header ≤80 characters
