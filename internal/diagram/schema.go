package diagram

// Diagram is the root structure of the infrastructure diagram JSON.
type Diagram struct {
	Metadata Metadata  `json:"metadata"`
	Nodes    []Node    `json:"nodes"`
	Edges    []Edge    `json:"edges"`
}

// Metadata holds diagram-level information.
type Metadata struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Environment string `json:"environment"`
}

// Node represents a single resource in the diagram.
type Node struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Label      string            `json:"label"`
	Position   Position         `json:"position"`
	Properties map[string]any    `json:"properties"`
}

// Position holds x,y coordinates (used by the diagram UI).
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Edge represents a relationship between two nodes.
type Edge struct {
	ID         string         `json:"id"`
	Source     string         `json:"source"`
	Target     string         `json:"target"`
	Type       string         `json:"type"` // contains, connects_to, depends_on
	Properties map[string]any `json:"properties"`
}
