package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestCleanupDB_ReturnsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := CleanupDB(ctx, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestCleanupDB_AllowsNilDB(t *testing.T) {
	err := CleanupDB(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCleanupDB_ReturnsContextErrorBeforeClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := CleanupDB(ctx, &sql.DB{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
