package module

import "fmt"

// Get resolves a provider of type T from the resolver.
// It returns an error if the resolution fails or if the resolved value is not of type T.
func Get[T any](r Resolver, token Token) (T, error) {
	var zero T
	val, err := r.Get(token)
	if err != nil {
		return zero, err
	}

	typed, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("provider %q resolved to %T, expected %T", token, val, zero)
	}

	return typed, nil
}
