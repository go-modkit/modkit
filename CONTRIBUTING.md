# Contributing to modkit

Thanks for your interest in contributing! modkit is in early development, so the best way to help is to coordinate with maintainers before large changes.

## Getting Started

### Prerequisites

- Go 1.25+
- Docker (for running the example app)
- Make

### Clone and Test

```bash
git clone https://github.com/aryeko/modkit.git
cd modkit
go test ./...
```

### Run the Example App

```bash
cd examples/hello-mysql
make run
```

Then test:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/users
```

### Setup Git Hooks

After cloning, ensure `$GOPATH/bin` is in your PATH:

```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)
export PATH="$(go env GOPATH)/bin:$PATH"
```

Then run once to enable commit message validation:

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

**Note**: Commit message headers must be ≤50 characters.

## Development Workflow

### Format Code

```bash
make fmt
```

Runs `gofmt` and `goimports`.

### Lint

```bash
make lint
```

Runs `golangci-lint`. See `.golangci.yml` for configuration.

### Vulnerability Check

```bash
make vuln
```

Runs `govulncheck`.

### Run Tests

```bash
make test
```

### Install Development Tools

```bash
# goimports (for make fmt)
go install golang.org/x/tools/cmd/goimports@latest

# golangci-lint (for make lint)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# govulncheck (for make vuln)
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## Contribution Guidelines

### Before You Start

- Check existing issues to avoid duplicating work
- For large changes, open an issue first to discuss the approach
- Read the [Architecture Guide](docs/architecture.md) to understand the codebase

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Make your changes with tests
4. Run `make fmt && make lint && make test`
5. Commit with a conventional prefix (`feat:`, `fix:`, `docs:`, `chore:`)
6. Open a pull request with a clear description

### Commit Prefixes

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation only
- `test:` — Test changes
- `chore:` — Build, CI, or tooling changes
- `refactor:` — Code change that doesn't fix a bug or add a feature

### Code Style

- Follow Go formatting (`gofmt`)
- Keep exported API minimal
- Prefer explicit errors over panics
- Write tests for new functionality

## Code of Conduct

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md).

## Good First Issues

New to modkit? Look for issues labeled [`good first issue`](https://github.com/aryeko/modkit/labels/good%20first%20issue) for beginner-friendly tasks:

- Documentation improvements
- Test coverage
- Example app enhancements
- Bug fixes with clear reproduction steps

## Questions?

Open an issue or start a discussion. We're happy to help!
