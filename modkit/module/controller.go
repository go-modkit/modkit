// Package module defines the core types for modkit's module system.
package module

// ControllerDef describes how to build a controller instance.
type ControllerDef struct {
	Name  string
	Build func(r Resolver) (any, error)
}
