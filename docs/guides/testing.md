# Testing

modkit uses standard Go testing. This guide covers patterns for testing modules, providers, and controllers.

## Testing Principles

1. **Test at the right level** — Unit test business logic, integration test module wiring
2. **Use the kernel** — Bootstrap real modules in tests to verify wiring
3. **Mock at boundaries** — Replace external dependencies (DB, HTTP) with test doubles

## Unit Testing Providers

Test provider logic in isolation:

```go
func TestUsersService_Create(t *testing.T) {
    // Arrange: mock repository
    repo := &mockRepository{
        createFn: func(ctx context.Context, user User) error {
            return nil
        },
    }
    svc := NewUsersService(repo)

    // Act
    err := svc.Create(context.Background(), User{Name: "Ada"})

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

## Testing Module Wiring

Bootstrap modules to verify providers resolve correctly:

```go
func TestUsersModuleBootstrap(t *testing.T) {
    // Arrange: create module with mock DB
    db := &DatabaseModule{db: mockDB}
    users := NewUsersModule(db)

    // Act
    app, err := kernel.Bootstrap(users)

    // Assert
    if err != nil {
        t.Fatalf("bootstrap failed: %v", err)
    }
    if len(app.Controllers) != 1 {
        t.Fatalf("expected 1 controller, got %d", len(app.Controllers))
    }
}
```

## Testing Visibility

Verify modules can only access exported tokens:

```go
func TestVisibilityEnforcement(t *testing.T) {
    internal := &InternalModule{}  // has provider but doesn't export it
    consumer := &ConsumerModule{imports: internal}

    _, err := kernel.Bootstrap(consumer)

    var visErr *kernel.TokenNotVisibleError
    if !errors.As(err, &visErr) {
        t.Fatalf("expected TokenNotVisibleError, got %v", err)
    }
}
```

## Table-Driven Tests

Use table-driven tests for graph validation:

```go
func TestBuildGraphErrors(t *testing.T) {
    tests := []struct {
        name    string
        root    module.Module
        wantErr error
    }{
        {
            name:    "nil root",
            root:    nil,
            wantErr: &kernel.RootModuleNilError{},
        },
        {
            name:    "duplicate module name",
            root:    moduleWithDuplicateImport(),
            wantErr: &kernel.DuplicateModuleNameError{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := kernel.BuildGraph(tt.root)
            if err == nil {
                t.Fatal("expected error")
            }
            // Check error type matches
        })
    }
}
```

## Testing Controllers

Test controllers as HTTP handlers:

```go
func TestUsersController_List(t *testing.T) {
    // Arrange
    svc := &mockUsersService{
        listFn: func(ctx context.Context) ([]User, error) {
            return []User{{ID: 1, Name: "Ada"}}, nil
        },
    }
    controller := NewUsersController(svc)

    req := httptest.NewRequest(http.MethodGet, "/users", nil)
    rec := httptest.NewRecorder()

    // Act
    controller.List(rec, req)

    // Assert
    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}
```

## Integration Testing with Test Modules

Create test-specific modules that swap real dependencies:

```go
func TestIntegration(t *testing.T) {
    // Test database module with in-memory DB
    testDB := &TestDatabaseModule{db: setupTestDB(t)}
    
    // Real users module with test DB
    users := NewUsersModule(testDB)
    
    app, err := kernel.Bootstrap(users)
    if err != nil {
        t.Fatal(err)
    }

    router := mkhttp.NewRouter()
    mkhttp.RegisterRoutes(mkhttp.AsRouter(router), app.Controllers)

    // Test via HTTP
    req := httptest.NewRequest(http.MethodGet, "/users", nil)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}
```

## Smoke Tests with Testcontainers

For full integration tests, use testcontainers:

```go
func TestSmoke(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    ctx := context.Background()
    
    // Start MySQL container
    mysql, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "mysql:8",
            ExposedPorts: []string{"3306/tcp"},
            // ...
        },
        Started: true,
    })
    if err != nil {
        t.Fatal(err)
    }
    defer mysql.Terminate(ctx)

    // Get connection string and bootstrap
    dsn := getContainerDSN(mysql)
    app := bootstrapApp(dsn)
    
    // Run tests against real stack
}
```

See `examples/hello-mysql/internal/smoke/smoke_test.go` for a complete example.

## Running Tests

Run all tests:

```bash
go test ./...
```

Run library tests only:

```bash
go test ./modkit/...
```

Run a specific test:

```bash
go test ./modkit/kernel -run TestBuildGraph
```

Skip integration tests:

```bash
go test -short ./...
```

## Tips

- Keep unit tests fast; use mocks for external dependencies
- Use `t.Parallel()` for independent tests
- Test error cases, not just happy paths
- Integration tests should clean up after themselves
