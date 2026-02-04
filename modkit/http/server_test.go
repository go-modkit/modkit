package http

import (
	"errors"
	"net/http"
	"testing"
)

func TestServe_UsesHTTPServer(t *testing.T) {
	originalListen := listenAndServe
	defer func() {
		listenAndServe = originalListen
	}()

	var gotAddr string
	var gotHandler http.Handler
	listenAndServe = func(addr string, handler http.Handler) error {
		gotAddr = addr
		gotHandler = handler
		return errors.New("boom")
	}

	router := NewRouter()
	err := Serve("127.0.0.1:12345", router)

	if gotAddr != "127.0.0.1:12345" {
		t.Fatalf("expected addr %q, got %q", "127.0.0.1:12345", gotAddr)
	}
	if gotHandler != router {
		t.Fatalf("expected handler to be router")
	}
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected error from listenAndServe, got %v", err)
	}
}
