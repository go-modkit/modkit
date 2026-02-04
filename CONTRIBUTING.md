# Contributing to modkit

Thanks for your interest in contributing. This project is in early MVP development, so the best way to help is to coordinate with maintainers before large changes.

## Development setup

```bash
go test ./...
```

### Setup Git Hooks

After cloning the repository, run once to enable commit message validation:

```bash
make setup-hooks
```

This installs git hooks that validate commit messages follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <short summary>
```

Examples:
- `feat: add user authentication`
- `fix(http): handle connection timeout`
- `docs: update installation guide`

Valid types: `feat`, `fix`, `docs`, `test`, `chore`, `refactor`, `perf`, `ci`

## Guidelines

- Contribute via fork + pull request (recommended). Direct pushes to the main repo are restricted.
- Follow Go formatting with `gofmt`.
- Run `make fmt` before committing.
- Run `make lint` for lint checks.
- Run `make vuln` for Go vulnerability checks.
- See `docs/tooling.md` for tool install and usage details.
- Keep changes focused and aligned to the current phase docs under `docs/implementation/`.
- Prefer small, reviewable PRs.

## Code of Conduct

This project follows the Code of Conduct in `CODE_OF_CONDUCT.md`.
