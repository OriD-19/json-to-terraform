package registry

import (
	"sync"

	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/result"
)

// RefMap maps node IDs to Terraform resource addresses (e.g. "node-3" -> "aws_vpc.node_3").
type RefMap map[string]string

// ResourceHandler is the interface each resource type handler must implement.
type ResourceHandler interface {
	ResourceType() string
	Validate(node *diagram.Node) ([]result.Error, []result.Warning)
	GenerateHCL(node *diagram.Node, d *diagram.Diagram, refs RefMap) ([]byte, error)
}

// Default is the global handler registry.
var Default = New()

// Registry holds resource type handlers.
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]ResourceHandler
}

// New returns a new empty registry.
func New() *Registry {
	return &Registry{handlers: make(map[string]ResourceHandler)}
}

// Register adds a handler for the given resource type.
func (r *Registry) Register(resourceType string, h ResourceHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[resourceType] = h
}

// Get returns the handler for the resource type, or nil and false.
func (r *Registry) Get(resourceType string) (ResourceHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[resourceType]
	return h, ok
}

// ListSupportedTypes returns all registered resource types.
func (r *Registry) ListSupportedTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}
