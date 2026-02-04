package module

// ProviderDef describes how to build a provider for a token.
type ProviderDef struct {
	Token Token
	Build func(r Resolver) (any, error)
}
