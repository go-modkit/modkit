package kernel

import (
	"context"

	"github.com/go-modkit/modkit/modkit/module"
)

// ProviderOverride replaces provider build/cleanup behavior for a token at bootstrap time.
type ProviderOverride struct {
	Token   module.Token
	Build   func(module.Resolver) (any, error)
	Cleanup func(context.Context) error
}

// BootstrapOption configures advanced bootstrap behavior.
type BootstrapOption interface {
	apply(*bootstrapConfig)
}

type providerOverridesOption struct {
	overrides []ProviderOverride
}

func (o providerOverridesOption) apply(cfg *bootstrapConfig) {
	cfg.addProviderOverrides(o.overrides)
}

// WithProviderOverrides applies token-level provider overrides for bootstrap.
func WithProviderOverrides(overrides ...ProviderOverride) BootstrapOption {
	cloned := make([]ProviderOverride, len(overrides))
	copy(cloned, overrides)
	return providerOverridesOption{overrides: cloned}
}
