// Package testkit provides a canonical harness for module-level tests.
package testkit

import (
	"fmt"
)

// ControllerNotFoundError is returned when a controller key is missing in the harness app.
type ControllerNotFoundError struct {
	Module string
	Name   string
}

func (e *ControllerNotFoundError) Error() string {
	return fmt.Sprintf("controller not found: module=%q name=%q", e.Module, e.Name)
}

// TypeAssertionError is returned when a typed helper cannot assert the requested type.
type TypeAssertionError struct {
	Target  string
	Actual  string
	Context string
}

func (e *TypeAssertionError) Error() string {
	return fmt.Sprintf("type assertion failed: target=%s actual=%s context=%s", e.Target, e.Actual, e.Context)
}

// HarnessCloseError aggregates cleanup hook and app close errors.
type HarnessCloseError struct {
	HookErr  error
	CloseErr error
}

func (e *HarnessCloseError) Error() string {
	if e.HookErr != nil && e.CloseErr != nil {
		return fmt.Sprintf("harness close failed: hooks=%v; close=%v", e.HookErr, e.CloseErr)
	}
	if e.HookErr != nil {
		return fmt.Sprintf("harness close failed: hooks=%v", e.HookErr)
	}
	if e.CloseErr != nil {
		return fmt.Sprintf("harness close failed: close=%v", e.CloseErr)
	}
	return "harness close failed"
}

// Unwrap returns both hook and close errors for errors.Is/errors.As matching.
func (e *HarnessCloseError) Unwrap() []error {
	errs := make([]error, 0, 2)
	if e.HookErr != nil {
		errs = append(errs, e.HookErr)
	}
	if e.CloseErr != nil {
		errs = append(errs, e.CloseErr)
	}
	return errs
}

// NilOptionError is returned when a nil testkit option is passed.
type NilOptionError struct {
	Index int
}

func (e *NilOptionError) Error() string {
	return fmt.Sprintf("nil testkit option: index=%d", e.Index)
}
