# MODKIT PACKAGE KNOWLEDGE BASE

## OVERVIEW
Core framework packages: metadata (`module`), runtime (`kernel`), HTTP adapter (`http`), logging (`logging`).

## STRUCTURE
```text
modkit/
|- module/    # Public metadata and validation
|- kernel/    # Graph + visibility + container + bootstrap
|- http/      # Router/server/middleware adapter around chi
`- logging/   # Logger contract + adapters
```

## WHERE TO LOOK
| Task | Location | Notes |
|---|---|---|
| Define/validate module metadata | `modkit/module/module.go` | Deterministic definition contract |
| Provider/controller token contracts | `modkit/module/{provider,controller,token}.go` | Exported API entry points |
| Build graph and visibility | `modkit/kernel/{graph,visibility}.go` | Import traversal and export checks |
| Runtime dependency resolution | `modkit/kernel/container.go` | Lazy singleton provider creation |
| Bootstrap flow | `modkit/kernel/bootstrap.go` | App creation + controller instantiation |
| Route registration contract | `modkit/http/{router,server,register}.go` | Explicit registration only |
| Logging adapters | `modkit/logging/{logger,slog,nop}.go` | Keep logger contract minimal |

## CONVENTIONS
- Prefer explicit behavior over metaprogramming; keep API surfaces clear.
- Keep exported APIs minimal; new exports require strong justification.
- Error values are typed/sentinel where useful; wrap with `%w`.
- Keep tests adjacent to implementation.

## ANTI-PATTERNS
- Do not add reflection/decorator-driven dependency injection.
- Do not hide dependency lookup failures; always return contextual errors.
- Do not introduce framework-global mutable state.
- Do not weaken visibility checks for convenience.

## PACKAGE COMMANDS
```bash
go test ./modkit/...
go test -race ./modkit/...
```

## NOTES
- Root policies still apply; this file adds package-local guidance only.
