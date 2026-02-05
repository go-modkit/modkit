package auth

import (
	"context"
	"testing"
)

func TestUserContextHelpers(t *testing.T) {
	ctx := context.Background()
	user := User{ID: "demo", Email: "demo@example.com"}

	ctx = WithUser(ctx, user)
	got, ok := UserFromContext(ctx)
	if !ok {
		t.Fatal("expected user in context")
	}
	if got.Email != user.Email {
		t.Fatalf("expected %s, got %s", user.Email, got.Email)
	}
}
