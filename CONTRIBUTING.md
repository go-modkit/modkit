# Contributing to modkit

Thanks for your interest in contributing! modkit is in early development, so the best way to help is to coordinate with maintainers before large changes.

## Getting Started

### Prerequisites

- Go 1.25.7+
- Docker (for running the example app)
- Make

We pin the patch level to 1.25.7 in CI to align with vulnerability scanning and keep a consistent security posture.

### Clone and Test

```bash
git clone https://github.com/go-modkit/modkit.git
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

```text
<type>(<scope>): <short summary>
```

Examples:
- `feat: add user authentication`
- `fix(http): handle connection timeout`
- `docs: update installation guide`

Valid types: `feat`, `fix`, `docs`, `test`, `chore`, `refactor`, `perf`, `ci`

**Note**: Commit message headers must be <=80 characters.

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

### Run CLI Smoke Checks

These are the same checks used by the CI `cli-smoke` job:

```bash
make cli-smoke-build
make cli-smoke-scaffold
```

### Install Development Tools

```bash
make tools
```

`make tools` installs tool versions pinned by the repository.

## Contribution Guidelines

### Before You Start

- Check existing issues to avoid duplicating work
- For large changes, open an issue first to discuss the approach
- Read the [Architecture Guide](docs/architecture.md) to understand the codebase

### Pull Request Process

1. Fork the repository
2. Update `main` and create a feature worktree branch:
   - `git fetch origin && git switch main && git pull --ff-only origin main`
   - `git worktree add .worktrees/my-feature -b feat/my-feature main`
   - Work from `.worktrees/my-feature` (do not commit on `main`)
3. Make your changes with tests
4. Run `make fmt && make lint && make vuln && make test && make test-coverage`
   - Also run CLI gate: `make cli-smoke-build && make cli-smoke-scaffold`
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

New to modkit? Look for issues labeled [`good first issue`](https://github.com/go-modkit/modkit/labels/good%20first%20issue) for beginner-friendly tasks:

- Documentation improvements
- Test coverage
- Example app enhancements
- Bug fixes with clear reproduction steps

## Questions?

Open an issue or start a discussion. We're happy to help!

## Releases

Releases are automated using [go-semantic-release](https://github.com/go-semantic-release/semantic-release).

### How It Works

When changes are merged to `main`, the release workflow analyzes commit messages:

| Commit Type | Release Action | Example |
|-------------|----------------|---------|
| `feat:` | Minor version bump (0.1.0 -> 0.2.0) | New API method |
| `fix:` | Patch version bump (0.1.0 -> 0.1.1) | Bug fix |
| `feat!:` or `BREAKING CHANGE` | v0.x: minor bump; v1+: major bump | Breaking API change |
| `docs:`, `chore:`, `refactor:`, `test:`, `ci:` | No release | Documentation, tooling |

### Versioning Strategy

modkit follows [Semantic Versioning](https://semver.org/):

- **v0.x.x** (current): API is evolving. Breaking changes (`feat!:`) bump the minor version (0.1.0 → 0.2.0) due to `allow-initial-development-versions` setting
- **v1.0.0** (future): Stable API with backward compatibility guarantees. Breaking changes will bump major version (1.0.0 → 2.0.0)
- **v2+**: Major versions for breaking changes (requires `/v2` import path)

### Using a Specific Version

```bash
go get github.com/go-modkit/modkit@v0.1.0
go get github.com/go-modkit/modkit@latest
```
