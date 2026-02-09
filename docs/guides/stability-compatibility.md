# Stability and Compatibility Policy

This document defines what modkit guarantees today and what changes before and after `v1.0.0`.

## Current Phase

- Current phase: `v0.x` (pre-1.0)
- APIs may evolve quickly.
- Breaking changes can occur in minor releases when necessary.

## Versioning Rules

modkit uses Semantic Versioning with early-development behavior:

- `fix:` -> patch release (`v0.14.0` -> `v0.14.1`)
- `feat:` -> minor release (`v0.14.0` -> `v0.15.0`)
- `feat!:` / `BREAKING CHANGE:` -> minor release while on `v0.x`

After `v1.0.0`:

- Breaking changes will require a major release (`v1.x` -> `v2.0.0`)
- Backward compatibility is expected for public APIs inside a major line.

## Compatibility Guarantees

### During `v0.x`

- We keep migration notes in release PRs/changelog for behavior-impacting changes.
- We prioritize preserving documented behavior in guides and examples.
- We do not guarantee zero breaking changes across minor versions.

### After `v1.0.0`

- Public API changes will follow SemVer major boundaries.
- Deprecated APIs will include replacement guidance before removal.

## Go Toolchain Support Policy

- Supported baseline: Go `1.25.x`
- CI enforcement baseline: Go `1.25.7`

What this means:

- A change is considered supported only if CI validates it.
- If support policy changes, we update CI and docs in the same change.

## Upgrade Expectations

Before upgrading between modkit versions:

1. Read release notes.
2. Run project verification gates:

```bash
make fmt && make lint && make vuln && make test && make test-coverage
make cli-smoke-build && make cli-smoke-scaffold
```

3. Validate your app bootstrapping and HTTP routes in staging.

## Non-Guarantees (Important)

- We do not guarantee stability for undocumented/internal behavior.
- We do not guarantee migration tooling for every breaking change during `v0.x`.
- We do not guarantee compatibility claims that are not backed by CI.
