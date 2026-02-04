# S1: Commit Validation with Lefthook + Go Commitlint

## Status

🟢 Complete

## Overview

Add automated commit message validation using Go-native tooling to enforce conventional commits format. This provides immediate feedback to developers and enables future automation of changelogs and releases.

## Goals

1. Enforce conventional commit format at commit time
2. Use Go-native tools (no Node.js dependency)
3. Minimal setup friction for contributors
4. Enable future changelog/release automation

## Non-Goals

- Pre-commit hooks for formatting/linting (deferred to S4)
- CI-based commit validation (local validation is sufficient)
- Custom commit message templates

---

## Tools

### Lefthook

- **Purpose**: Git hooks manager written in Go
- **Why**: Fast, single binary, no runtime dependencies
- **Install**: `go install github.com/evilmartians/lefthook@latest`
- **Docs**: https://github.com/evilmartians/lefthook

### conventionalcommit/commitlint

- **Purpose**: Validates commit messages against conventional commits spec
- **Why**: Go-native, simple CLI
- **Install**: `go install github.com/conventionalcommit/commitlint@latest`
- **Docs**: https://github.com/conventionalcommit/commitlint

---

## Implementation

### 1. Create lefthook.yml

Configuration file at repository root:

```yaml
# Lefthook configuration for git hooks
# Setup: make setup-hooks

commit-msg:
  commands:
    commitlint:
      run: '"${GOBIN:-$(go env GOPATH)/bin}/commitlint" lint --message "{1}"'
```

**Notes:**
- `{1}` is Lefthook's placeholder for the commit message file path
- The hook runs on both `git commit` and `git commit --amend` by default

### 2. Create tools/tools.go

Track development tool dependencies using the Go tools pattern:

```go
//go:build tools
// +build tools

// Package tools tracks development tool dependencies.
package tools

import (
	_ "github.com/conventionalcommit/commitlint"
	_ "github.com/evilmartians/lefthook"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
```

Then run `go mod tidy` to add these to `go.mod` with pinned versions.

### 3. Update Makefile

Add targets for setup and validation:

```makefile
# Install all development tools (tracked in tools/tools.go)
.PHONY: tools
tools:
	@echo "Installing development tools..."
	@cat tools/tools.go | grep _ | awk '{print $$2}' | xargs -I {} sh -c 'go install {}'
	@echo "✓ All tools installed"

# Install development tools and setup git hooks
setup-hooks: tools
	@echo "Setting up git hooks..."
	lefthook install
	@echo "✓ Git hooks installed successfully"

# Validate a commit message (for CI or manual testing)
lint-commit:
	@echo "$(MSG)" | $(COMMITLINT) lint
```

**Usage:**
```bash
# Install all tools
make tools

# Setup hooks (one-time after clone, runs 'make tools' first)
make setup-hooks

# Manual validation (testing)
make lint-commit MSG="feat: add new feature"
```

### 4. Update CONTRIBUTING.md

Add to "Getting Started" section after "Prerequisites":

```markdown
### Setup Git Hooks

After cloning the repository, run once to enable commit message validation:

\`\`\`bash
make setup-hooks
\`\`\`

This installs git hooks that validate commit messages follow the [Conventional Commits](https://www.conventionalcommits.org/) format.

If you see a commit validation error, ensure your message follows this format:

\`\`\`
<type>(<scope>): <short summary>

<optional body>

<optional footer>
\`\`\`

Examples:
- `feat: add user authentication`
- `fix(http): handle connection timeout`
- `docs: update installation guide`
```

---

## Validation Rules

### Enforced by commitlint

1. **Header format**: `<type>(<scope>): <description>`
   - Scope is optional
   - Max 50 characters for full header (per CONTRIBUTING.md)

2. **Valid types**: 
   - `feat` - New feature
   - `fix` - Bug fix
   - `docs` - Documentation changes
   - `test` - Test changes
   - `chore` - Build/tooling changes
   - `refactor` - Code refactoring
   - `perf` - Performance improvements
   - `ci` - CI/CD changes

3. **Description rules**:
   - Lowercase
   - No period at end
   - Imperative mood ("add" not "added")

4. **Breaking changes**:
   - Add `!` after type/scope: `feat!: breaking change`
   - Or use footer: `BREAKING CHANGE: description`

---

## Testing

### Manual Testing

Test the hook installation:

```bash
# 1. Setup hooks
make setup-hooks

# 2. Try a bad commit message
git commit -m "bad message"
# Should fail with validation error

# 3. Try a good commit message
git commit -m "feat: test commit validation"
# Should succeed
```

### Bypass (for emergencies only)

```bash
# Skip hooks if absolutely necessary
git commit --no-verify -m "emergency fix"
```

---

## Migration Notes

### Existing Contributors

Contributors who already have the repository cloned should:

1. Pull the latest changes
2. Run `make setup-hooks` once
3. Continue normal workflow

### CI/CD

No changes needed - validation happens locally at commit time.

---

## Future Extensions

Once S1 is complete, this foundation enables:

- **S2**: Changelog generation from commit messages
- **S3**: Automated releases with semantic versioning
- **S4**: Pre-commit hooks for `make fmt && make lint`

---

## References

- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Lefthook Documentation](https://github.com/evilmartians/lefthook)
- [Semantic Versioning](https://semver.org/)
