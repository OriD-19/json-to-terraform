package diagram

import (
	"fmt"
)

// ValidationError represents a single validation failure (schema/structure level).
type ValidationError struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"` // error
	NodeID     string `json:"node_id,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Validate checks required fields and structure of the diagram.
// Resource-specific validation is done by handlers.
func Validate(d *Diagram) []ValidationError {
	var errs []ValidationError

	if d == nil {
		return []ValidationError{{Type: "schema_error", Severity: "error", Message: "diagram is nil"}}
	}

	if d.Metadata.Version == "" {
		errs = append(errs, ValidationError{
			Type: "schema_error", Severity: "error",
			Message: "metadata.version is required", Suggestion: "Set metadata.version (e.g. \"1.0\")",
		})
	}

	seenNodeIDs := make(map[string]bool)
	for i := range d.Nodes {
		n := &d.Nodes[i]
		if n.ID == "" {
			errs = append(errs, ValidationError{
				Type: "schema_error", Severity: "error", NodeID: n.ID,
				Message: fmt.Sprintf("node at index %d has empty id", i), Suggestion: "Set node.id",
			})
		} else if seenNodeIDs[n.ID] {
			errs = append(errs, ValidationError{
				Type: "schema_error", Severity: "error", NodeID: n.ID,
				Message: "duplicate node id: " + n.ID, Suggestion: "Use unique ids for each node",
			})
		} else {
			seenNodeIDs[n.ID] = true
		}
		if n.Type == "" {
			errs = append(errs, ValidationError{
				Type: "schema_error", Severity: "error", NodeID: n.ID,
				Message: "node.type is required", Suggestion: "Set node.type (e.g. ec2_instance, vpc)",
			})
		}
		if n.Properties == nil {
			n.Properties = make(map[string]any)
		}
	}

	for i := range d.Edges {
		e := &d.Edges[i]
		if e.Source == "" || e.Target == "" {
			errs = append(errs, ValidationError{
				Type: "schema_error", Severity: "error",
				Message: fmt.Sprintf("edge at index %d must have source and target", i),
				Suggestion: "Set edge.source and edge.target to node ids",
			})
		} else if !seenNodeIDs[e.Source] {
			errs = append(errs, ValidationError{
				Type: "schema_error", Severity: "error",
				Message: "edge source node not found: " + e.Source,
				Suggestion: "Reference an existing node id",
			})
		} else if !seenNodeIDs[e.Target] {
			errs = append(errs, ValidationError{
				Type: "schema_error", Severity: "error",
				Message: "edge target node not found: " + e.Target,
				Suggestion: "Reference an existing node id",
			})
		}
		if e.Type == "" {
			e.Type = "depends_on"
		}
	}

	return errs
}

// NodeByID returns the node with the given id, or nil.
func (d *Diagram) NodeByID(id string) *Node {
	for i := range d.Nodes {
		if d.Nodes[i].ID == id {
			return &d.Nodes[i]
		}
	}
	return nil
}

// EdgesWithTarget returns edges whose target is the given node id.
func (d *Diagram) EdgesWithTarget(targetID string) []Edge {
	var out []Edge
	for _, e := range d.Edges {
		if e.Target == targetID {
			out = append(out, e)
		}
	}
	return out
}

// EdgesWithSource returns edges whose source is the given node id.
func (d *Diagram) EdgesWithSource(sourceID string) []Edge {
	var out []Edge
	for _, e := range d.Edges {
		if e.Source == sourceID {
			out = append(out, e)
		}
	}
	return out
}

// GetStr gets a string property; empty if missing or not a string.
func GetStr(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetBool gets a bool property.
func GetBool(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	v, ok := m[key]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// GetInt gets an int property (from float64 JSON number).
func GetInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

// GetMap gets a nested map (e.g. tags, environment_variables).
func GetMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	mm, _ := v.(map[string]any)
	return mm
}

// GetStrMap returns a map of string -> string (e.g. tags).
func GetStrMap(m map[string]any, key string) map[string]string {
	raw := GetMap(m, key)
	if raw == nil {
		return nil
	}
	out := make(map[string]string)
	for k, v := range raw {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

