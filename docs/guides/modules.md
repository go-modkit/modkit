# Modules

modkit modules are declarative metadata used by the kernel to build a graph, enforce visibility, and construct providers/controllers.

## ModuleDef

`module.ModuleDef` has four key fields:
- `Imports`: other modules whose exported tokens are visible to this module.
- `Providers`: providers created in this module.
- `Controllers`: controllers created in this module.
- `Exports`: tokens from this module that are visible to importers.

A module is any type that implements:

```go
Definition() module.ModuleDef
```

## Tokens and Providers

Tokens are `module.Token` values that identify providers. Providers are singletons created lazily when requested.

```go
const TokenDB module.Token = "db"

module.ProviderDef{
    Token: TokenDB,
    Build: func(r module.Resolver) (any, error) {
        return openDB(), nil
    },
}
```

## Controllers

Controllers are built by the kernel during bootstrap and returned in `kernel.App.Controllers`. The HTTP adapter expects controllers to implement `http.RouteRegistrar`.

```go
module.ControllerDef{
    Name: "UsersController",
    Build: func(r module.Resolver) (any, error) {
        dbAny, err := r.Get(TokenDB)
        if err != nil {
            return nil, err
        }
        return NewUsersController(dbAny.(*sql.DB)), nil
    },
}
```

## Visibility Rules

Visibility is enforced by the kernel:
- Providers in the same module are visible to each other.
- Imported modules expose only tokens listed in their `Exports`.
- Re-exported tokens must already be visible through providers or imports.

This keeps module boundaries explicit and prevents accidental leakage of internal providers.

## Example Layout

```
modkit/
  module/
  kernel/
  http/

app/
  module.go
  controllers.go
  providers.go
```

## Tips

- Keep module names unique across the graph to avoid bootstrap errors.
- Prefer small modules with explicit exports rather than large monoliths.
- Use `docs/design/mvp.md` for canonical architecture details.
