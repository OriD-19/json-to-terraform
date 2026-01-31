package handler

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/registry"
	"github.com/json-to-terraform/parser/internal/result"
	"github.com/json-to-terraform/parser/internal/terraform"
)

type ec2Handler struct{}

func init() {
	registry.Default.Register("ec2_instance", &ec2Handler{})
}

func (ec2Handler) ResourceType() string { return "ec2_instance" }

func (ec2Handler) Validate(node *diagram.Node) ([]result.Error, []result.Warning) {
	var errs []result.Error
	var warns []result.Warning
	p := node.Properties
	if diagram.GetStr(p, "ami") == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "ami is required", Suggestion: "Set properties.ami",
		})
	}
	if diagram.GetStr(p, "instance_type") == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "instance_type is required", Suggestion: "Set properties.instance_type (e.g. t3.micro)",
		})
	}
	return errs, warns
}

func (ec2Handler) GenerateHCL(node *diagram.Node, d *diagram.Diagram, refs RefMap) ([]byte, error) {
	name := terraform.SanitizeName(node.ID)
	block := terraform.ResourceBlock("aws_instance", name)
	body := block.Body()

	p := node.Properties
	terraform.SetAttributeStr(body, "ami", diagram.GetStr(p, "ami"))
	terraform.SetAttributeStr(body, "instance_type", diagram.GetStr(p, "instance_type"))
	terraform.SetAttributeStr(body, "key_name", diagram.GetStr(p, "key_name"))

	var sgRefs []string
	for _, e := range d.EdgesWithTarget(node.ID) {
		if e.Type == "contains" {
			sourceNode := d.NodeByID(e.Source)
			if sourceNode != nil && sourceNode.Type == "subnet" {
				if addr, ok := refs[e.Source]; ok {
					body.SetAttributeTraversal("subnet_id", refTraversal(addr, "id"))
				}
			}
		}
		if e.Type == "connects_to" {
			if addr, ok := refs[e.Source]; ok {
				sgRefs = append(sgRefs, addr)
			}
		}
	}
	if len(sgRefs) > 0 {
		tokens := make([]hclwrite.Tokens, len(sgRefs))
		for i, addr := range sgRefs {
			tokens[i] = hclwrite.TokensForTraversal(refTraversal(addr, "id"))
		}
		body.SetAttributeRaw("vpc_security_group_ids", hclwrite.TokensForTuple(tokens))
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
