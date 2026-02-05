package kernel_test

import (
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
