package handler

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/registry"
	"github.com/json-to-terraform/parser/internal/result"
	"github.com/json-to-terraform/parser/internal/terraform"
	"github.com/zclconf/go-cty/cty"
)

type securityGroupHandler struct{}

func init() {
	registry.Default.Register("security_group", &securityGroupHandler{})
}

func (securityGroupHandler) ResourceType() string { return "security_group" }

func (securityGroupHandler) Validate(node *diagram.Node) ([]result.Error, []result.Warning) {
	var errs []result.Error
	var warns []result.Warning
	p := node.Properties
	if diagram.GetStr(p, "name") == "" && node.Label == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "name or label is required", Suggestion: "Set properties.name or node.label",
		})
	}
	return errs, warns
}

func (securityGroupHandler) GenerateHCL(node *diagram.Node, d *diagram.Diagram, refs RefMap) ([]byte, error) {
	name := terraform.SanitizeName(node.ID)
	block := terraform.ResourceBlock("aws_security_group", name)
	body := block.Body()

	p := node.Properties
	sgName := diagram.GetStr(p, "name")
	if sgName == "" {
		sgName = node.Label
	}
	terraform.SetAttributeStr(body, "name", sgName)
	terraform.SetAttributeStr(body, "description", diagram.GetStr(p, "description"))

	for _, e := range d.EdgesWithTarget(node.ID) {
		if e.Type == "contains" {
			if addr, ok := refs[e.Source]; ok {
				body.SetAttributeTraversal("vpc_id", refTraversal(addr, "id"))
				break
			}
		}
	}

	// Ingress/egress: simplified as list of blocks from properties
	if rules, ok := p["ingress"].([]any); ok && len(rules) > 0 {
		for _, r := range rules {
			rm, _ := r.(map[string]any)
			if rm == nil {
				continue
			}
			ing := body.AppendNewBlock("ingress", nil)
			ingBody := ing.Body()
			if v, ok := rm["from_port"].(float64); ok {
				ingBody.SetAttributeValue("from_port", cty.NumberIntVal(int64(v)))
			}
			if v, ok := rm["to_port"].(float64); ok {
				ingBody.SetAttributeValue("to_port", cty.NumberIntVal(int64(v)))
			}
			if v, ok := rm["protocol"].(string); ok {
				ingBody.SetAttributeValue("protocol", cty.StringVal(v))
			}
			if v, ok := rm["cidr_blocks"].([]any); ok && len(v) > 0 {
				var list []cty.Value
				for _, c := range v {
					if s, ok := c.(string); ok {
						list = append(list, cty.StringVal(s))
					}
				}
				if len(list) > 0 {
					ingBody.SetAttributeValue("cidr_blocks", cty.ListVal(list))
				}
			}
		}
	}
	if rules, ok := p["egress"].([]any); ok && len(rules) > 0 {
		for _, r := range rules {
			rm, _ := r.(map[string]any)
			if rm == nil {
				continue
			}
			eg := body.AppendNewBlock("egress", nil)
			egBody := eg.Body()
			if v, ok := rm["from_port"].(float64); ok {
				egBody.SetAttributeValue("from_port", cty.NumberIntVal(int64(v)))
			}
			if v, ok := rm["to_port"].(float64); ok {
				egBody.SetAttributeValue("to_port", cty.NumberIntVal(int64(v)))
			}
			if v, ok := rm["protocol"].(string); ok {
				egBody.SetAttributeValue("protocol", cty.StringVal(v))
			}
			if v, ok := rm["cidr_blocks"].([]any); ok && len(v) > 0 {
				var list []cty.Value
				for _, c := range v {
					if s, ok := c.(string); ok {
						list = append(list, cty.StringVal(s))
					}
				}
				if len(list) > 0 {
					egBody.SetAttributeValue("cidr_blocks", cty.ListVal(list))
				}
			}
		}
	}
	// Default egress if none
	if _, hasEgress := p["egress"]; !hasEgress {
		eg := body.AppendNewBlock("egress", nil)
		eg.Body().SetAttributeValue("from_port", cty.NumberIntVal(0))
		eg.Body().SetAttributeValue("to_port", cty.NumberIntVal(0))
		eg.Body().SetAttributeValue("protocol", cty.StringVal("-1"))
		eg.Body().SetAttributeValue("cidr_blocks", cty.ListVal([]cty.Value{cty.StringVal("0.0.0.0/0")}))
	}

	tags := diagram.GetStrMap(p, "tags")
	if node.Label != "" && (tags == nil || tags["Name"] == "") {
		if tags == nil {
			tags = make(map[string]string)
		}
		tags["Name"] = node.Label
	}
	terraform.SetAttributeMap(body, "tags", tags)

	f := hclwrite.NewEmptyFile()
	f.Body().AppendBlock(block)
	return f.Bytes(), nil
}
