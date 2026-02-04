package kernel

import (
	"sync"

	"github.com/aryeko/modkit/modkit/module"
)

type providerEntry struct {
	moduleName string
	build      func(r module.Resolver) (any, error)
}

type Container struct {
	providers  map[module.Token]providerEntry
	instances  map[module.Token]any
	visibility Visibility
	locks      map[module.Token]*sync.Mutex
	mu         sync.Mutex
}

func newContainer(graph *Graph, visibility Visibility) (*Container, error) {
	providers := make(map[module.Token]providerEntry)
	for _, node := range graph.Modules {
		for _, provider := range node.Def.Providers {
			if existing, exists := providers[provider.Token]; exists {
				return nil, &DuplicateProviderTokenError{
					Token:   provider.Token,
					Modules: []string{existing.moduleName, node.Name},
				}
			}
			providers[provider.Token] = providerEntry{
				moduleName: node.Name,
				build:      provider.Build,
			}
		}
	}

	return &Container{
		providers:  providers,
		instances:  make(map[module.Token]any),
		visibility: visibility,
		locks:      make(map[module.Token]*sync.Mutex),
	}, nil
}

// Get resolves a provider without module visibility checks.
// Visibility enforcement is applied via module-scoped resolvers.
func (c *Container) Get(token module.Token) (any, error) {
	return c.getWithStack(token, "", nil)
}

func (c *Container) getWithStack(token module.Token, requester string, stack []module.Token) (any, error) {
	for _, item := range stack {
		if item == token {
			return nil, &ProviderCycleError{Token: token}
		}
	}

	entry, ok := c.providers[token]
	if !ok {
		return nil, &ProviderNotFoundError{Module: requester, Token: token}
	}

	c.mu.Lock()
	instance, ok := c.instances[token]
	lock, lockExists := c.locks[token]
	if ok {
		c.mu.Unlock()
		return instance, nil
	}
	if !lockExists {
		lock = &sync.Mutex{}
		c.locks[token] = lock
	}
	c.mu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	c.mu.Lock()
	instance, ok = c.instances[token]
	c.mu.Unlock()
	if ok {
		return instance, nil
	}

	nextStack := append(append([]module.Token{}, stack...), token)
	resolver := moduleResolver{
		container:  c,
		moduleName: entry.moduleName,
		stack:      nextStack,
	}
	instance, err := entry.build(resolver)
	if err != nil {
		return nil, &ProviderBuildError{Module: entry.moduleName, Token: token, Err: err}
	}

	c.mu.Lock()
	c.instances[token] = instance
	c.mu.Unlock()
	return instance, nil
}

type moduleResolver struct {
	container  *Container
	moduleName string
	stack      []module.Token
}

func (r moduleResolver) Get(token module.Token) (any, error) {
	visibility := r.container.visibility[r.moduleName]
	if !visibility[token] {
		return nil, &TokenNotVisibleError{Module: r.moduleName, Token: token}
	}
	return r.container.getWithStack(token, r.moduleName, r.stack)
}

func (c *Container) resolverFor(moduleName string) module.Resolver {
	return moduleResolver{
		container:  c,
		moduleName: moduleName,
	}
}
