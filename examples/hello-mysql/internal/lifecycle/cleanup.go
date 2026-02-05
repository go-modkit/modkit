package lifecycle

import (
	"context"
	"errors"
)

// CleanupHook defines a shutdown cleanup function.
type CleanupHook func(ctx context.Context) error

// RunCleanup executes hooks in LIFO order and returns any combined errors.
func RunCleanup(ctx context.Context, hooks []CleanupHook) error {
	var joined error
	for i := len(hooks) - 1; i >= 0; i-- {
		if hooks[i] == nil {
			continue
		}
		if err := hooks[i](ctx); err != nil {
			joined = errors.Join(joined, err)
		}
	}
	return joined
}

// FromFuncs wraps raw cleanup functions into CleanupHook values.
func FromFuncs(funcs []func(context.Context) error) []CleanupHook {
	if len(funcs) == 0 {
		return nil
	}
	hooks := make([]CleanupHook, len(funcs))
	for i, fn := range funcs {
		if fn == nil {
			continue
		}
		hooks[i] = CleanupHook(fn)
	}
	return hooks
}

type shutdowner interface {
	Shutdown(ctx context.Context) error
}

// ShutdownServer shuts down the server, then runs cleanup hooks.
func ShutdownServer(ctx context.Context, server shutdowner, hooks []CleanupHook) error {
	shutdownErr := server.Shutdown(ctx)
	cleanupErr := RunCleanup(ctx, hooks)
	if shutdownErr != nil {
		return shutdownErr
	}
	return cleanupErr
}
