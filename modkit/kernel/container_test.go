package kernel_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

type recordingCloser struct {
	name   string
	closed *[]string
}

func (c *recordingCloser) Close() error {
	*c.closed = append(*c.closed, c.name)
	return nil
}

type erroringCloser struct {
	name   string
	closed *[]string
	err    error
}

func (c *erroringCloser) Close() error {
	*c.closed = append(*c.closed, c.name)
	return c.err
}

type testCloser interface {
	Close() error
}

func newTestAppWithClosers(t *testing.T, closers ...testCloser) *kernel.App {
	t.Helper()

	providers := make([]module.ProviderDef, 0, len(closers))
	tokens := make([]module.Token, 0, len(closers))

	for i, closer := range closers {
		token := module.Token(fmt.Sprintf("closer.%d", i))
		c := closer
		providers = append(providers, module.ProviderDef{
			Token: token,
			Build: func(r module.Resolver) (any, error) {
				return c, nil
			},
		})
		tokens = append(tokens, token)
	}

	modA := mod("A", nil, providers, nil, nil)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	for _, token := range tokens {
		if _, err := app.Get(token); err != nil {
			t.Fatalf("Get %s failed: %v", token, err)
		}
	}

	return app
}

func TestAppGetRejectsNotVisibleToken(t *testing.T) {
	modA := mod("A", nil, nil, nil, nil)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	_, err = app.Get(module.Token("missing"))
	if err == nil {
		t.Fatalf("expected missing token error")
	}

	var notVisible *kernel.TokenNotVisibleError
	if !errors.As(err, &notVisible) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if notVisible.Token != module.Token("missing") {
		t.Fatalf("unexpected token: %q", notVisible.Token)
	}
}

func TestContainerSingletonConcurrent(t *testing.T) {
	shared := module.Token("shared")
	var calls int32
	started := make(chan struct{})
	release := make(chan struct{})

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: shared,
			Build: func(_ module.Resolver) (any, error) {
				if atomic.AddInt32(&calls, 1) == 1 {
					close(started)
				}
				<-release
				return "value", nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = app.Get(shared)
		}()
	}

	<-started
	close(release)
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected single build call, got %d", got)
	}
}

func TestContainerDetectsSelfCycle(t *testing.T) {
	self := module.Token("self")

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: self,
			Build: func(r module.Resolver) (any, error) {
				return r.Get(self)
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	_, err = app.Get(self)
	if err == nil {
		t.Fatalf("expected dependency cycle error")
	}

	var cycleErr *kernel.ProviderCycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestContainerDetectsMutualCycle(t *testing.T) {
	a := module.Token("a")
	b := module.Token("b")

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: a,
			Build: func(r module.Resolver) (any, error) {
				return r.Get(b)
			},
		}, {
			Token: b,
			Build: func(r module.Resolver) (any, error) {
				return r.Get(a)
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	_, err = app.Get(a)
	if err == nil {
		t.Fatalf("expected dependency cycle error")
	}

	var cycleErr *kernel.ProviderCycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestContainerDetectsConcurrentMutualCycle(t *testing.T) {
	a := module.Token("a")
	b := module.Token("b")
	start := make(chan struct{})

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: a,
			Build: func(r module.Resolver) (any, error) {
				<-start
				return r.Get(b)
			},
		}, {
			Token: b,
			Build: func(r module.Resolver) (any, error) {
				<-start
				return r.Get(a)
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	results := make(chan error, 2)
	go func() {
		_, err := app.Get(a)
		results <- err
	}()
	go func() {
		_, err := app.Get(b)
		results <- err
	}()

	close(start)

	deadline := time.After(2 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case err := <-results:
			if err == nil {
				t.Fatalf("expected dependency cycle error")
			}
			var cycleErr *kernel.ProviderCycleError
			if !errors.As(err, &cycleErr) {
				t.Fatalf("unexpected error type: %T", err)
			}
		case <-deadline:
			t.Fatalf("timeout waiting for cycle detection")
		}
	}
}

// TestContainerGetWrapsProviderBuildError verifies that Container.Get wraps
// provider build errors in ProviderBuildError and preserves the original error.
// This tests the error wrapping path when a provider's build function fails.
func TestContainerGetWrapsProviderBuildError(t *testing.T) {
	badToken := module.Token("bad")
	sentinel := errors.New("build failed sentinel")

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: badToken,
			Build: func(_ module.Resolver) (any, error) {
				return nil, sentinel
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	// Trigger the build error by requesting the token
	_, err = app.Get(badToken)
	if err == nil {
		t.Fatalf("expected error for build failure")
	}

	var buildErr *kernel.ProviderBuildError
	if !errors.As(err, &buildErr) {
		t.Fatalf("unexpected error type: %T, wanted ProviderBuildError", err)
	}

	if buildErr.Token != badToken {
		t.Fatalf("expected Token %q, got %q", badToken, buildErr.Token)
	}
	if buildErr.Module != "A" {
		t.Fatalf("expected Module %q, got %q", "A", buildErr.Module)
	}

	// Verify the original error is preserved in the error chain
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected error chain to include sentinel, got: %v", err)
	}
}

// TestContainerGetMissingTokenError verifies that Container.Get reports errors
// correctly when a requested token is not found. We test this by creating a
// non-root module and bypassing it via module imports.
func TestContainerGetMissingTokenError(t *testing.T) {
	modB := mod("B", nil, nil, nil, nil)
	modA := mod("A", []module.Module{modB}, nil, nil, nil)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	missingToken := module.Token("missing")
	// Try to get a missing token from a non-root module context
	// This should trigger TokenNotVisibleError for a token outside visibility
	resolver := app.Resolver()
	_, err = resolver.Get(missingToken)
	if err == nil {
		t.Fatalf("expected error for missing token")
	}

	var notVisible *kernel.TokenNotVisibleError
	if !errors.As(err, &notVisible) {
		t.Fatalf("expected TokenNotVisibleError, got %T", err)
	}
	if notVisible.Module != "A" || notVisible.Token != missingToken {
		t.Fatalf("unexpected error fields: %+v", notVisible)
	}
}

// TestContainerGetSingletonBehavior verifies that Container.Get caches instances
// and reuses them on subsequent calls, demonstrating singleton semantics.
func TestContainerGetSingletonBehavior(t *testing.T) {
	token := module.Token("cached")
	var buildCount int32
	type cachedInstance struct{}

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: token,
			Build: func(_ module.Resolver) (any, error) {
				buildCount++
				return &cachedInstance{}, nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	// First call should build
	val1, err := app.Get(token)
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}
	instance1, ok := val1.(*cachedInstance)
	if !ok {
		t.Fatalf("expected *cachedInstance, got %T", val1)
	}
	if buildCount != 1 {
		t.Fatalf("expected 1 build call, got %d", buildCount)
	}

	// Second call should use cache
	val2, err := app.Get(token)
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}
	instance2, ok := val2.(*cachedInstance)
	if !ok {
		t.Fatalf("expected *cachedInstance, got %T", val2)
	}
	if buildCount != 1 {
		t.Fatalf("expected still 1 build call (cached), got %d", buildCount)
	}

	// Both should be the same instance
	if instance1 != instance2 {
		t.Fatalf("expected same cached instance")
	}
}

// TestContainerGetRegistersCleanupHooks verifies that Container.Get registers
// cleanup hooks when a provider has a cleanup function.
func TestContainerGetRegistersCleanupHooks(t *testing.T) {
	token := module.Token("with.cleanup")
	var cleanupCalled bool

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: token,
			Build: func(_ module.Resolver) (any, error) {
				return "instance", nil
			},
			Cleanup: func(_ context.Context) error {
				cleanupCalled = true
				return nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	_, err = app.Get(token)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	hooks := app.CleanupHooks()
	if len(hooks) != 1 {
		t.Fatalf("expected 1 cleanup hook, got %d", len(hooks))
	}

	err = hooks[0](context.Background())
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	if !cleanupCalled {
		t.Fatalf("expected cleanup to be called")
	}
}

func TestAppCloseReverseOrder(t *testing.T) {
	var closed []string
	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: "closer.a",
			Build: func(_ module.Resolver) (any, error) {
				return &recordingCloser{name: "a", closed: &closed}, nil
			},
		}, {
			Token: "closer.b",
			Build: func(_ module.Resolver) (any, error) {
				return &recordingCloser{name: "b", closed: &closed}, nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if _, err := app.Get("closer.a"); err != nil {
		t.Fatalf("Get closer.a failed: %v", err)
	}
	if _, err := app.Get("closer.b"); err != nil {
		t.Fatalf("Get closer.b failed: %v", err)
	}

	if err := app.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if len(closed) != 2 || closed[0] != "b" || closed[1] != "a" {
		t.Fatalf("expected reverse close order, got %v", closed)
	}
}

func TestAppCloseOrderWithDependencies(t *testing.T) {
	var closed []string

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: "closer.a",
			Build: func(_ module.Resolver) (any, error) {
				return &recordingCloser{name: "a", closed: &closed}, nil
			},
		}, {
			Token: "closer.b",
			Build: func(r module.Resolver) (any, error) {
				if _, err := r.Get("closer.a"); err != nil {
					return nil, err
				}
				return &recordingCloser{name: "b", closed: &closed}, nil
			},
		}, {
			Token: "closer.c",
			Build: func(r module.Resolver) (any, error) {
				if _, err := r.Get("closer.b"); err != nil {
					return nil, err
				}
				return &recordingCloser{name: "c", closed: &closed}, nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	_, _ = app.Get("closer.c")

	if err := app.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	want := []string{"c", "b", "a"}
	if !reflect.DeepEqual(closed, want) {
		t.Fatalf("expected %v, got %v", want, closed)
	}
}

func TestAppCloseContinuesAfterError(t *testing.T) {
	var closed []string
	errB := errors.New("close failed")
	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: "closer.a",
			Build: func(_ module.Resolver) (any, error) {
				return &recordingCloser{name: "a", closed: &closed}, nil
			},
		}, {
			Token: "closer.b",
			Build: func(_ module.Resolver) (any, error) {
				return &erroringCloser{name: "b", closed: &closed, err: errB}, nil
			},
		}, {
			Token: "closer.c",
			Build: func(_ module.Resolver) (any, error) {
				return &recordingCloser{name: "c", closed: &closed}, nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if _, err := app.Get("closer.a"); err != nil {
		t.Fatalf("Get closer.a failed: %v", err)
	}
	if _, err := app.Get("closer.b"); err != nil {
		t.Fatalf("Get closer.b failed: %v", err)
	}
	_, _ = app.Get("closer.c")

	if err := app.Close(); !errors.Is(err, errB) {
		t.Fatalf("expected error %v, got %v", errB, err)
	}

	if len(closed) != 3 || closed[0] != "c" || closed[1] != "b" || closed[2] != "a" {
		t.Fatalf("expected reverse close order with all closers, got %v", closed)
	}
}

func TestAppCloseAggregatesErrors(t *testing.T) {
	var closed []string
	errA := errors.New("close a failed")
	errB := errors.New("close b failed")

	app := newTestAppWithClosers(
		t,
		&erroringCloser{name: "a", closed: &closed, err: errA},
		&erroringCloser{name: "b", closed: &closed, err: errB},
	)

	err := app.Close()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, errA) || !errors.Is(err, errB) {
		t.Fatalf("expected aggregated error to include both errA and errB")
	}
}

func TestAppCloseIsIdempotent(t *testing.T) {
	var closed []string
	app := newTestAppWithClosers(
		t,
		&recordingCloser{name: "a", closed: &closed},
	)

	if err := app.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := app.Close(); err != nil {
		t.Fatalf("expected nil error on second close, got %v", err)
	}

	if got := len(closed); got != 1 {
		t.Fatalf("expected 1 close call, got %d", got)
	}
}
