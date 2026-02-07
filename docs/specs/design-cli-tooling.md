# Design Spec: modkit CLI

**Status:** Draft
**Date:** 2026-02-07
**Author:** Sisyphus (AI Agent)

## 1. Overview

The `modkit` CLI is a developer tool to accelerate the creation and management of modkit applications. It automates repetitive tasks like creating new modules, providers, and controllers, ensuring that generated code follows the framework's strict architectural patterns.

## 2. Goals

*   **Speed:** Reduce setup time for new features from ~5 minutes to ~5 seconds.
*   **Consistency:** Enforce project structure and naming conventions (e.g., `NewXxxModule`, `Definition()`).
*   **Discovery:** Help users discover framework features through interactive prompts and templates.
*   **Zero Magic:** Generated code should be plain, readable Go code that the user owns.

## 3. Scope

### Phase 1 (MVP)
*   `modkit new app <name>`: Scaffold a new project with go.mod, main.go, and initial app module.
*   `modkit new module <name>`: Create a new module directory with `module.go`.
*   `modkit new provider <name>`: Add a provider to an existing module.
*   `modkit new controller <name>`: Add a controller to an existing module.

### Phase 2 (Future)
*   Interactive mode (TUI).
*   Graph visualization (`modkit graph`).
*   Migration helpers.

## 4. User Experience

### 4.1. Creating a New App

```bash
$ modkit new app my-service
Created my-service/
Created my-service/go.mod
Created my-service/cmd/api/main.go
Created my-service/internal/modules/app/module.go

Run:
  cd my-service
  go mod tidy
  go run cmd/api/main.go
```

### 4.2. Adding a Module

```bash
$ cd my-service
$ modkit new module users
Created internal/modules/users/module.go
Created internal/modules/users/module_test.go
```

### 4.3. Adding a Provider

```bash
$ modkit new provider service --module users
Created internal/modules/users/service.go
Updated internal/modules/users/module.go (registered provider)
```

## 5. Implementation Details

### 5.1. Directory Structure

The CLI will assume a standard layout but allow configuration:

```text
root/
|- cmd/
|- internal/
   |- modules/
```

### 5.2. Templates

Templates will be embedded in the CLI binary using `embed`. They will use `text/template`.

**Example `module.go.tpl`:**

```go
package {{.Package}}

import "github.com/go-modkit/modkit/modkit/module"

type {{.Name | Title}}Module struct {}

func (m *{{.Name | Title}}Module) Definition() module.ModuleDef {
    return module.ModuleDef{
        Name: "{{.Name}}",
        Providers: []module.ProviderDef{},
        Controllers: []module.ControllerDef{},
    }
}
```

### 5.3. Code Modification

For `new provider` and `new controller`, the CLI needs to edit existing `module.go` files to register the new components.
*   **Strategy:** AST parsing and modification using `dave/dst` (preferred for preserving comments) or regex-based insertion if AST is too complex for MVP.
*   **MVP Decision:** Use AST parsing to robustly locate `Providers: []module.ProviderDef{...}` and append the new definition.

## 6. Architecture

```text
cli/
|- main.go           # Entry point (cobra)
|- internal/
   |- cmd/           # Command implementations (new, version, etc.)
   |- generator/     # Template rendering logic
   |- ast/           # Code modification logic
   |- templates/     # Embedded .tpl files
```

## 7. Dependencies

*   `spf13/cobra`: Command structure.
*   `dave/dst`: AST manipulation (decorator-aware).
*   `golang.org/x/mod`: Go module manipulation.

## 8. Success Metrics

1.  **Time to Hello World:** < 1 minute including installation.
2.  **Correctness:** Generated code compiles immediately (`go build ./...` passes).
3.  **Adoption:** Used in the "Getting Started" guide as the primary way to start.
