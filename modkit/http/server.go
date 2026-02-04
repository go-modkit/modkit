package http

import "net/http"

// Serve starts an HTTP server on the given address using the provided handler.
func Serve(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}
