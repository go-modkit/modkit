# C1: Controller Registry Scoping

**Status:** 🔴 Not started  
**Type:** Code change  
**Priority:** Medium

---

## Motivation

Currently, controller names must be globally unique across all modules. If two modules each define a controller named `UsersController`, bootstrap fails with `DuplicateControllerNameError`.

This is unnecessarily restrictive. In a modular architecture, different modules should be able to use the same controller names without conflict. For example:
- `admin` module with `UsersController`
- `api` module with `UsersController`

The fix is to namespace controller registry keys by module name (e.g., `admin:UsersController`, `api:UsersController`), while still enforcing uniqueness within a single module.

---

## Assumptions

1. Controller name uniqueness only matters within a module, not globally
2. The HTTP adapter (`RegisterRoutes`) doesn't depend on specific controller key format
3. Changing the key format is a non-breaking change since `App.Controllers` is `map[string]any`
4. Existing code that accesses controllers by name may need updating (check examples)

---

## Requirements

### R1: Namespace controller keys

Controller registry keys should be `moduleName:controllerName` instead of just `controllerName`.

### R2: Enforce uniqueness per module only

Duplicate controller names in the same module should still error. Duplicate names across different modules should be allowed.

### R3: Update DuplicateControllerNameError

Consider adding `Module` field to the error for better debugging.

---

## Files to Modify

| File | Change |
|------|--------|
| `modkit/kernel/bootstrap.go` | Namespace keys, per-module duplicate check |
| `modkit/kernel/errors.go` | Optionally add `Module` field to `DuplicateControllerNameError` |
| `modkit/kernel/bootstrap_test.go` | Add tests for cross-module same-name controllers |

---

## Implementation

### Step 1: Add helper function in bootstrap.go

```go
func controllerKey(moduleName, controllerName string) string {
    return moduleName + ":" + controllerName
}
```

### Step 2: Update controller registration loop

Replace the current implementation in `Bootstrap()`:

```go
// Current implementation
controllers := make(map[string]any)
for _, node := range graph.Modules {
    resolver := container.resolverFor(node.Name)
    for _, controller := range node.Def.Controllers {
        if _, exists := controllers[controller.Name]; exists {
            return nil, &DuplicateControllerNameError{Name: controller.Name}
        }
        instance, err := controller.Build(resolver)
        if err != nil {
            return nil, &ControllerBuildError{Module: node.Name, Controller: controller.Name, Err: err}
        }
        controllers[controller.Name] = instance
    }
}
```

With:

```go
controllers := make(map[string]any)
perModule := make(map[string]map[string]bool)

for _, node := range graph.Modules {
    if perModule[node.Name] == nil {
        perModule[node.Name] = make(map[string]bool)
    }
    resolver := container.resolverFor(node.Name)
    for _, controller := range node.Def.Controllers {
        // Check for duplicates within the same module
        if perModule[node.Name][controller.Name] {
            return nil, &DuplicateControllerNameError{Module: node.Name, Name: controller.Name}
        }
        perModule[node.Name][controller.Name] = true

        instance, err := controller.Build(resolver)
        if err != nil {
            return nil, &ControllerBuildError{Module: node.Name, Controller: controller.Name, Err: err}
        }
        controllers[controllerKey(node.Name, controller.Name)] = instance
    }
}
```

### Step 3: Update DuplicateControllerNameError (optional enhancement)

In `errors.go`, add `Module` field:

```go
type DuplicateControllerNameError struct {
    Module string
    Name   string
}

func (e *DuplicateControllerNameError) Error() string {
    if e.Module != "" {
        return fmt.Sprintf("duplicate controller name in module %q: %s", e.Module, e.Name)
    }
    return fmt.Sprintf("duplicate controller name: %s", e.Name)
}
```

---

## Validation

### Unit Tests

Add to `modkit/kernel/bootstrap_test.go`:

```go
func TestBootstrapAllowsSameControllerNameAcrossModules(t *testing.T) {
    // Module B has controller "Shared"
    modB := &testModule{
        name: "B",
        controllers: []module.ControllerDef{{
            Name: "Shared",
            Build: func(r module.Resolver) (any, error) {
                return "controller-from-B", nil
            },
        }},
    }

    // Module A imports B and also has controller "Shared"
    modA := &testModule{
        name:    "A",
        imports: []module.Module{modB},
        controllers: []module.ControllerDef{{
            Name: "Shared",
            Build: func(r module.Resolver) (any, error) {
                return "controller-from-A", nil
            },
        }},
    }

    app, err := kernel.Bootstrap(modA)
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    // Should have 2 controllers with namespaced keys
    if len(app.Controllers) != 2 {
        t.Fatalf("expected 2 controllers, got %d", len(app.Controllers))
    }

    // Verify both controllers are accessible
    if _, ok := app.Controllers["A:Shared"]; !ok {
        t.Error("expected controller 'A:Shared' not found")
    }
    if _, ok := app.Controllers["B:Shared"]; !ok {
        t.Error("expected controller 'B:Shared' not found")
    }

    // Verify correct instances
    if app.Controllers["A:Shared"] != "controller-from-A" {
        t.Error("controller A:Shared has wrong value")
    }
    if app.Controllers["B:Shared"] != "controller-from-B" {
        t.Error("controller B:Shared has wrong value")
    }
}

func TestBootstrapRejectsDuplicateControllerInSameModule(t *testing.T) {
    mod := &testModule{
        name: "test",
        controllers: []module.ControllerDef{
            {Name: "Dup", Build: func(r module.Resolver) (any, error) { return "a", nil }},
            {Name: "Dup", Build: func(r module.Resolver) (any, error) { return "b", nil }},
        },
    }

    _, err := kernel.Bootstrap(mod)
    if err == nil {
        t.Fatal("expected error for duplicate controller name in same module")
    }

    var dupErr *kernel.DuplicateControllerNameError
    if !errors.As(err, &dupErr) {
        t.Fatalf("expected DuplicateControllerNameError, got: %T", err)
    }
    if dupErr.Name != "Dup" {
        t.Errorf("expected name 'Dup', got %q", dupErr.Name)
    }
}

func TestControllerKeyFormat(t *testing.T) {
    mod := &testModule{
        name: "users",
        controllers: []module.ControllerDef{{
            Name:  "Controller",
            Build: func(r module.Resolver) (any, error) { return nil, nil },
        }},
    }

    app, err := kernel.Bootstrap(mod)
    if err != nil {
        t.Fatal(err)
    }

    // Key should be "module:controller"
    if _, ok := app.Controllers["users:Controller"]; !ok {
        t.Errorf("expected key 'users:Controller', got keys: %v", keys(app.Controllers))
    }
}

func keys(m map[string]any) []string {
    k := make([]string, 0, len(m))
    for key := range m {
        k = append(k, key)
    }
    return k
}
```

### Integration Test

Verify HTTP adapter still works:

```bash
go test ./modkit/http/...
go test ./examples/...
```

### Check Examples

Verify example apps don't access `app.Controllers` by raw name. If they do, update them.

---

## Acceptance Criteria

- [ ] Controller keys are namespaced as `moduleName:controllerName`
- [ ] Same controller name in different modules is allowed
- [ ] Same controller name in same module is rejected with `DuplicateControllerNameError`
- [ ] `DuplicateControllerNameError` includes module name in error message
- [ ] `RegisterRoutes` works correctly (iterates all controllers regardless of key format)
- [ ] All existing tests pass
- [ ] New tests cover:
  - [ ] Cross-module same name allowed
  - [ ] Same-module duplicate rejected
  - [ ] Key format is correct
- [ ] Example apps work correctly
- [ ] `make lint` passes
- [ ] `make test` passes

---

## References

- Current implementation: `modkit/kernel/bootstrap.go` (lines 27-40)
- Error types: `modkit/kernel/errors.go` (lines 81-87)
- HTTP adapter: `modkit/http/router.go` (`RegisterRoutes` function)
- Existing bootstrap tests: `modkit/kernel/bootstrap_test.go`
