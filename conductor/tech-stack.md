# Technology Stack

## Core
- **Language:** Go 1.25
- **Framework:** `modkit` (This project is the framework itself)

## Runtime Dependencies
- **HTTP Router:** `github.com/go-chi/chi/v5` - Used as the underlying router for the `modkit/http` adapter.
- **Logging:** `log/slog` (Standard Library) - The core logging interface wraps the standard structured logger.

## Development & Tooling
- **Build System:** `Make` - Orchestrates build, test, and lint tasks.
- **Linter:** `golangci-lint` - Enforces strict code style and static analysis.
- **Git Hooks:** `lefthook` - Manages pre-commit and pre-push hooks.
- **Commit Standards:** `commitlint` - Enforces [Conventional Commits](https://www.conventionalcommits.org/).
- **Security:** `govulncheck` - Scans for known vulnerabilities in dependencies.
- **Testing:** Standard `go test` with race detection (`-race`).
