package provider

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/x2d7/interlude/chat"
)

// ProviderEnvelope wraps provider-specific config JSON with a type discriminator.
type ProviderEnvelope struct {
	Provider string          `json:"provider"`
	Config   json.RawMessage `json:"config"`
}

// Handler deserializes provider-specific config bytes into a chat.Client.
type Handler func(configBytes []byte) (chat.Client, error)

// Provider describes a registered, creatable provider and the JSON Schema for its config.
type Provider struct {
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`
}

// Registry maps provider names to their deserialization handlers.
type Registry struct {
	mu        sync.RWMutex
	handlers  map[string]Handler
	schemaMap map[string]SchemaHandler
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		handlers:  make(map[string]Handler),
		schemaMap: make(map[string]SchemaHandler),
	}
}

// Register adds a deserialization handler for a provider.
func (r *Registry) Register(name string, handler Handler) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[name]; exists {
		return fmt.Errorf("provider: %q already registered", name)
	}
	r.handlers[name] = handler
	return nil
}

// Get returns the handler for a provider name, or nil if not registered.
func (r *Registry) Get(name string) Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.handlers[name]
}

// RegisterSchema adds a schema handler for a provider.
func (r *Registry) RegisterSchema(name string, handler SchemaHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.schemaMap[name]; exists {
		return fmt.Errorf("provider: schema for %q already registered", name)
	}
	r.schemaMap[name] = handler
	return nil
}

// GetSchema returns the schema handler for a provider name, or nil if not registered.
func (r *Registry) GetSchema(name string) SchemaHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.schemaMap[name]
}

// Schema returns the JSON Schema for a provider's config by name.
func (r *Registry) Schema(name string) (map[string]any, error) {
	handler := r.GetSchema(name)
	if handler == nil {
		return nil, fmt.Errorf("provider: schema for %q not registered", name)
	}
	return handler()
}

// Providers returns all registered providers that have a schema handler, each with
// its config schema, sorted by name. A provider whose schema handler fails is skipped.
func (r *Registry) Providers() []Provider {
	r.mu.RLock()
	schemas := make(map[string]SchemaHandler, len(r.schemaMap))
	for name, handler := range r.schemaMap {
		schemas[name] = handler
	}
	r.mu.RUnlock()

	out := make([]Provider, 0, len(schemas))
	for name, handler := range schemas {
		schema, err := handler()
		if err != nil {
			continue
		}
		out = append(out, Provider{Name: name, Schema: schema})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// DefaultRegistry is the global registry. Provider packages register their handlers in init().
var DefaultRegistry = NewRegistry()
