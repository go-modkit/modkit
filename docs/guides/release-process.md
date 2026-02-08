# Release Process

This repository uses a unified SDLC flow for CI and releases:

1. Pull request checks run on each PR:
   - PR title semantic validation
   - lint + vulnerability scan
   - coverage tests
   - CLI smoke scaffolding checks
2. Merges to `main` trigger the release workflow.
3. `go-semantic-release` determines whether a new semantic version should be released from Conventional Commits.
4. If a version is released, GoReleaser builds and publishes CLI artifacts to the GitHub Release:
   - `darwin/amd64`
   - `darwin/arm64`
   - `linux/amd64`
   - `linux/arm64`
   - `windows/amd64`
5. Release assets include archives and `checksums.txt`.

## Quality Gates

Before creating or updating a PR, run:

```bash
make fmt && make lint && make vuln && make test && make test-coverage
make cli-smoke-build && make cli-smoke-scaffold
```

Branch protection should require these workflow checks on pull requests:

- `Validate PR Title`
- `Quality (Lint & Vuln)`
- `test`
- `CLI Smoke Test`
- `Analyze` (CodeQL)

## Versioning Rules

- `fix:` -> patch
- `feat:` -> minor
- `feat!:` / `BREAKING CHANGE:` -> major (or minor during initial development when configured)
- docs/chore/test/ci-only commits do not release by default
