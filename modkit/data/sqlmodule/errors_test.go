package sqlmodule

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildErrorIncludesProvider(t *testing.T) {
	inner := errors.New("boom")
	be := &BuildError{Provider: "postgres", Token: "db", Stage: StageOpen, Err: inner}
	msg := be.Error()

	if !strings.Contains(msg, "postgres provider build failed") {
		t.Fatalf("expected provider in message, got %q", msg)
	}
	if !strings.Contains(msg, "token=\"db\"") {
		t.Fatalf("expected token in message, got %q", msg)
	}
	if !strings.Contains(msg, "stage=open") {
		t.Fatalf("expected stage in message, got %q", msg)
	}
	if !errors.Is(be, inner) {
		t.Fatalf("expected error to unwrap")
	}
}

func TestBuildErrorWithEmptyProviderUsesGenericPrefix(t *testing.T) {
	inner := errors.New("boom")
	be := &BuildError{Token: "db", Stage: StageOpen, Err: inner}
	msg := be.Error()

	if !strings.Contains(msg, "sql provider build failed") {
		t.Fatalf("expected generic prefix, got %q", msg)
	}
}
