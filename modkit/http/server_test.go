package http

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestServe_UsesHTTPServer(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(addr, router)
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	var resp *http.Response
	for i := 0; i < 25; i++ {
		resp, err = client.Get("http://" + addr + "/ping")
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	select {
	case serveErr := <-errCh:
		t.Fatalf("serve returned early: %v", serveErr)
	default:
	}
}
