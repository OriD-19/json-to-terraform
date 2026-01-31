package terraform

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// SanitizeName converts a node id to a Terraform-safe resource name (e.g. node-1 -> node_1).
func SanitizeName(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}

// ResourceBlock creates a resource "type" "name" { } block; body can be filled by the caller.
func ResourceBlock(resourceType, name string) *hclwrite.Block {
	return hclwrite.NewBlock("resource", []string{resourceType, name})
}

// SetAttributeStr sets a string attribute on a block body.
func SetAttributeStr(body *hclwrite.Body, name, value string) {
	if value != "" {
		body.SetAttributeValue(name, cty.StringVal(value))
	}
}

// SetAttributeBool sets a bool attribute.
func SetAttributeBool(body *hclwrite.Body, name string, value bool) {
	body.SetAttributeValue(name, cty.BoolVal(value))
}

// SetAttributeInt sets an int attribute (Terraform numbers are arbitrary precision; we use int).
func SetAttributeInt(body *hclwrite.Body, name string, value int) {
	body.SetAttributeValue(name, cty.NumberIntVal(int64(value)))
}

// SetAttributeMap sets a map(string) attribute (e.g. tags).
func SetAttributeMap(body *hclwrite.Body, name string, m map[string]string) {
	if len(m) == 0 {
		return
	}
	ctyMap := make(map[string]cty.Value)
	for k, v := range m {
		ctyMap[k] = cty.StringVal(v)
	}
	body.SetAttributeValue(name, cty.MapVal(ctyMap))
}

// BlockToBytes formats a block and returns its bytes (with newline).
func BlockToBytes(block *hclwrite.Block) []byte {
	f := hclwrite.NewEmptyFile()
	f.Body().AppendBlock(block)
	return f.Bytes()
}
