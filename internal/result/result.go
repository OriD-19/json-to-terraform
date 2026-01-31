package result

// Error represents a validation or generation error (AGENTS.md format).
type Error struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	NodeID     string `json:"node_id,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Warning represents a best-practice or non-fatal warning.
type Warning struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	NodeID     string `json:"node_id,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ParseResult is the result of parsing a diagram.
type ParseResult struct {
	Success       bool              `json:"success"`
	TerraformFiles map[string][]byte `json:"-"` // filename -> content
	Errors        []Error           `json:"errors,omitempty"`
	Warnings      []Warning         `json:"warnings,omitempty"`
}
