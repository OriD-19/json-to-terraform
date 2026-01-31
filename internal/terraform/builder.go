package terraform

import (
	"bytes"
)

// TerraformBuilder collects resource blocks and template content for the final Terraform config.
type TerraformBuilder struct {
	resources   [][]byte
	variables   []byte
	outputs     []byte
	versions    []byte
	tfvars      []byte
	emitTfvars  bool
}

// NewBuilder returns a new TerraformBuilder.
func NewBuilder(emitTfvars bool) *TerraformBuilder {
	return &TerraformBuilder{
		resources:  nil,
		emitTfvars: emitTfvars,
	}
}

// AddResource appends a resource block (raw bytes from handler).
func (b *TerraformBuilder) AddResource(block []byte) {
	if len(block) == 0 {
		return
	}
	b.resources = append(b.resources, block)
}

// SetVariables sets the variables.tf content.
func (b *TerraformBuilder) SetVariables(content []byte) {
	b.variables = content
}

// SetOutputs sets the outputs.tf content.
func (b *TerraformBuilder) SetOutputs(content []byte) {
	b.outputs = content
}

// SetVersions sets the versions.tf content (terraform block + provider).
func (b *TerraformBuilder) SetVersions(content []byte) {
	b.versions = content
}

// SetTfvars sets the terraform.tfvars content (optional).
func (b *TerraformBuilder) SetTfvars(content []byte) {
	b.tfvars = content
}

// Build returns a map of filename -> content for all Terraform files.
func (b *TerraformBuilder) Build() map[string][]byte {
	out := make(map[string][]byte)
	if len(b.versions) > 0 {
		out["versions.tf"] = b.versions
	}
	if len(b.variables) > 0 {
		out["variables.tf"] = b.variables
	}
	var mainBuf bytes.Buffer
	for i, r := range b.resources {
		if i > 0 {
			mainBuf.WriteString("\n\n")
		}
		mainBuf.Write(r)
	}
	if mainBuf.Len() > 0 {
		out["main.tf"] = mainBuf.Bytes()
	}
	if len(b.outputs) > 0 {
		out["outputs.tf"] = b.outputs
	}
	if b.emitTfvars && len(b.tfvars) > 0 {
		out["terraform.tfvars"] = b.tfvars
	}
	return out
}
