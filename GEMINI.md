# modkit

**Context:** Go framework for building modular backend services, inspired by NestJS.

## Project Overview

`modkit` is a framework designed to bring NestJS-style module organization to Go. It focuses on explicit configuration over magic, avoiding reflection and decorators in favor of type-safe, deterministic graph construction.

**Key Features:**
- **Module System:** Explicit boundaries with `imports`, `exports`, `providers`, and `controllers`.
- **Dependency Injection:** String token-based injection without reflection.
- **Visibility:** Enforced access rules; modules can only access what they import.
- **HTTP:** Thin adapter over `chi` router.

## Directory Structure

- `modkit/`: Core framework source code.
    - `kernel/`: Graph builder and bootstrap logic.
    - `module/`: Metadata types (`ModuleDef`, `ProviderDef`).
    - `http/`: HTTP adapter and router.
    - `logging/`: Logging interfaces.
- `examples/`: Usage examples.
    - `hello-simple/`: Minimal example.
    - `hello-mysql/`: Full CRUD example with MySQL.
- `docs/`: Documentation and guides.
- `tools/`: Development tool dependencies.

## Building and Running

The project uses `make` for common tasks.

### Core Commands
- **Format:** `make fmt` (Runs `gofmt` and `goimports`)
- **Lint:** `make lint` (Runs `golangci-lint`)
- **Test:** `make test` (Runs `go test -race` for all modules)
- **Coverage:** `make test-coverage` (Generates coverage report)
- **Vulnerability Check:** `make vuln` (Runs `govulncheck`)

### Setup
- **Install Tools:** `make tools`
- **Git Hooks:** `make setup-hooks` (Installs `lefthook` for pre-commit checks)

### Running Examples
To run the `hello-mysql` example:
```bash
cd examples/hello-mysql
make run
```
Then verify: `curl http://localhost:8080/health`

## Development Conventions

### Coding Style
- Follow standard Go idioms.
- Use `make fmt` to ensure compliance with `gofmt` and `goimports`.
- **Linting:** Strict linting via `golangci-lint`. Ensure `make lint` passes before committing.

### Commit Messages
- **Format:** [Conventional Commits](https://www.conventionalcommits.org/) are enforced via `commitlint`.
- **Structure:** `<type>(<scope>): <description>`
    - Types: `feat`, `fix`, `docs`, `test`, `chore`, `refactor`, `perf`, `ci`.
    - Example: `feat(http): add middleware support`

### Architecture
- **Modules:** Must be pointers to struct types to ensure stable identity.
- **Providers:** Registered via factories, built lazily, and cached as singletons.
- **Controllers:** Registered with the router; must implement route registration explicitly.
- **No Global State:** Dependencies should flow through the `App` instance created by `kernel.Bootstrap`.

### Versioning
- Uses Semantic Versioning.
- Releases are automated based on commit messages.
