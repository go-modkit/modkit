# modkit

modkit is a Go-idiomatic backend service framework built around an explicit module system. The MVP focuses on deterministic bootstrapping, explicit dependency resolution, and a thin HTTP adapter.

## Status

This repository is in early MVP implementation. APIs and structure may change before v0.1.0.

## What Is modkit?

modkit provides:
- A module metadata model (imports/providers/controllers/exports) to compose application boundaries.
- A kernel that builds a module graph, enforces visibility, and resolves providers/controllers.
- A minimal HTTP adapter that wires controller instances to routing without reflection.

See `docs/design/mvp.md` for the canonical architecture and scope.

## Quickstart

```bash
go get github.com/aryeko/modkit
```

Minimal sketch (omitting error handling details):

```go
package main

import (
    "log"

    mkhttp "github.com/aryeko/modkit/modkit/http"
    "github.com/aryeko/modkit/modkit/kernel"
    "github.com/aryeko/modkit/modkit/module"
)

type AppModule struct{}

func (m AppModule) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "app",
        Providers: []module.ProviderDef{
            // ...
        },
        Controllers: []module.ControllerDef{
            // ...
        },
    }
}

func main() {
    app, err := kernel.Bootstrap(AppModule{})
    if err != nil {
        log.Fatal(err)
    }

    router := mkhttp.NewRouter()
    if err := mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers); err != nil {
        log.Fatal(err)
    }

    if err := mkhttp.Serve(":8080", router); err != nil {
        log.Fatal(err)
    }
}
```

Guides:
- `docs/guides/getting-started.md`
- `docs/guides/modules.md`
- `docs/guides/testing.md`

## Tooling

- Format: `make fmt` (uses `gofmt` and `goimports`)
- Lint: `make lint` (uses `golangci-lint`)
- Vulnerability scan: `make vuln` (uses `govulncheck`)
- Details: `docs/tooling.md`

## Architecture Overview

- **module**: metadata for imports/providers/controllers/exports.
- **kernel**: builds the module graph, enforces visibility, and bootstraps an app container.
- **http**: adapts controller instances to routing without reflection.

For details, start with `docs/design/mvp.md`.

## NestJS Inspiration

The module metadata model is inspired by NestJS modules (imports/providers/controllers/exports), but the implementation is Go-idiomatic and avoids reflection. NestJS is a conceptual reference only.
