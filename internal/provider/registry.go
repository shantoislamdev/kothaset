package provider

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Config contains provider configuration for creating a provider instance
type Config struct {
	Name       string
	Type       string
	BaseURL    string
	APIKey     string
	Model      string
	MaxRetries int
	Timeout    time.Duration
	RateLimit  int // requests per minute
	Headers    map[string]string
}

// Registry manages provider instances
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	factories map[string]Factory
}

// Factory creates a provider from configuration
type Factory func(cfg *Config) (Provider, error)

// Global registry instance
var globalRegistry = NewRegistry()

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	r := &Registry{
		providers: make(map[string]Provider),
		factories: make(map[string]Factory),
	}
	// Register built-in factories
	r.RegisterFactory("openai", NewOpenAIProvider)
	return r
}

// RegisterFactory registers a provider factory by type
func (r *Registry) RegisterFactory(providerType string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[providerType] = factory
}

// Register adds a provider instance to the registry
func (r *Registry) Register(name string, provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider already registered: %s", name)
	}

	r.providers[name] = provider
	return nil
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if provider, ok := r.providers[name]; ok {
		return provider, nil
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

// GetOrCreate retrieves a provider or creates it from config
func (r *Registry) GetOrCreate(cfg *Config) (Provider, error) {
	// Check if already created
	r.mu.RLock()
	if provider, ok := r.providers[cfg.Name]; ok {
		r.mu.RUnlock()
		return provider, nil
	}
	r.mu.RUnlock()

	// Create new provider
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if provider, ok := r.providers[cfg.Name]; ok {
		return provider, nil
	}

	// Find factory for provider type
	factory, ok := r.factories[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unknown provider type: %s", cfg.Type)
	}

	// Create provider
	provider, err := factory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Register and return
	r.providers[cfg.Name] = provider
	return provider, nil
}

// List returns all registered provider names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Close closes all providers
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for name, provider := range r.providers {
		if err := provider.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close provider %s: %w", name, err))
		}
	}
	r.providers = make(map[string]Provider)

	return errors.Join(errs...)
}

// Global registry functions

// RegisterFactory registers a factory in the global registry
func RegisterFactory(providerType string, factory Factory) {
	globalRegistry.RegisterFactory(providerType, factory)
}

// Register adds a provider to the global registry
func Register(name string, provider Provider) error {
	return globalRegistry.Register(name, provider)
}

// Get retrieves a provider from the global registry
func Get(name string) (Provider, error) {
	return globalRegistry.Get(name)
}

// GetOrCreate retrieves or creates a provider in the global registry
func GetOrCreate(cfg *Config) (Provider, error) {
	return globalRegistry.GetOrCreate(cfg)
}

// List returns all providers in the global registry
func List() []string {
	return globalRegistry.List()
}

// CloseAll closes all providers in the global registry
func CloseAll() error {
	return globalRegistry.Close()
}
