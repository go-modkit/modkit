package testkit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-modkit/modkit/modkit/module"
	"github.com/go-modkit/modkit/modkit/testkit"
)

type testModule struct {
	def module.ModuleDef
}

func (m *testModule) Definition() module.ModuleDef {
	return m.def
}

func mod(
	providers []module.ProviderDef,
	controllers []module.ControllerDef,
	exports []module.Token,
) module.Module {
	return &testModule{
		def: module.ModuleDef{
			Name:        "root",
			Providers:   providers,
			Controllers: controllers,
			Exports:     exports,
		},
	}
}

type tbStub struct {
	cleanup []func()
	failed  bool
	msg     string
}

func (t *tbStub) Helper() {}

func (t *tbStub) Fatalf(format string, _ ...any) {
	t.failed = true
	t.msg = format
}

func (t *tbStub) Cleanup(fn func()) {
	t.cleanup = append(t.cleanup, fn)
}

type closerErr struct {
	err error
}

func (c *closerErr) Close() error {
	return c.err
}

type closerCount struct {
	count int
}

func (c *closerCount) Close() error {
	c.count++
	return nil
}

func TestNewE_BootstrapsHarness(t *testing.T) {
	token := module.Token("svc.token")
	root := mod(
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) { return "ok", nil },
		}},
		[]module.ControllerDef{{
			Name:  "Controller",
			Build: func(r module.Resolver) (any, error) { return r.Get(token) },
		}},
		[]module.Token{token},
	)

	h, err := testkit.NewE(t, root)
	if err != nil {
		t.Fatalf("NewE failed: %v", err)
	}

	got, err := testkit.GetE[string](h, token)
	if err != nil {
		t.Fatalf("GetE failed: %v", err)
	}
	if got != "ok" {
		t.Fatalf("unexpected value: %v", got)
	}

	controller, err := testkit.ControllerE[string](h, "root", "Controller")
	if err != nil {
		t.Fatalf("ControllerE failed: %v", err)
	}
	if controller != "ok" {
		t.Fatalf("unexpected controller value: %v", controller)
	}
}

func TestNew_UsesFatalfOnBootstrapError(t *testing.T) {
	tb := &tbStub{}

	h := testkit.New(tb, nil)
	if h != nil {
		t.Fatal("expected nil harness")
	}
	if !tb.failed {
		t.Fatal("expected Fatalf to be called")
	}
}

func TestWithOverrides_ReplacesProvider(t *testing.T) {
	token := module.Token("svc.token")
	root := mod(
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) { return "real", nil },
		}},
		nil,
		[]module.Token{token},
	)

	h, err := testkit.NewE(t, root, testkit.WithOverrides(testkit.OverrideValue(token, "fake")))
	if err != nil {
		t.Fatalf("NewE failed: %v", err)
	}

	got, err := testkit.GetE[string](h, token)
	if err != nil {
		t.Fatalf("GetE failed: %v", err)
	}
	if got != "fake" {
		t.Fatalf("unexpected value: %v", got)
	}
}

func TestGetE_ReturnsTypeAssertionError(t *testing.T) {
	token := module.Token("svc.token")
	root := mod(
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) { return 1, nil },
		}},
		nil,
		[]module.Token{token},
	)

	h, err := testkit.NewE(t, root)
	if err != nil {
		t.Fatalf("NewE failed: %v", err)
	}

	_, err = testkit.GetE[string](h, token)
	if err == nil {
		t.Fatal("expected type assertion error")
	}

	var typeErr *testkit.TypeAssertionError
	if !errors.As(err, &typeErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestControllerE_ReturnsControllerNotFoundError(t *testing.T) {
	root := mod(nil, nil, nil)
	h, err := testkit.NewE(t, root)
	if err != nil {
		t.Fatalf("NewE failed: %v", err)
	}

	_, err = testkit.ControllerE[string](h, "root", "Missing")
	if err == nil {
		t.Fatal("expected not found error")
	}

	var notFoundErr *testkit.ControllerNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestCloseContext_IdempotentAndAggregatesErrors(t *testing.T) {
	token := module.Token("svc.token")
	hookErr := errors.New("hook boom")
	closeErr := errors.New("close boom")

	root := mod(
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) {
				return &closerErr{err: closeErr}, nil
			},
			Cleanup: func(context.Context) error { return nil },
		}},
		nil,
		[]module.Token{token},
	)

	h, err := testkit.NewE(t, root, testkit.WithOverrides(testkit.Override{
		Token: token,
		Build: func(module.Resolver) (any, error) {
			return &closerErr{err: closeErr}, nil
		},
		Cleanup: func(context.Context) error { return hookErr },
	}))
	if err != nil {
		t.Fatalf("NewE failed: %v", err)
	}

	_, err = testkit.GetE[*closerErr](h, token)
	if err != nil {
		t.Fatalf("GetE failed: %v", err)
	}

	err = h.Close()
	if err == nil {
		t.Fatal("expected close error")
	}

	var closeAggregate *testkit.HarnessCloseError
	if !errors.As(err, &closeAggregate) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if !errors.Is(err, hookErr) {
		t.Fatalf("expected hook error to be wrapped")
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("expected close error to be wrapped")
	}

	err2 := h.Close()
	if !errors.Is(err2, hookErr) || !errors.Is(err2, closeErr) {
		t.Fatalf("expected idempotent close to return same wrapped errors")
	}
}

func TestWithoutAutoClose_DisablesCleanupRegistration(t *testing.T) {
	tb := &tbStub{}
	root := mod(nil, nil, nil)

	_, err := testkit.NewE(tb, root, testkit.WithoutAutoClose())
	if err != nil {
		t.Fatalf("NewE failed: %v", err)
	}

	if len(tb.cleanup) != 0 {
		t.Fatalf("expected no cleanup registration, got %d", len(tb.cleanup))
	}
}

func TestNewE_DefaultAutoCloseRegistersCleanup(t *testing.T) {
	tb := &tbStub{}
	root := mod(nil, nil, nil)

	_, err := testkit.NewE(tb, root)
	if err != nil {
		t.Fatalf("NewE failed: %v", err)
	}

	if len(tb.cleanup) != 1 {
		t.Fatalf("expected one cleanup registration, got %d", len(tb.cleanup))
	}
}

func TestNewE_RejectsNilOption(t *testing.T) {
	root := mod(nil, nil, nil)

	_, err := testkit.NewE(t, root, nil)
	if err == nil {
		t.Fatal("expected nil option error")
	}

	var nilOptErr *testkit.NilOptionError
	if !errors.As(err, &nilOptErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestCloseContext_CancelledContextCanRetryAndClosesOnce(t *testing.T) {
	token := module.Token("svc.token")
	tracker := &closerCount{}

	root := mod(
		[]module.ProviderDef{{
			Token: token,
			Build: func(module.Resolver) (any, error) {
				return tracker, nil
			},
			Cleanup: func(context.Context) error { return nil },
		}},
		nil,
		[]module.Token{token},
	)

	h, err := testkit.NewE(t, root)
	if err != nil {
		t.Fatalf("NewE failed: %v", err)
	}

	_, err = testkit.GetE[*closerCount](h, token)
	if err != nil {
		t.Fatalf("GetE failed: %v", err)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	err = h.CloseContext(canceled)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}

	if err := h.Close(); err != nil {
		t.Fatalf("expected retry close to succeed, got %v", err)
	}
	if tracker.count != 1 {
		t.Fatalf("expected closer called once, got %d", tracker.count)
	}

	if err := h.Close(); err != nil {
		t.Fatalf("expected idempotent close, got %v", err)
	}
	if tracker.count != 1 {
		t.Fatalf("expected closer still called once after idempotent close, got %d", tracker.count)
	}
}
