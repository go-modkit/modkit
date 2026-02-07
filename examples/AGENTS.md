# EXAMPLES KNOWLEDGE BASE

## OVERVIEW
`examples/` contains runnable consuming apps in separate Go modules.

## STRUCTURE
```text
examples/
|- hello-simple/   # Minimal single-file app
`- hello-mysql/    # Full app with auth/users/db/sqlc/swagger
```

## WHERE TO LOOK
| Task | Location | Notes |
|---|---|---|
| Minimal usage of modkit bootstrap | `examples/hello-simple/main.go` | Canonical first example |
| Production-like wiring | `examples/hello-mysql/cmd/api/main.go` | Real server startup/shutdown |
| Example app workflow | `examples/hello-mysql/Makefile` | Compose/migrate/seed/sqlc/swagger |
| Example docs and runbook | `examples/hello-mysql/README.md` | Endpoints, auth, config, middleware |

## CONVENTIONS
- Keep examples runnable as standalone modules (`go.mod` local to each example).
- Favor explicit composition and readable flow over framework internals.
- Keep docs and commands in each example synchronized with behavior.

## ANTI-PATTERNS
- Do not couple examples to private internals from other example modules.
- Do not add hidden setup steps that are missing from README/Makefile.
- Do not hand-edit generated files in example subtrees (`sqlc`, swagger output).

## COMMANDS
```bash
go test ./examples/hello-simple/...
go test ./examples/hello-mysql/...
```

## NOTES
- See `examples/hello-mysql/AGENTS.md` for generation and integration-test rules.
