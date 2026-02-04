# Contributing to modkit

Thanks for your interest in contributing. This project is in early MVP development, so the best way to help is to coordinate with maintainers before large changes.

## Development setup

```bash
go test ./...
```

## Guidelines

- Follow Go formatting with `gofmt`.
- Run `make fmt` before committing.
- Run `make lint` for lint checks.
- Run `make vuln` for Go vulnerability checks.
- See `docs/tooling.md` for tool install and usage details.
- Keep changes focused and aligned to the current phase docs under `docs/implementation/`.
- Prefer small, reviewable PRs.

## Code of Conduct

This project follows the Code of Conduct in `CODE_OF_CONDUCT.md`.
