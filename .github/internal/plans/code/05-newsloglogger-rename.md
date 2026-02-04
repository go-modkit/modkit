# C5: Rename NewSlog to NewSlogLogger

**Status:** 🔴 Not started  
**Type:** Code change  
**Priority:** Low

---

## Motivation

Documentation refers to `logging.NewSlogLogger(slog.Default())` but the actual function is named `logging.NewSlog()`. This causes confusion when users copy code from docs.

The longer name `NewSlogLogger` is clearer about what it returns (a `Logger` that wraps slog).

---

## Assumptions

1. No external code depends on `NewSlog` (project is pre-v0.1.0)
2. The rename is straightforward with no behavioral changes
3. Example apps may need updating if they use this function

---

## Requirements

### R1: Rename function

Change `NewSlog` to `NewSlogLogger` in implementation.

### R2: Update all usages

Find and update any internal usages of the function.

---

## Files to Modify

| File | Change |
|------|--------|
| `modkit/logging/slog.go` | Rename `NewSlog` → `NewSlogLogger` |
| `modkit/logging/logger_test.go` | Update test calls |
| Any example files using `NewSlog` | Update calls |

---

## Implementation

### Step 1: Rename in slog.go

Change:

```go
func NewSlog(logger *slog.Logger) Logger {
```

To:

```go
func NewSlogLogger(logger *slog.Logger) Logger {
```

### Step 2: Find and update usages

Search for usages:

```bash
grep -r "NewSlog" --include="*.go" .
```

Update each occurrence from `NewSlog(` to `NewSlogLogger(`.

---

## Validation

### Compile Check

```bash
go build ./...
```

### Test

```bash
go test ./modkit/logging/...
go test ./examples/...
```

### Grep Verification

After changes, only `NewSlogLogger` should exist:

```bash
grep -r "NewSlog[^L]" --include="*.go" .
# Should return empty
```

---

## Acceptance Criteria

- [ ] Function is named `NewSlogLogger`
- [ ] All usages updated
- [ ] No references to old `NewSlog` name remain (except in git history)
- [ ] All tests pass
- [ ] Function name matches `docs/reference/api.md` documentation

---

## References

- Current implementation: `modkit/logging/slog.go` (line 9)
- Documentation: `docs/reference/api.md` (lines 218-221)
