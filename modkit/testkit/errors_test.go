package testkit_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-modkit/modkit/modkit/testkit"
)

func TestErrorStrings(t *testing.T) {
	if msg := (&testkit.ControllerNotFoundError{Module: "m", Name: "c"}).Error(); !strings.Contains(msg, "controller not found") {
		t.Fatalf("unexpected message: %q", msg)
	}

	if msg := (&testkit.TypeAssertionError{Target: "string", Actual: "int", Context: "token=\"x\""}).Error(); !strings.Contains(msg, "type assertion failed") {
		t.Fatalf("unexpected message: %q", msg)
	}

	if msg := (&testkit.NilOptionError{Index: 2}).Error(); !strings.Contains(msg, "nil testkit option") {
		t.Fatalf("unexpected message: %q", msg)
	}
}

func TestHarnessCloseErrorMessagesAndUnwrap(t *testing.T) {
	hookErr := errors.New("hook")
	closeErr := errors.New("close")

	both := &testkit.HarnessCloseError{HookErr: hookErr, CloseErr: closeErr}
	if msg := both.Error(); !strings.Contains(msg, "hooks=") || !strings.Contains(msg, "close=") {
		t.Fatalf("unexpected message: %q", msg)
	}
	if !errors.Is(both, hookErr) || !errors.Is(both, closeErr) {
		t.Fatalf("expected wrapped hook and close errors")
	}

	hooksOnly := &testkit.HarnessCloseError{HookErr: hookErr}
	if msg := hooksOnly.Error(); !strings.Contains(msg, "hooks=") {
		t.Fatalf("unexpected message: %q", msg)
	}

	closeOnly := &testkit.HarnessCloseError{CloseErr: closeErr}
	if msg := closeOnly.Error(); !strings.Contains(msg, "close=") {
		t.Fatalf("unexpected message: %q", msg)
	}

	none := &testkit.HarnessCloseError{}
	if got := none.Error(); got != "harness close failed" {
		t.Fatalf("unexpected message: %q", got)
	}
}
