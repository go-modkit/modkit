# C1: Controller Registry Scoping

**Status:** 🔴 Not started  
**Type:** Code change  
**Priority:** Independent (can be done anytime)

---

## Goal

Allow controllers with the same name in different modules without collisions by namespacing controller keys at bootstrap.

## Architecture

Keep `App` as the single controller registry, but compute map keys using `moduleName + ":" + controllerName` and only enforce duplicate names within the same module. This preserves existing HTTP adapter usage while removing a global uniqueness constraint.

## Tech Stack

Go, `testing`, existing `modkit/kernel` and `modkit/http`.

---

## Task 1: Namespace controller keys in bootstrap

**Files:**
- Modify: `modkit/kernel/bootstrap.go`

### Step 1: Write the failing test

Add a new test to allow duplicate controller names across modules:

```go
func TestBootstrapAllowsSameControllerNameAcrossModules(t *testing.T) {
    modB := mod("B", nil, nil,
        []module.ControllerDef{{
            Name: "Shared",
            Build: func(r module.Resolver) (any, error) {
                return "b", nil
            },
        }},
        nil,
    )

    modA := mod("A", []module.Module{modB}, nil,
        []module.ControllerDef{{
            Name: "Shared",
            Build: func(r module.Resolver) (any, error) {
                return "a", nil
            },
        }},
        nil,
    )

    app, err := kernel.Bootstrap(modA)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if len(app.Controllers) != 2 {
        t.Fatalf("expected 2 controllers, got %d", len(app.Controllers))
    }
}
```

### Step 2: Run test to verify it fails

```bash
go test ./modkit/kernel -run TestBootstrapAllowsSameControllerNameAcrossModules
```

Expected: FAIL with `DuplicateControllerNameError`

### Step 3: Write minimal implementation

Update bootstrap to namespace controller keys and only error on duplicates within the same module:

```go
func controllerKey(moduleName, controllerName string) string {
    return moduleName + ":" + controllerName
}

// inside Bootstrap
controllers := make(map[string]any)
perModule := make(map[string]map[string]bool)
for _, node := range graph.Modules {
    if perModule[node.Name] == nil {
        perModule[node.Name] = make(map[string]bool)
    }
    resolver := container.resolverFor(node.Name)
    for _, controller := range node.Def.Controllers {
        if perModule[node.Name][controller.Name] {
            return nil, &DuplicateControllerNameError{Name: controller.Name}
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

### Step 4: Run test to verify it passes

```bash
go test ./modkit/kernel -run TestBootstrapAllowsSameControllerNameAcrossModules
```

Expected: PASS

### Step 5: Commit

```bash
git add modkit/kernel/bootstrap.go modkit/kernel/bootstrap_test.go
git commit -m "feat: namespace controller registry keys"
```

---

## Task 2: Update existing bootstrap tests

**Files:**
- Modify: `modkit/kernel/bootstrap_test.go`

### Step 1: Update the duplicate controller test

Keep the duplicate-name test but ensure it only checks duplication within the same module. No additional code change is required if it already uses the same module.

### Step 2: Run tests

```bash
go test ./modkit/kernel -run TestBootstrapRejectsDuplicateControllerNames
```

Expected: PASS

### Step 3: Commit

```bash
git add modkit/kernel/bootstrap_test.go
git commit -m "test: allow duplicate controller names across modules"
```
