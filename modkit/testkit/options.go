package testkit

import (
	"context"

	"github.com/go-modkit/modkit/modkit/module"
)

type config struct {
	overrides []Override
	autoClose bool
}

func defaultConfig() config {
	return config{
		overrides: make([]Override, 0),
		autoClose: true,
	}
}

// Option configures testkit harness construction.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (f optionFunc) apply(cfg *config) {
	f(cfg)
}

// Override describes a token-level provider override for tests.
type Override struct {
	Token   module.Token
	Build   func(module.Resolver) (any, error)
	Cleanup func(context.Context) error
}

// WithOverrides applies provider overrides for harness bootstrap.
func WithOverrides(overrides ...Override) Option {
	cloned := make([]Override, len(overrides))
	copy(cloned, overrides)

	return optionFunc(func(cfg *config) {
		cfg.overrides = append(cfg.overrides, cloned...)
	})
}

// OverrideValue returns a static value override.
func OverrideValue(token module.Token, value any) Override {
	return Override{
		Token: token,
		Build: func(module.Resolver) (any, error) {
			return value, nil
		},
	}
}

// OverrideBuild returns a dynamic build override.
func OverrideBuild(token module.Token, build func(module.Resolver) (any, error)) Override {
	return Override{Token: token, Build: build}
}

// WithoutAutoClose disables automatic tb.Cleanup registration.
func WithoutAutoClose() Option {
	return optionFunc(func(cfg *config) {
		cfg.autoClose = false
	})
}
