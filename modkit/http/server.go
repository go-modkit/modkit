package http

import "net/http"

var listenAndServe = http.ListenAndServe

// Serve starts an HTTP server on the given address using the provided handler.
func Serve(addr string, handler http.Handler) error {
	return listenAndServe(addr, handler)
}
