package schema

import (
	"fmt"
	"sync"
)

// Registry manages schema instances
type Registry struct {
	mu      sync.RWMutex
	schemas map[string]Schema
}

// Global registry instance
var globalRegistry = NewRegistry()

// NewRegistry creates a new schema registry
func NewRegistry() *Registry {
	r := &Registry{
		schemas: make(map[string]Schema),
	}
	// Register built-in schemas
	r.registerBuiltins()
	return r
}

// registerBuiltins registers all built-in schemas
func (r *Registry) registerBuiltins() {
	builtins := []Schema{
		NewInstructionSchema(),
		NewChatSchema(),
		NewPreferenceSchema(),
		NewClassificationSchema(),
	}
	for _, s := range builtins {
		if err := r.Register(s); err != nil {
			panic(fmt.Sprintf("failed to register built-in schema %q: %v", s.Name(), err))
		}
	}
}

// Register adds a schema to the registry
func (r *Registry) Register(schema Schema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := schema.Name()
	if _, exists := r.schemas[name]; exists {
		return fmt.Errorf("schema already registered: %s", name)
	}

	r.schemas[name] = schema
	return nil
}

// Get retrieves a schema by name
func (r *Registry) Get(name string) (Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if schema, ok := r.schemas[name]; ok {
		return schema, nil
	}
	return nil, fmt.Errorf("schema not found: %s", name)
}

// List returns all registered schema names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}
	return names
}

// ListByStyle returns schemas matching a specific style
func (r *Registry) ListByStyle(style DatasetStyle) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0)
	for name, schema := range r.schemas {
		if schema.Style() == style {
			names = append(names, name)
		}
	}
	return names
}

// Global registry functions

// Register adds a schema to the global registry
func Register(schema Schema) error {
	return globalRegistry.Register(schema)
}

// Get retrieves a schema from the global registry
func Get(name string) (Schema, error) {
	return globalRegistry.Get(name)
}

// List returns all schemas in the global registry
func List() []string {
	return globalRegistry.List()
}

// ListByStyle returns schemas matching a style
func ListByStyle(style DatasetStyle) []string {
	return globalRegistry.ListByStyle(style)
}
