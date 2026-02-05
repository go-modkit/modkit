package kernel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/kernel"
	"github.com/go-modkit/modkit/modkit/module"
)

type recorder struct {
	id    string
	calls *[]string
}

func (r *recorder) Close() error {
	*r.calls = append(*r.calls, r.id)
	return nil
}

type closerErr struct{ err error }

func (c *closerErr) Close() error { return c.err }

func TestAppCloseClosesProvidersInReverseBuildOrder(t *testing.T) {
	tokenB := module.Token("close.b")
	tokenA := module.Token("close.a")
	calls := make([]string, 0, 2)

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: tokenB,
			Build: func(r module.Resolver) (any, error) {
				return &recorder{id: "B", calls: &calls}, nil
			},
		}, {
			Token: tokenA,
			Build: func(r module.Resolver) (any, error) {
				_, err := r.Get(tokenB)
				if err != nil {
					return nil, err
				}
				return &recorder{id: "A", calls: &calls}, nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if _, err := app.Get(tokenA); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if err := app.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 close calls, got %d", len(calls))
	}
	if calls[0] != "A" || calls[1] != "B" {
		t.Fatalf("unexpected close order: %v", calls)
	}
}

func TestAppCloseAggregatesErrors(t *testing.T) {
	tokenA := module.Token("close.err.a")
	tokenB := module.Token("close.err.b")
	errA := errors.New("close a")
	errB := errors.New("close b")

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: tokenA,
			Build: func(r module.Resolver) (any, error) { return &closerErr{err: errA}, nil },
		}, {
			Token: tokenB,
			Build: func(r module.Resolver) (any, error) { return &closerErr{err: errB}, nil },
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if _, err := app.Get(tokenA); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if _, err := app.Get(tokenB); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	err = app.Close()
	if err == nil {
		t.Fatalf("expected close error")
	}

	if !errors.Is(err, errA) || !errors.Is(err, errB) {
		t.Fatalf("expected aggregated errors, got %v", err)
	}
}

func TestAppCloseContextCanceledWithNoClosers(t *testing.T) {
	modA := mod("A", nil, nil, nil, nil)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = app.CloseContext(ctx)
	if err == nil {
		t.Fatalf("expected close error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled error, got %v", err)
	}
}

func TestAppCloseContextCanceledWithClosers(t *testing.T) {
	tokenB := module.Token("close.ctx.b")
	tokenA := module.Token("close.ctx.a")
	calls := make([]string, 0, 2)

	modA := mod("A", nil,
		[]module.ProviderDef{{
			Token: tokenB,
			Build: func(r module.Resolver) (any, error) {
				return &recorder{id: "B", calls: &calls}, nil
			},
		}, {
			Token: tokenA,
			Build: func(r module.Resolver) (any, error) {
				_, err := r.Get(tokenB)
				if err != nil {
					return nil, err
				}
				return &recorder{id: "A", calls: &calls}, nil
			},
		}},
		nil,
		nil,
	)

	app, err := kernel.Bootstrap(modA)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	if _, err := app.Get(tokenA); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = app.CloseContext(ctx)
	if err == nil {
		t.Fatalf("expected close error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled error, got %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 close call, got %d", len(calls))
	}
	if calls[0] != "A" {
		t.Fatalf("unexpected close order: %v", calls)
	}
}
