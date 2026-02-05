package auth

import "context"

type User struct {
	ID    string
	Email string
}

type userKey struct{}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	val := ctx.Value(userKey{})
	user, ok := val.(User)
	return user, ok
}
