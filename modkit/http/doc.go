// Package http adapts controller instances to HTTP routing.
//
// Route registration is explicit: controllers implement RouteRegistrar and are
// invoked via RegisterRoutes. No reflection is used.
//
// NewRouter returns a chi.Router with baseline middleware. Use AsRouter to adapt
// it to the method-based Router interface.
package http
