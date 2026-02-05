package kernel_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

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
			Build: func(r module.Resolver) (any, error) {
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
			Build: func(r module.Resolver) (any, error) {
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
			Build: func(r module.Resolver) (any, error) {
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
			Build: func(r module.Resolver) (any, error) {
				return "instance", nil
			},
			Cleanup: func(ctx context.Context) error {
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

func TestContainerRecordsClosersInBuildOrder(t *testing.T) {
	closerA := module.Token("closer.a")
	closerB := module.Token("closer.b")

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: closerA,
			Build: func(r module.Resolver) (any, error) {
				return &testCloser{name: "a"}, nil
			},
		}, {
			Token: closerB,
			Build: func(r module.Resolver) (any, error) {
				return &testCloser{name: "b"}, nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	_, _ = app.Get(closerA)
	_, _ = app.Get(closerB)

	closers := app.Closers()
	if len(closers) != 2 {
		t.Fatalf("expected 2 closers, got %d", len(closers))
	}

	first, ok := closers[0].(*testCloser)
	if !ok {
		t.Fatalf("expected *testCloser, got %T", closers[0])
	}
	second, ok := closers[1].(*testCloser)
	if !ok {
		t.Fatalf("expected *testCloser, got %T", closers[1])
	}
	if first.Name() != "a" || second.Name() != "b" {
		t.Fatalf("unexpected order: %v", closers)
	}
}

type testCloser struct {
	name string
}

func (c *testCloser) Close() error {
	return nil
}

func (c *testCloser) Name() string {
	return c.name
}
