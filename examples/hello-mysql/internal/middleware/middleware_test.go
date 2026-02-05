package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	modkitlogging "github.com/go-modkit/modkit/modkit/logging"
)

type testLogger struct {
	infoMessages []string
	infoArgs     [][]any
}

func (t *testLogger) Debug(string, ...any) {}
func (t *testLogger) Info(msg string, args ...any) {
	t.infoMessages = append(t.infoMessages, msg)
	t.infoArgs = append(t.infoArgs, args)
}
func (t *testLogger) Warn(string, ...any)  {}
func (t *testLogger) Error(string, ...any) {}
func (t *testLogger) With(...any) modkitlogging.Logger {
	return t
}

func TestCORS_AddsHeaders(t *testing.T) {
	cors := NewCORS(CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: nil,
	})

	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("expected allow origin header to be set")
	}
	if rec.Header().Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Fatalf("expected allow methods header to be set")
	}
	if _, ok := rec.Header()["Access-Control-Allow-Headers"]; !ok {
		t.Fatalf("expected allow headers header to be set")
	}
}

func TestRateLimit_BlocksAfterBurst(t *testing.T) {
	limiter := NewRateLimit(RateLimitConfig{
		RequestsPerSecond: 1,
		Burst:             2,
	})

	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("expected application/problem+json content type")
	}
	if !strings.Contains(rec.Body.String(), "rate limit exceeded") {
		t.Fatalf("expected rate limit detail message")
	}
	retryAfter := rec.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Fatalf("expected Retry-After header to be set")
	}
	seconds, err := strconv.Atoi(retryAfter)
	if err != nil || seconds < 1 {
		t.Fatalf("expected Retry-After to be a positive integer, got %q", retryAfter)
	}
}

func TestTiming_LogsDuration(t *testing.T) {
	logger := &testLogger{}
	timing := NewTiming(logger)

	handler := timing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if len(logger.infoMessages) != 1 {
		t.Fatalf("expected 1 info log, got %d", len(logger.infoMessages))
	}
	if logger.infoMessages[0] != "request timing" {
		t.Fatalf("expected log message %q, got %q", "request timing", logger.infoMessages[0])
	}

	args := logger.infoArgs[0]
	attributes := make(map[string]any, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		attributes[key] = args[i+1]
	}

	if attributes["method"] != http.MethodGet {
		t.Fatalf("expected method %q, got %v", http.MethodGet, attributes["method"])
	}
	if attributes["path"] != "/api/v1/health" {
		t.Fatalf("expected path %q, got %v", "/api/v1/health", attributes["path"])
	}
	if attributes["status"] != http.StatusCreated {
		t.Fatalf("expected status %d, got %v", http.StatusCreated, attributes["status"])
	}
	duration, ok := attributes["duration"].(time.Duration)
	if !ok {
		t.Fatalf("expected duration to be time.Duration, got %T", attributes["duration"])
	}
	if duration <= 0 {
		t.Fatalf("expected duration to be > 0, got %v", duration)
	}
}
