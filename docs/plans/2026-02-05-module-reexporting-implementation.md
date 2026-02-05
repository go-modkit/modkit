# Module Re-exporting Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement module re-exporting with validation, transitive visibility, and tests for re-export scenarios.

**Architecture:** Update kernel visibility validation to allow exporting imported tokens, add ambiguity detection for duplicate exports from imports, and ensure effective exports propagate transitively. Add unit tests for valid/invalid re-exports and transitive visibility.

**Tech Stack:** Go, modkit kernel tests.

---

### Task 1: Add failing visibility tests for re-exports

**Files:**
- Modify: `modkit/kernel/visibility_test.go`
- Test: `modkit/kernel/visibility_test.go`

**Step 1: Write the failing tests**

```go
func TestVisibilityAllowsReExportFromImport(t *testing.T) {
    token := module.Token("shared.token")
    imported := mod("Imported", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, []module.Token{token})
    reexporter := mod("Reexporter", []module.Module{imported}, nil, nil, []module.Token{token})

    g, err := kernel.BuildGraph(reexporter)
    if err != nil {
        t.Fatalf("BuildGraph failed: %v", err)
    }

    if _, err := kernel.BuildVisibility(g); err != nil {
        t.Fatalf("BuildVisibility failed: %v", err)
    }
}

func TestVisibilityRejectsReExportOfNonExportedImportToken(t *testing.T) {
    token := module.Token("private.token")
    imported := mod("Imported", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, nil)
    reexporter := mod("Reexporter", []module.Module{imported}, nil, nil, []module.Token{token})

    g, err := kernel.BuildGraph(reexporter)
    if err != nil {
        t.Fatalf("BuildGraph failed: %v", err)
    }

    _, err = kernel.BuildVisibility(g)
    if err == nil {
        t.Fatalf("expected error for re-exporting non-exported token")
    }

    var exportErr *kernel.ExportNotVisibleError
    if !errors.As(err, &exportErr) {
        t.Fatalf("unexpected error type: %T", err)
    }
    if exportErr.Module != "Reexporter" {
        t.Fatalf("unexpected module: %q", exportErr.Module)
    }
    if exportErr.Token != token {
        t.Fatalf("unexpected token: %q", exportErr.Token)
    }
}

func TestVisibilityRejectsAmbiguousReExport(t *testing.T) {
    token := module.Token("shared.token")
    left := mod("Left", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, []module.Token{token})
    right := mod("Right", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, []module.Token{token})
    reexporter := mod("Reexporter", []module.Module{left, right}, nil, nil, []module.Token{token})

    g, err := kernel.BuildGraph(reexporter)
    if err == nil {
        t.Fatalf("expected BuildGraph error for duplicate provider token")
    }
    var dupErr *kernel.DuplicateProviderTokenError
    if !errors.As(err, &dupErr) {
        t.Fatalf("unexpected error type: %T", err)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./modkit/kernel -run TestVisibility`  
Expected: FAIL because `BuildVisibility` does not exist and re-export logic is not implemented.

**Step 3: Commit**

```bash
git add modkit/kernel/visibility_test.go
git commit -m "test: add visibility re-export tests"
```

---

### Task 2: Add failing transitive re-export test

**Files:**
- Modify: `modkit/kernel/graph_test.go`
- Test: `modkit/kernel/graph_test.go`

**Step 1: Write the failing test**

```go
func TestGraphTransitiveReExportVisibility(t *testing.T) {
    token := module.Token("shared.token")

    modC := mod("C", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, []module.Token{token})
    modB := mod("B", []module.Module{modC}, nil, nil, []module.Token{token})
    modA := mod("A", []module.Module{modB}, nil, nil, nil)

    g, err := kernel.BuildGraph(modA)
    if err != nil {
        t.Fatalf("BuildGraph failed: %v", err)
    }

    visibility, err := kernel.BuildVisibility(g)
    if err != nil {
        t.Fatalf("BuildVisibility failed: %v", err)
    }

    if !visibility["A"][token] {
        t.Fatalf("expected token visible in A via transitive re-export")
    }
}

func TestGraphRejectsNonExportedTransitiveToken(t *testing.T) {
    token := module.Token("shared.token")

    modC := mod("C", nil, []module.ProviderDef{{Token: token, Build: buildNoop}}, nil, nil)
    modB := mod("B", []module.Module{modC}, nil, nil, nil)
    modA := mod("A", []module.Module{modB}, nil, nil, nil)

    g, err := kernel.BuildGraph(modA)
    if err != nil {
        t.Fatalf("BuildGraph failed: %v", err)
    }

    visibility, err := kernel.BuildVisibility(g)
    if err != nil {
        t.Fatalf("BuildVisibility failed: %v", err)
    }

    if visibility["A"][token] {
        t.Fatalf("did not expect token visible in A without exports")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./modkit/kernel -run TestGraph`  
Expected: FAIL because `BuildVisibility` does not exist and transitive re-exports not implemented.

**Step 3: Commit**

```bash
git add modkit/kernel/graph_test.go
git commit -m "test: add transitive re-export test"
```

---

### Task 3: Implement visibility building and re-export validation

**Files:**
- Modify: `modkit/kernel/visibility.go`
- Modify: `modkit/kernel/errors.go`

**Step 1: Implement minimal code**

- Export `BuildVisibility` that wraps the existing `buildVisibility` to allow tests to call it.
- Update visibility logic to allow re-exports of imported tokens.
- Add error for ambiguous re-exports (new error type in `errors.go`).

```go
// errors.go
var ErrExportAmbiguous = errors.New("export token is ambiguous across imports")

type ExportAmbiguousError struct {
    Module string
    Token  module.Token
    Imports []string
}

func (e *ExportAmbiguousError) Error() string {
    return fmt.Sprintf("export token %q in module %q is exported by multiple imports: %v", e.Token, e.Module, e.Imports)
}

func (e *ExportAmbiguousError) Unwrap() error { return ErrExportAmbiguous }
```

```go
// visibility.go
func BuildVisibility(graph *Graph) (Visibility, error) {
    return buildVisibility(graph)
}
```

Update the export validation loop to:
- Track which imports export each token.
- If a token is exported by multiple imports and the current module attempts to re-export it, return `ExportAmbiguousError`.

**Step 2: Run tests to verify they pass**

Run: `go test ./modkit/kernel -run TestVisibility`  
Expected: PASS

Run: `go test ./modkit/kernel -run TestGraph`  
Expected: PASS

**Step 3: Commit**

```bash
git add modkit/kernel/visibility.go modkit/kernel/errors.go
git commit -m "feat: support module re-exports in visibility"
```

---

### Task 4: Refactor and ensure visibility tests compile

**Files:**
- Modify: `modkit/kernel/visibility_test.go`

**Step 1: Ensure tests use helper builders and imports**

- Add any missing helper functions or imports (e.g., `buildNoop`) to tests.
- Keep tests minimal and table-driven where appropriate.

**Step 2: Run full kernel tests**

Run: `go test ./modkit/kernel`  
Expected: PASS

**Step 3: Commit**

```bash
git add modkit/kernel/visibility_test.go
git commit -m "test: finalize re-export visibility coverage"
```

---

## Execution

Plan complete and saved to `docs/plans/2026-02-05-module-reexporting-implementation.md`. Two execution options:

1. Subagent-Driven (this session) - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. Parallel Session (separate) - Open new session with executing-plans, batch execution with checkpoints

Which approach?
