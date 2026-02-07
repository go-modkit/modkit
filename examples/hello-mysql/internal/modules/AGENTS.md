# HELLO-MYSQL MODULES KNOWLEDGE BASE

## OVERVIEW
Domain module layer for the sample app (`app`, `auth`, `users`, `database`, `audit`). Keep boundaries explicit via imports/exports.

## STRUCTURE
```text
internal/modules/
|- app/       # Top-level app composition and health routes
|- auth/      # Login, claims, middleware, token config
|- users/     # CRUD endpoints, service, repo integration
|- database/  # DB provider module + cleanup hooks
`- audit/     # Audit service module
```

## WHERE TO LOOK
| Task | Location | Notes |
|---|---|---|
| Root composition order | `app/module.go` | Imports all required modules |
| Token/auth behavior | `auth/{module,token,middleware,handler}.go` | Protected route contracts |
| Users domain flow | `users/{module,controller,service,repo_mysql}.go` | Request -> service -> persistence |
| DB lifecycle hooks | `database/{module,cleanup}.go` | Close pools using provider cleanup |
| Audit integration | `audit/{module,service}.go` | Consumes exports, no hidden coupling |

## CONVENTIONS
- Module names and provider tokens stay stable and explicit.
- `Definition()` methods remain deterministic and side-effect free.
- Services return explicit errors; controllers map to Problem Details.
- Keep tests adjacent to module code; preserve table/subtest style.

## ANTI-PATTERNS
- Do not resolve dependencies and ignore `err` from `r.Get()`.
- Do not add cross-module direct struct wiring that bypasses exports.
- Do not mix transport validation logic deep inside repositories.
- Do not mutate module definition fields at runtime.

## COMMANDS
```bash
go test ./examples/hello-mysql/internal/modules/...
go test ./examples/hello-mysql/internal/modules/auth/...
go test ./examples/hello-mysql/internal/modules/users/...
```

## NOTES
- Inherit root and `examples/hello-mysql` guidance; this file only adds module-layer rules.
