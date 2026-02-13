package sqlite

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCleanupDBWrapsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := CleanupDB(ctx, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected wrapped context error")
	}
	if !strings.Contains(err.Error(), "cleanup") {
		t.Fatalf("expected cleanup context in error, got %q", err.Error())
	}
}
