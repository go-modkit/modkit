# Repository Guidelines

This file provides short, focused guidance for contributors and AI agents. Keep instructions concise and non-conflicting. For path-specific guidance, prefer scoped instruction files rather than growing this document.

## Project Structure
- Core library packages: `modkit/` (`module`, `kernel`, `http`, `logging`).
- Example apps: `examples/` (see `examples/hello-mysql/README.md`).
- Design and phase docs: `docs/design/` and `docs/implementation/`.

## Tooling & Commands
- Format: `make fmt` (runs `gofmt`, `goimports`).
- Lint: `make lint` (runs `golangci-lint`).
- Vulnerability scan: `make vuln` (runs `govulncheck`).
- Tests: `make test` and `go test ./examples/hello-mysql/...`.

## Coding Conventions
- Use `gofmt` formatting and standard Go naming.
- Packages are lowercase, short, and stable.
- Keep exported API minimal; prefer explicit errors over panics.

## Testing Guidance
- Use Go’s `testing` package and keep tests close to code.
- Name tests `TestXxx` and use table-driven tests where it clarifies cases.
- Integration tests should be deterministic; keep external dependencies isolated.

## Commit & PR Hygiene
- Use conventional prefixes: `feat:`, `fix:`, `docs:`, `chore:`.
- One logical change per commit.
- PRs should include summary + validation commands run.

## Agent Instruction Layout
- Agent instructions can be stored in `AGENTS.md` files; the closest `AGENTS.md` in the directory tree takes precedence.
- Keep instructions scoped and avoid conflicts across files.
