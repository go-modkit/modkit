package testkit

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

// TB is the minimal test interface required by TestKit helpers.
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
	Cleanup(func())
}

// Harness wraps a bootstrapped app for test ergonomics.
type Harness struct {
	app              *kernel.App
	closeMu          sync.Mutex
	hooksInitialized bool
	hooks            []func(context.Context) error
	nextHook         int
	hookErrs         []error
	closed           bool
	closeErr         error
}

// New bootstraps a test harness and fails the test if bootstrap fails.
func New(tb TB, root module.Module, opts ...Option) *Harness {
	tb.Helper()

	h, err := NewE(tb, root, opts...)
	if err != nil {
		tb.Fatalf("testkit.New bootstrap failed: %v", err)
		return nil
	}

	return h
}

// NewE bootstraps a test harness and returns bootstrap errors.
func NewE(tb TB, root module.Module, opts ...Option) (*Harness, error) {
	if tb != nil {
		tb.Helper()
	}

	cfg := defaultConfig()
	for idx, opt := range opts {
		if opt == nil {
			return nil, &NilOptionError{Index: idx}
		}
		opt.apply(&cfg)
	}

	kernelOverrides := make([]kernel.ProviderOverride, 0, len(cfg.overrides))
	for _, override := range cfg.overrides {
		kernelOverrides = append(kernelOverrides, kernel.ProviderOverride{
			Token:   override.Token,
			Build:   override.Build,
			Cleanup: override.Cleanup,
		})
	}

	app, err := kernel.BootstrapWithOptions(root, kernel.WithProviderOverrides(kernelOverrides...))
	if err != nil {
		return nil, err
	}

	h := &Harness{app: app}
	if cfg.autoClose && tb != nil {
		tb.Cleanup(func() {
			_ = h.Close()
		})
	}

	return h, nil
}

// App returns the underlying bootstrapped app.
func (h *Harness) App() *kernel.App {
	return h.app
}

// Close closes cleanup hooks and closers with background context.
func (h *Harness) Close() error {
	return h.CloseContext(context.Background())
}

// CloseContext closes cleanup hooks then app closers. Calls are idempotent.
func (h *Harness) CloseContext(ctx context.Context) error {
	h.closeMu.Lock()
	defer h.closeMu.Unlock()

	if h.closed {
		return h.closeErr
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if !h.hooksInitialized {
		h.hooks = h.app.CleanupHooks()
		h.hooksInitialized = true
	}

	for h.nextHook < len(h.hooks) {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := h.hooks[h.nextHook](ctx); err != nil {
			h.hookErrs = append(h.hookErrs, err)
		}
		h.nextHook++
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	closeErr := h.app.CloseContext(ctx)
	if errors.Is(closeErr, context.Canceled) || errors.Is(closeErr, context.DeadlineExceeded) {
		return closeErr
	}

	hookErr := joinErrors(h.hookErrs)
	if hookErr != nil || closeErr != nil {
		h.closeErr = &HarnessCloseError{HookErr: hookErr, CloseErr: closeErr}
		h.closed = true
		return h.closeErr
	}

	h.closed = true
	h.closeErr = nil
	return nil
}

// Get resolves a typed token or fails the test.
func Get[T any](tb TB, h *Harness, token module.Token) T {
	tb.Helper()

	v, err := GetE[T](h, token)
	if err != nil {
		tb.Fatalf("testkit.Get[%v] failed for token %q: %v", reflect.TypeFor[T](), token, err)
		var zero T
		return zero
	}

	return v
}

// GetE resolves a typed token.
func GetE[T any](h *Harness, token module.Token) (T, error) {
	var zero T
	val, err := h.app.Get(token)
	if err != nil {
		return zero, err
	}

	typed, ok := val.(T)
	if !ok {
		targetType := reflect.TypeFor[T]().String()
		return zero, &TypeAssertionError{
			Target:  targetType,
			Actual:  fmt.Sprintf("%T", val),
			Context: fmt.Sprintf("token=%q", token),
		}
	}

	return typed, nil
}

// Controller returns a typed controller or fails the test.
func Controller[T any](tb TB, h *Harness, moduleName, controllerName string) T {
	tb.Helper()

	v, err := ControllerE[T](h, moduleName, controllerName)
	if err != nil {
		tb.Fatalf("testkit.Controller[%v] failed for %s:%s: %v", reflect.TypeFor[T](), moduleName, controllerName, err)
		var zero T
		return zero
	}

	return v
}

// ControllerE returns a typed controller instance.
func ControllerE[T any](h *Harness, moduleName, controllerName string) (T, error) {
	var zero T
	key := moduleName + ":" + controllerName
	val, ok := h.app.Controllers[key]
	if !ok {
		return zero, &ControllerNotFoundError{Module: moduleName, Name: controllerName}
	}

	typed, ok := val.(T)
	if !ok {
		targetType := reflect.TypeFor[T]().String()
		return zero, &TypeAssertionError{
			Target:  targetType,
			Actual:  fmt.Sprintf("%T", val),
			Context: fmt.Sprintf("controller=%s:%s", moduleName, controllerName),
		}
	}

	return typed, nil
}

func joinErrors(errs []error) error {
	var nonNil []error
	for _, err := range errs {
		if err != nil {
			nonNil = append(nonNil, err)
		}
	}

	if len(nonNil) == 0 {
		return nil
	}

	return fmt.Errorf("cleanup hook failures: %w", errors.Join(nonNil...))
}
