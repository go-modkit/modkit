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

Guides:
- `docs/guides/getting-started.md`
- `docs/guides/modules.md`
- `docs/guides/testing.md`

Example app:
- `examples/hello-mysql/README.md`

## Tooling

- See `docs/tooling.md`

## Architecture Overview

- **module**: metadata for imports/providers/controllers/exports.
- **kernel**: builds the module graph, enforces visibility, and bootstraps an app container.
- **http**: adapts controller instances to routing without reflection.

For details, start with `docs/design/mvp.md`.

## NestJS Inspiration

The module metadata model is inspired by NestJS modules (imports/providers/controllers/exports), but the implementation is Go-idiomatic and avoids reflection. NestJS is a conceptual reference only.
