# Design Spec: Postgres + SQLite Module Support (P1)

**Status:** Ready for implementation
**Date:** 2026-02-09
**Author:** Sisyphus (AI Agent)
**Related PRD:** `docs/specs/prd-modkit-core.md` (adoption follow-up)

## 1. Overview

This spec defines first-class Postgres and SQLite support for modkit while preserving the current modular architecture:

- explicit module boundaries,
- explicit provider tokens,
- deterministic bootstrap,
- no reflection or hidden global state.

The goal is to make data-layer adoption easier than the current MySQL-only path and provide a stable contract that feature modules can depend on regardless of SQL driver choice.

## 2. Current State and Problem

The repository currently demonstrates SQL integration through `examples/hello-mysql/internal/modules/database` and exports a single DB token (`database.db`).

Adoption friction:

1. Only MySQL is showcased, while many evaluators prefer Postgres or SQLite first.
2. Config keys are MySQL-specific (`MYSQL_DSN`) in the main production-like example.
3. No documented portable contract for "SQL provider module" beyond example-specific code.

## 3. Goals

1. Ship Postgres and SQLite support first (before additional DB providers).
2. Keep feature modules database-agnostic by using a stable exported DB token.
3. Preserve compatibility with existing modular design and visibility rules.
4. Support both quick local evaluation (SQLite) and production-like setups (Postgres).
5. Keep implementation explicit and lightweight (standard `database/sql` first).

## 4. Non-Goals

1. No ORM abstraction layer.
2. No automatic query portability between SQL dialects.
3. No runtime driver hot-swap after bootstrap.
4. No replacement of existing MySQL example in this phase.
5. No generic migration framework in core modkit.

## 5. Design Principles

1. **Stable contract token:** consumers depend on one DB token, not driver internals.
2. **Driver modules are normal modules:** each driver has explicit imports/providers/exports.
3. **Dialect is explicit:** dialect-specific behavior is provided by explicit provider token, not implicit type checks.
4. **Portable feature modules:** users/auth/etc. modules should import a DB contract module and remain unchanged across drivers.
5. **Fail fast with typed errors:** invalid DSN/path/driver config should be explicit and deterministic.

## 6. Proposed Package and Module Shape

### 6.1 Core Contract Package

Add a small reusable contract package:

- `modkit/data/sqlmodule`

Proposed tokens:

```go
const (
    TokenDB      module.Token = "database.db"
    TokenDialect module.Token = "database.dialect"
)

type Dialect string

const (
    DialectPostgres Dialect = "postgres"
    DialectSQLite   Dialect = "sqlite"
    DialectMySQL    Dialect = "mysql" // for existing compatibility
)
```

Rationale: this avoids token drift and gives feature modules one stable import path.

### 6.2 Driver Modules

Add provider modules under core examples-first path:

- `modkit/data/postgres`
- `modkit/data/sqlite`

Both modules:

1. Import a config module for DSN/path and pool settings.
2. Export `sqlmodule.TokenDB` (`*sql.DB`).
3. Export `sqlmodule.TokenDialect`.
4. Register cleanup hook to close DB pool in provider cleanup.

### 6.3 Options

```go
type Options struct {
    Config module.Module
}

func NewModule(opts Options) module.Module
```

Semantics follow existing pattern used by `auth` and `database` modules in examples.

## 7. Configuration Contract

### 7.1 Postgres

Required (minimum):

- `POSTGRES_DSN`

Optional:

- `POSTGRES_MAX_OPEN_CONNS`
- `POSTGRES_MAX_IDLE_CONNS`
- `POSTGRES_CONN_MAX_LIFETIME`

### 7.2 SQLite

Required (minimum):

- `SQLITE_PATH` (or DSN)

Optional:

- `SQLITE_BUSY_TIMEOUT`
- `SQLITE_JOURNAL_MODE`

Config values are resolved through `modkit/config` helpers and exported as explicit tokens in app config modules.

## 8. API and Resolver Semantics

Consumer modules should resolve dependencies through the contract package only:

```go
db, err := module.Get[*sql.DB](r, sqlmodule.TokenDB)
dialect, err := module.Get[sqlmodule.Dialect](r, sqlmodule.TokenDialect)
```

Behavior requirements:

1. Provider remains lazy singleton.
2. Cleanup remains LIFO through existing kernel lifecycle.
3. Visibility remains unchanged; only exported tokens are reachable.

## 9. Modular Alignment Check and Suggested Improvements

### 9.1 What Already Aligns

1. Existing `examples/hello-mysql/internal/modules/database/module.go` already uses explicit provider + cleanup and exports one DB token.
2. Existing module token style (`"module.component"`) already matches modkit conventions.

### 9.2 Improvements Required

1. **Move token ownership to shared contract package** to avoid driver-specific token duplication.
2. **Split driver concerns from feature modules** so `users`/`audit` never import driver packages directly.
3. **Add explicit dialect token** for rare SQL branching without hidden assumptions.
4. **Normalize config key guidance** with driver-specific prefixes and a cross-driver migration section.

No architectural mismatch was found that requires changing kernel/module APIs.

## 10. Testing Strategy

### 10.1 Unit Tests

Per driver module:

1. Build fails with missing required config.
2. Build returns `*sql.DB` and correct dialect token.
3. Cleanup handles nil/closed DB safely.

### 10.2 Integration Tests

1. Postgres: testcontainers-backed smoke test validates bootstrap + simple query.
2. SQLite: file-backed and in-memory smoke tests validate bootstrap + CRUD roundtrip.
3. Visibility test ensures only exported DB tokens are visible to importers.

### 10.3 Compatibility Tests

1. Existing MySQL example remains unchanged and green.
2. New Postgres and SQLite examples compile and pass `make test` in their modules.

## 11. Documentation Deliverables

1. New guide: `docs/guides/database-providers.md` (contract + driver modules).
2. Update `docs/guides/getting-started.md` with SQLite fast-start path.
3. Update `README.md` feature matrix with Postgres + SQLite examples.
4. Add migration note from MySQL-only pattern to shared SQL contract tokens.

## 12. Rollout Plan

### Story P1.1 - Shared SQL Contract

1. Introduce `modkit/data/sqlmodule` tokens and dialect type.
2. Keep MySQL example compatible by aliasing/adopting contract token constants.

### Story P1.2 - Postgres Module + Example

1. Add `modkit/data/postgres` module.
2. Add `examples/hello-postgres` as production-like reference app.

### Story P1.3 - SQLite Module + Example

1. Add `modkit/data/sqlite` module.
2. Add `examples/hello-sqlite` as fast-evaluation app.

### Story P1.4 - Docs + CI

1. Add provider docs and cross-link from README.
2. Add smoke checks for Postgres and SQLite examples.

## 13. Acceptance Criteria

This phase is complete when all are true:

1. A shared SQL contract token package exists and is documented.
2. Postgres and SQLite provider modules exist and export contract tokens.
3. At least one Postgres example and one SQLite example are runnable.
4. Existing MySQL path remains backward-compatible.
5. CI includes smoke coverage for new example paths.

## 14. Risks and Mitigations

1. **Risk:** Token drift across provider modules.
   - **Mitigation:** single contract package owns DB/dialect tokens.
2. **Risk:** Over-abstraction hides SQL differences.
   - **Mitigation:** keep only DB + dialect contract in v1; no query abstraction.
3. **Risk:** Driver setup complexity slows adoption.
   - **Mitigation:** prioritize SQLite quickstart and doc-smoke it.

## 15. Future Enhancements (Out of Scope)

1. MSSQL and Oracle providers.
2. Transaction helper module conventions.
3. Read/write split contracts.
4. Connection health/metrics adapters for observability packages.
