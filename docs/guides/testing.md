# Testing

modkit uses standard Go testing. Focus on deterministic graph construction, visibility, and provider/controller resolution.

## Suggested Focus Areas

- Module definition validation and compile-time expectations.
- Graph construction errors, including cycles and duplicate module names.
- Visibility enforcement and export rules.
- Provider resolution and cycle detection in the container.
- Controller build failures surfaced during bootstrap.

## Table-Driven Style

Table-driven tests work well for graph validation and error cases.

```go
func TestBuildGraphRejectsInvalidModules(t *testing.T) {
    tests := []struct {
        name string
        root module.Module
        wantErr any
    }{
        {name: "nil root", root: nil, wantErr: &kernel.RootModuleNilError{}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := kernel.BuildGraph(tt.root)
            if err == nil {
                t.Fatalf("expected error")
            }
        })
    }
}
```

## Running Tests

Run all tests:

```bash
go test ./...
```

Run a single package:

```bash
go test ./modkit/kernel -run TestBuildGraphRejectsInvalidModules
```

Run only library tests:

```bash
go test ./modkit/...
```
