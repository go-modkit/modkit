package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	modkithttp "github.com/go-modkit/modkit/modkit/http"
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
		AllowedHeaders: []string{"Content-Type", "Authorization"},
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
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization" {
		t.Fatalf("expected allow headers header to be set, got %q", got)
	}
}

func TestCORS_PreflightShortCircuits(t *testing.T) {
	cors := NewCORS(CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	})

	called := false
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("expected preflight to short-circuit without calling next")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected Access-Control-Allow-Origin header, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST" {
		t.Fatalf("expected Access-Control-Allow-Methods header, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("expected Access-Control-Allow-Headers header, got %q", got)
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
	minDuration := 10 * time.Millisecond

	handler := timing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(minDuration)
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if len(logger.infoMessages) != 1 {
		t.Fatalf("expected 1 info log, got %d", len(logger.infoMessages))
	}
	if logger.infoMessages[0] != "http.request.duration" {
		t.Fatalf("expected log message %q, got %q", "http.request.duration", logger.infoMessages[0])
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

	if len(attributes) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(attributes))
	}
	if attributes["metric"] != "http.request.duration" {
		t.Fatalf("expected metric %q, got %v", "http.request.duration", attributes["metric"])
	}
	duration, ok := attributes["duration"].(time.Duration)
	if !ok {
		t.Fatalf("expected duration to be time.Duration, got %T", attributes["duration"])
	}
	if duration < minDuration {
		t.Fatalf("expected duration to be >= %v, got %v", minDuration, duration)
	}
}

func TestGroupScope_OnlyAppliesMiddlewareToGroup(t *testing.T) {
	logger := &testLogger{}
	timing := NewTiming(logger)

	router := modkithttp.NewRouter()
	root := modkithttp.AsRouter(router)

	router.Get("/docs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	root.Group("/api/v1", func(r modkithttp.Router) {
		r.Use(timing)
		r.Handle(http.MethodGet, "/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	})

	docsRec := httptest.NewRecorder()
	docsReq := httptest.NewRequest(http.MethodGet, "/docs", nil)
	router.ServeHTTP(docsRec, docsReq)
	if docsRec.Code != http.StatusOK {
		t.Fatalf("expected docs status 200, got %d", docsRec.Code)
	}
	if len(logger.infoMessages) != 0 {
		t.Fatalf("expected no timing logs for /docs, got %d", len(logger.infoMessages))
	}

	apiRec := httptest.NewRecorder()
	apiReq := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	router.ServeHTTP(apiRec, apiReq)
	if apiRec.Code != http.StatusOK {
		t.Fatalf("expected api status 200, got %d", apiRec.Code)
	}
	if len(logger.infoMessages) != 1 {
		t.Fatalf("expected timing log for /api/v1, got %d", len(logger.infoMessages))
	}
}

func TestMiddlewareOrdering_CORSBeforeRateLimit(t *testing.T) {
	cors := NewCORS(CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	})
	limiter := NewRateLimit(RateLimitConfig{
		RequestsPerSecond: 1,
		Burst:             1,
	})

	handler := cors(limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected Access-Control-Allow-Origin header, got %q", got)
	}
}
