package handler

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/registry"
	"github.com/json-to-terraform/parser/internal/result"
	"github.com/json-to-terraform/parser/internal/terraform"
)

type vpcHandler struct{}

func init() {
	registry.Default.Register("vpc", &vpcHandler{})
}

func (vpcHandler) ResourceType() string { return "vpc" }

func (vpcHandler) Validate(node *diagram.Node) ([]result.Error, []result.Warning) {
	var errs []result.Error
	var warns []result.Warning
	p := node.Properties
	if diagram.GetStr(p, "cidr_block") == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "cidr_block is required", Suggestion: "Set properties.cidr_block (e.g. 10.0.0.0/16)",
		})
	}
	return errs, warns
}

func (vpcHandler) GenerateHCL(node *diagram.Node, d *diagram.Diagram, refs RefMap) ([]byte, error) {
	name := terraform.SanitizeName(node.ID)
	block := terraform.ResourceBlock("aws_vpc", name)
	body := block.Body()

	p := node.Properties
	terraform.SetAttributeStr(body, "cidr_block", diagram.GetStr(p, "cidr_block"))
	terraform.SetAttributeBool(body, "enable_dns_hostnames", diagram.GetBool(p, "enable_dns_hostnames"))
	terraform.SetAttributeBool(body, "enable_dns_support", diagram.GetBool(p, "enable_dns_support"))

	tags := diagram.GetStrMap(p, "tags")
	if node.Label != "" {
		if tags == nil {
			tags = make(map[string]string)
		}
		if _, has := tags["Name"]; !has {
			tags["Name"] = node.Label
		}
	}
	terraform.SetAttributeMap(body, "tags", tags)

	f := hclwrite.NewEmptyFile()
	f.Body().AppendBlock(block)
	return f.Bytes(), nil
}
