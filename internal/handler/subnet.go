package handler

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/registry"
	"github.com/json-to-terraform/parser/internal/result"
	"github.com/json-to-terraform/parser/internal/terraform"
)

type subnetHandler struct{}

func init() {
	registry.Default.Register("subnet", &subnetHandler{})
}

func (subnetHandler) ResourceType() string { return "subnet" }

func (subnetHandler) Validate(node *diagram.Node) ([]result.Error, []result.Warning) {
	var errs []result.Error
	var warns []result.Warning
	p := node.Properties
	if diagram.GetStr(p, "cidr_block") == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "cidr_block is required", Suggestion: "Set properties.cidr_block",
		})
	}
	return errs, warns
}

func (subnetHandler) GenerateHCL(node *diagram.Node, d *diagram.Diagram, refs RefMap) ([]byte, error) {
	name := terraform.SanitizeName(node.ID)
	block := terraform.ResourceBlock("aws_subnet", name)
	body := block.Body()

	p := node.Properties
	terraform.SetAttributeStr(body, "cidr_block", diagram.GetStr(p, "cidr_block"))
	terraform.SetAttributeStr(body, "availability_zone", diagram.GetStr(p, "availability_zone"))

	// vpc_id from "contains" edge: source is VPC (refs store "aws_vpc.node_3")
	for _, e := range d.EdgesWithTarget(node.ID) {
		if e.Type == "contains" {
			if addr, ok := refs[e.Source]; ok {
				body.SetAttributeTraversal("vpc_id", refTraversal(addr, "id"))
				break
			}
		}
	}

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
