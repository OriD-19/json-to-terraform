package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/zclconf/go-cty/cty"
)

// VersionsTF returns content for versions.tf (terraform block + aws provider).
func VersionsTF() []byte {
	f := hclwrite.NewEmptyFile()
	body := f.Body()

	tfBlock := body.AppendNewBlock("terraform", nil)
	tfBody := tfBlock.Body()
	tfBody.SetAttributeValue("required_version", cty.StringVal(">= 1.0"))
	reqProv := tfBody.AppendNewBlock("required_providers", nil)
	reqProv.Body().SetAttributeValue("aws", cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal("hashicorp/aws"),
		"version": cty.StringVal("~> 5.0"),
	}))

	body.AppendNewline()
	provBlock := body.AppendNewBlock("provider", []string{"aws"})
	provBlock.Body().SetAttributeTraversal("region", varTraversal("aws_region"))

	return f.Bytes()
}

// VariablesTF returns content for variables.tf (aws_region and optional vars).
func VariablesTF() []byte {
	f := hclwrite.NewEmptyFile()
	body := f.Body()

	// aws_region
	regionBlock := body.AppendNewBlock("variable", []string{"aws_region"})
	regionBlock.Body().SetAttributeValue("description", cty.StringVal("AWS region"))
	regionBlock.Body().SetAttributeValue("type", cty.StringVal("string"))
	regionBlock.Body().SetAttributeValue("default", cty.StringVal("us-east-1"))

	return f.Bytes()
}

// OutputsTF returns minimal outputs.tf (optional; can be extended from metadata).
func OutputsTF() []byte {
	f := hclwrite.NewEmptyFile()
	// Empty outputs for now; parser or handlers can add outputs later
	return f.Bytes()
}

// TfvarsFromMetadata generates terraform.tfvars from diagram metadata.
func TfvarsFromMetadata(m *diagram.Metadata) []byte {
	if m == nil {
		return nil
	}
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.SetAttributeValue("aws_region", cty.StringVal("us-east-1"))
	if m.Environment != "" {
		// Could add more metadata-driven vars here
	}
	return f.Bytes()
}

// varTraversal builds hcl.Traversal for var.name (e.g. var.aws_region).
func varTraversal(name string) hcl.Traversal {
	return hcl.Traversal{
		&hcl.TraverseRoot{Name: "var"},
		&hcl.TraverseAttr{Name: name},
	}
}
