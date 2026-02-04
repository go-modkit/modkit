package http

import "fmt"

// RouteRegistrationError indicates a controller does not expose route registration.
type RouteRegistrationError struct {
	Name string
}

func (e *RouteRegistrationError) Error() string {
	return fmt.Sprintf("controller does not implement RouteRegistrar: %s", e.Name)
}
