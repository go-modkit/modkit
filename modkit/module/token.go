package module

// Token identifies a provider for resolution.
type Token string

// Resolver provides access to resolved provider instances.
type Resolver interface {
	Get(Token) (any, error)
}
