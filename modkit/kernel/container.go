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
	}, nil
}

func (c *Container) Get(token module.Token) (any, error) {
	return c.get(token, "")
}

func (c *Container) get(token module.Token, requester string) (any, error) {
	c.mu.Lock()
	instance, ok := c.instances[token]
	c.mu.Unlock()
	if ok {
		return instance, nil
	}

	entry, ok := c.providers[token]
	if !ok {
		return nil, &ProviderNotFoundError{Module: requester, Token: token}
	}

	resolver := moduleResolver{
		container:  c,
		moduleName: entry.moduleName,
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
}

func (r moduleResolver) Get(token module.Token) (any, error) {
	visibility := r.container.visibility[r.moduleName]
	if !visibility[token] {
		return nil, &TokenNotVisibleError{Module: r.moduleName, Token: token}
	}
	return r.container.get(token, r.moduleName)
}

func (c *Container) resolverFor(moduleName string) module.Resolver {
	return moduleResolver{
		container:  c,
		moduleName: moduleName,
	}
}
