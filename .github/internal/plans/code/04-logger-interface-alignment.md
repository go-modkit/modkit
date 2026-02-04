# C4: Align Logger Interface with Documentation

**Status:** 🔴 Not started  
**Type:** Code change  
**Priority:** Medium

---

## Motivation

The `logging.Logger` interface implementation doesn't match documentation:

| Aspect | Documented (api.md) | Actual (logger.go) |
|--------|---------------------|-------------------|
| Method signature | `Debug(msg string, args ...any)` | `Debug(msg string, attrs ...slog.Attr)` |
| Warn method | `Warn(msg string, args ...any)` | **Missing** |
| With signature | `With(args ...any) Logger` | `With(attrs ...slog.Attr) Logger` |

The documented `...any` signature is more ergonomic and matches `slog.Logger` patterns. The current `...slog.Attr` signature is more restrictive and less user-friendly.

---

## Assumptions

1. Changing the interface is acceptable since the project is pre-v0.1.0
2. The `...any` pattern aligns better with slog's variadic key-value approach
3. Adding `Warn` method provides feature parity with standard log levels

---

## Requirements

### R1: Update Logger interface to use ...any

Change method signatures to accept variadic `any` like slog does.

### R2: Add Warn method

Include `Warn(msg string, args ...any)` for feature parity.

### R3: Update slog adapter implementation

The `slogAdapter` must convert `...any` args to slog calls.

### R4: Update nop logger

The `nopLogger` must implement the new interface.

### R5: Update RequestLogger middleware

If it uses the Logger interface, ensure it works with new signature.

---

## Files to Modify

| File | Change |
|------|--------|
| `modkit/logging/logger.go` | Update interface definition |
| `modkit/logging/slog.go` | Update adapter implementation |
| `modkit/logging/nop.go` | Update nop implementation |
| `modkit/logging/logger_test.go` | Update/add tests |
| `modkit/http/logging.go` | Verify compatibility |

---

## Implementation

### Step 1: Update logger.go

Change:

```go
type Logger interface {
    Debug(msg string, attrs ...slog.Attr)
    Info(msg string, attrs ...slog.Attr)
    Error(msg string, attrs ...slog.Attr)
    With(attrs ...slog.Attr) Logger
}
```

To:

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    With(args ...any) Logger
}
```

### Step 2: Update slog.go

Change adapter to pass args directly to slog (which accepts `...any`):

```go
type slogAdapter struct {
    logger *slog.Logger
}

func NewSlog(logger *slog.Logger) Logger {
    if logger == nil {
        return Nop()
    }
    return slogAdapter{logger: logger}
}

func (s slogAdapter) Debug(msg string, args ...any) {
    s.logger.Debug(msg, args...)
}

func (s slogAdapter) Info(msg string, args ...any) {
    s.logger.Info(msg, args...)
}

func (s slogAdapter) Warn(msg string, args ...any) {
    s.logger.Warn(msg, args...)
}

func (s slogAdapter) Error(msg string, args ...any) {
    s.logger.Error(msg, args...)
}

func (s slogAdapter) With(args ...any) Logger {
    return slogAdapter{logger: s.logger.With(args...)}
}
```

Remove the `attrsToAny` helper function as it's no longer needed.

### Step 3: Update nop.go

```go
type nopLogger struct{}

func Nop() Logger {
    return nopLogger{}
}

func (nopLogger) Debug(string, ...any) {}
func (nopLogger) Info(string, ...any)  {}
func (nopLogger) Warn(string, ...any)  {}
func (nopLogger) Error(string, ...any) {}
func (nopLogger) With(...any) Logger   { return nopLogger{} }
```

### Step 4: Update http/logging.go

Check if `RequestLogger` uses `slog.Attr` directly. If so, update to use key-value pairs:

```go
logger.Info("http request",
    "method", r.Method,
    "path", r.URL.Path,
    "status", ww.Status(),
    "duration", duration,
)
```

---

## Validation

### Unit Tests

Add/update tests in `modkit/logging/logger_test.go`:

```go
func TestSlogAdapter(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewTextHandler(&buf, nil)
    logger := NewSlog(slog.New(handler))
    
    logger.Info("test message", "key", "value", "count", 42)
    
    output := buf.String()
    if !strings.Contains(output, "test message") {
        t.Error("message not logged")
    }
    if !strings.Contains(output, "key=value") {
        t.Error("key-value not logged")
    }
}

func TestLoggerWarn(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewTextHandler(&buf, nil)
    logger := NewSlog(slog.New(handler))
    
    logger.Warn("warning message")
    
    if !strings.Contains(buf.String(), "WARN") {
        t.Error("warn level not used")
    }
}

func TestNopLogger(t *testing.T) {
    logger := Nop()
    
    // Should not panic
    logger.Debug("msg")
    logger.Info("msg")
    logger.Warn("msg")
    logger.Error("msg")
    logger.With("key", "value").Info("msg")
}
```

### Integration Test

```bash
go test ./modkit/logging/...
go test ./modkit/http/...
go test ./examples/...
```

---

## Acceptance Criteria

- [ ] `Logger` interface uses `...any` for all methods
- [ ] `Logger` interface includes `Warn(msg string, args ...any)`
- [ ] `slogAdapter` implements new interface correctly
- [ ] `nopLogger` implements new interface correctly
- [ ] `RequestLogger` middleware works with new interface
- [ ] All existing tests pass
- [ ] New tests cover `Warn` method and key-value logging
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] Interface matches `docs/reference/api.md` documentation

---

## References

- Current implementation: `modkit/logging/logger.go`, `modkit/logging/slog.go`, `modkit/logging/nop.go`
- Documentation: `docs/reference/api.md` (lines 204-214)
- slog documentation: https://pkg.go.dev/log/slog
