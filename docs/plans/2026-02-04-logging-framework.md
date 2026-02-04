# Logging Hooks Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add framework-level logging hooks (interfaces + HTTP request logging middleware) and wire the example app to use them.

**Architecture:** Introduce a small `modkit/logging` package with a minimal interface plus adapters. Extend `modkit/http` with a request-logging middleware that depends on the interface. Update the example app to create a concrete logger and register the middleware with the router.

**Tech Stack:** Go 1.22+, `log/slog`, chi router, existing `modkit/http`.

---

### Task 1: Add logging interface + adapters in framework

**Files:**
- Create: `modkit/logging/logger.go`
- Create: `modkit/logging/nop.go`
- Create: `modkit/logging/slog.go`
- Create: `modkit/logging/logger_test.go`

**Step 1: Write the failing test**

```go
package logging

import (
	"log/slog"
	"testing"
)

type captureHandler struct{
	records []slog.Record
}

func (h *captureHandler) Enabled(_ slog.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ slog.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler     { return h }

func TestSlogAdapter_EmitsRecords(t *testing.T) {
	ch := &captureHandler{}
	base := slog.New(ch)
	logger := NewSlog(base)

	logger.Info("hello", slog.String("k", "v"))

	if len(ch.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(ch.records))
	}
	if ch.records[0].Message != "hello" {
		t.Fatalf("unexpected message: %s", ch.records[0].Message)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./modkit/logging -run TestSlogAdapter_EmitsRecords`
Expected: FAIL with "undefined: NewSlog" (or missing package).

**Step 3: Write minimal implementation**

```go
package logging

import "log/slog"

type Logger interface {
	Debug(msg string, attrs ...slog.Attr)
	Info(msg string, attrs ...slog.Attr)
	Error(msg string, attrs ...slog.Attr)
	With(attrs ...slog.Attr) Logger
}
```

```go
package logging

import "log/slog"

type nopLogger struct{}

func Nop() Logger {
	return nopLogger{}
}

func (nopLogger) Debug(string, ...slog.Attr) {}
func (nopLogger) Info(string, ...slog.Attr)  {}
func (nopLogger) Error(string, ...slog.Attr) {}
func (nopLogger) With(...slog.Attr) Logger  { return nopLogger{} }
```

```go
package logging

import "log/slog"

type slogAdapter struct {
	logger *slog.Logger
}

func NewSlog(logger *slog.Logger) Logger {
	if logger == nil {
		return Nop()
	}
	return slogAdapter{logger: logger}
}

func (s slogAdapter) Debug(msg string, attrs ...slog.Attr) {
	s.logger.Debug(msg, attrs...)
}

func (s slogAdapter) Info(msg string, attrs ...slog.Attr) {
	s.logger.Info(msg, attrs...)
}

func (s slogAdapter) Error(msg string, attrs ...slog.Attr) {
	s.logger.Error(msg, attrs...)
}

func (s slogAdapter) With(attrs ...slog.Attr) Logger {
	return slogAdapter{logger: s.logger.With(attrs...)}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./modkit/logging -run TestSlogAdapter_EmitsRecords`
Expected: PASS

**Step 5: Commit**

```bash
git add modkit/logging
git commit -m "feat: add logging interface and slog adapter"
```

---

### Task 2: Add HTTP request logging middleware

**Files:**
- Create: `modkit/http/logging.go`
- Create: `modkit/http/logging_test.go`

**Note:** `modkit/http/router.go` may not require changes if the new middleware can be added in a new file.

**Step 1: Write the failing test**

```go
package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"log/slog"

	"github.com/aryeko/modkit/modkit/logging"
)

type captureLogger struct{ messages []string }

func (c *captureLogger) Debug(msg string, _ ...slog.Attr) { c.messages = append(c.messages, msg) }
func (c *captureLogger) Info(msg string, _ ...slog.Attr)  { c.messages = append(c.messages, msg) }
func (c *captureLogger) Error(msg string, _ ...slog.Attr) { c.messages = append(c.messages, msg) }
func (c *captureLogger) With(...slog.Attr) logging.Logger { return c }

func TestRequestLogger_LogsRequests(t *testing.T) {
	logger := &captureLogger{}
	router := NewRouter()
	router.Use(RequestLogger(logger))
	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(rec, req)

	if len(logger.messages) == 0 {
		t.Fatalf("expected log message")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./modkit/http -run TestRequestLogger_LogsRequests`
Expected: FAIL with "undefined: RequestLogger".

**Step 3: Write minimal implementation**

```go
package http

import (
	"net/http"
	"time"

	"github.com/aryeko/modkit/modkit/logging"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

func RequestLogger(logger logging.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = logging.Nop()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Duration("duration", duration),
			)
		})
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./modkit/http -run TestRequestLogger_LogsRequests`
Expected: PASS

**Step 5: Commit**

```bash
git add modkit/http
git commit -m "feat: add http request logging middleware"
```

---

### Task 3: Wire logging into the example app

**Files:**
- Modify: `examples/hello-mysql/internal/platform/logging/log.go`
- Modify: `examples/hello-mysql/internal/httpserver/server.go`
- Modify: `examples/hello-mysql/README.md`
- Test: `examples/hello-mysql/internal/httpserver/server_test.go` (new if needed)

**Step 1: Write the failing test**

```go
package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aryeko/modkit/examples/hello-mysql/internal/modules/app"
)

func TestBuildHandler_Health(t *testing.T) {
	h, err := BuildHandler(app.Options{HTTPAddr: ":8080", MySQLDSN: "root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true"})
	if err != nil {
		t.Fatalf("build handler: %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./examples/hello-mysql/internal/httpserver -run TestBuildHandler_Health`
Expected: FAIL only if logging wiring introduces breaking change.

**Step 3: Write minimal implementation**

- Update `internal/platform/logging` to return a `logging.Logger` backed by `slog.New(slog.NewJSONHandler(os.Stdout, nil))`.
- In `BuildHandler`, create logger and register middleware:

```go
logger := logging.New()
router := modkithttp.NewRouter()
router.Use(modkithttp.RequestLogger(logger))
```

**Step 4: Run test to verify it passes**

Run: `go test ./examples/hello-mysql/internal/httpserver -run TestBuildHandler_Health`
Expected: PASS

**Step 5: Commit**

```bash
git add examples/hello-mysql/internal/platform/logging/log.go examples/hello-mysql/internal/httpserver/server.go examples/hello-mysql/README.md

git commit -m "feat: add logging to hello-mysql example"
```

---

### Task 4: Full validation

**Files:**
- None

**Step 1: Run full test suite**

Run: `go test ./...`
Expected: PASS (warnings from `go-m1cpu` are acceptable).

**Step 2: Commit any updated files**

If nothing new, skip.

---

Plan complete and saved to `docs/plans/2026-02-04-logging-framework.md`. Two execution options:

1. Subagent-Driven (this session) - I dispatch fresh subagent per task, review between tasks, fast iteration
2. Parallel Session (separate) - Open new session with executing-plans, batch execution with checkpoints

Which approach?
