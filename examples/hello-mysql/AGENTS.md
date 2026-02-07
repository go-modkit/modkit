# HELLO-MYSQL KNOWLEDGE BASE

## OVERVIEW
`hello-mysql` is a full consuming app module: HTTP API, MySQL, migrations, sqlc, swagger.

## STRUCTURE
```text
examples/hello-mysql/
|- cmd/            # api, migrate, seed entrypoints
|- internal/
|  |- modules/     # app, auth, users, database, audit
|  |- platform/    # mysql, config, logging adapters
|  |- middleware/  # CORS/rate-limit/timing
|  `- sqlc/        # generated queries/models (read-only)
|- migrations/     # SQL migrations
|- sql/            # source queries + sqlc config
`- docs/           # swagger output (generated)
```

## WHERE TO LOOK
| Task | Location | Notes |
|---|---|---|
| API startup and shutdown | `cmd/api/main.go` | Server wiring + graceful shutdown |
| DB migration CLI | `cmd/migrate/main.go` | Use for schema changes |
| Seed CLI | `cmd/seed/main.go` | Local data bootstrap |
| Domain wiring | `internal/modules` | Module boundaries and exports |
| MySQL connection lifecycle | `internal/platform/mysql` | `*sql.DB` setup and cleanup |
| RFC7807 errors | `internal/httpapi` + module `problem.go` files | Uniform error payloads |

## CONVENTIONS
- Keep routes grouped under `/api/v1`; keep `/docs` and `/swagger` outside API middleware group.
- Maintain RFC7807 response shape for validation and domain errors.
- Regenerate artifacts when source changes (`sql/queries.sql`, OpenAPI comments).

## ANTI-PATTERNS
- Do not edit generated files directly:
  - `internal/sqlc/*.go`
  - `docs/docs.go`
  - `docs/swagger.json`
  - `docs/swagger.yaml`
- Do not bypass module boundaries by cross-package direct state access.
- Do not hardcode production secrets in tests or sample config.

## COMMANDS
```bash
make run
make test
make migrate
make seed
make sqlc
make swagger
```

## NOTES
- Integration and smoke tests may require Docker/testcontainers availability.
- Module-specific conventions are in `examples/hello-mysql/internal/modules/AGENTS.md`.
