package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-modkit/modkit/modkit/logging"
)

type captureLogger struct{ messages []string }

func (c *captureLogger) Debug(msg string, _ ...any) { c.messages = append(c.messages, msg) }
func (c *captureLogger) Info(msg string, _ ...any)  { c.messages = append(c.messages, msg) }
func (c *captureLogger) Warn(msg string, _ ...any)  { c.messages = append(c.messages, msg) }
func (c *captureLogger) Error(msg string, _ ...any) { c.messages = append(c.messages, msg) }
func (c *captureLogger) With(...any) logging.Logger { return c }

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
