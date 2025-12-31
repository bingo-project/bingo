// ABOUTME: Provider registry for managing AI providers.
// ABOUTME: Supports registration, lookup, and model discovery.

package ai

import (
	"sync"
)

// Registry manages registered providers
type Registry struct {
	providers map[string]Provider
	models    map[string]Provider // model -> provider mapping
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		models:    make(map[string]Provider),
	}
}

// Register registers a provider
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[p.Name()] = p
	for _, m := range p.Models() {
		r.models[m.ID] = p
	}
}

// Get returns a provider by name
func (r *Registry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]

	return p, ok
}

// GetByModel returns a provider by model ID
func (r *Registry) GetByModel(model string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.models[model]

	return p, ok
}

// ListProviders returns all registered provider names
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}

	return names
}

// ListModels returns all registered models
func (r *Registry) ListModels() []ModelInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []ModelInfo
	for _, p := range r.providers {
		models = append(models, p.Models()...)
	}

	return models
}

// Clear removes all registered providers and models.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = make(map[string]Provider)
	r.models = make(map[string]Provider)
}
