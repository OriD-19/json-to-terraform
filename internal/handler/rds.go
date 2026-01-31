package handler

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/registry"
	"github.com/json-to-terraform/parser/internal/result"
	"github.com/json-to-terraform/parser/internal/terraform"
	"github.com/zclconf/go-cty/cty"
)

type rdsHandler struct{}

func init() {
	registry.Default.Register("rds_instance", &rdsHandler{})
}

func (rdsHandler) ResourceType() string { return "rds_instance" }

func (rdsHandler) Validate(node *diagram.Node) ([]result.Error, []result.Warning) {
	var errs []result.Error
	var warns []result.Warning
	p := node.Properties
	if diagram.GetStr(p, "engine") == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "engine is required", Suggestion: "Set properties.engine (e.g. postgres)",
		})
	}
	if diagram.GetStr(p, "instance_class") == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "instance_class is required", Suggestion: "Set properties.instance_class (e.g. db.t3.micro)",
		})
	}
	if diagram.GetInt(p, "allocated_storage") == 0 {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "allocated_storage is required", Suggestion: "Set properties.allocated_storage (GB)",
		})
	}
	return errs, warns
}

func (rdsHandler) GenerateHCL(node *diagram.Node, d *diagram.Diagram, refs RefMap) ([]byte, error) {
	name := terraform.SanitizeName(node.ID)
	block := terraform.ResourceBlock("aws_db_instance", name)
	body := block.Body()

	p := node.Properties
	terraform.SetAttributeStr(body, "engine", diagram.GetStr(p, "engine"))
	terraform.SetAttributeStr(body, "engine_version", diagram.GetStr(p, "engine_version"))
	terraform.SetAttributeStr(body, "instance_class", diagram.GetStr(p, "instance_class"))
	terraform.SetAttributeInt(body, "allocated_storage", diagram.GetInt(p, "allocated_storage"))
	terraform.SetAttributeStr(body, "storage_type", diagram.GetStr(p, "storage_type"))
	terraform.SetAttributeStr(body, "db_name", diagram.GetStr(p, "db_name"))
	terraform.SetAttributeStr(body, "username", diagram.GetStr(p, "username"))
	if pw := diagram.GetStr(p, "password"); pw != "" {
		body.SetAttributeValue("password", cty.StringVal(pw))
	}
	if diagram.GetBool(p, "skip_final_snapshot") {
		body.SetAttributeValue("skip_final_snapshot", cty.BoolVal(true))
	}
	if n := diagram.GetInt(p, "backup_retention_period"); n > 0 {
		terraform.SetAttributeInt(body, "backup_retention_period", n)
	}
	terraform.SetAttributeBool(body, "multi_az", diagram.GetBool(p, "multi_az"))

	// db_subnet_group_name from "contains" edge (source = db_subnet_group); vpc_security_group_ids from "connects_to" (source = security_group)
	var sgRefs []string
	for _, e := range d.EdgesWithTarget(node.ID) {
		if e.Type == "contains" {
			sourceNode := d.NodeByID(e.Source)
			if sourceNode != nil && sourceNode.Type == "db_subnet_group" {
				if addr, ok := refs[e.Source]; ok {
					body.SetAttributeTraversal("db_subnet_group_name", refTraversal(addr, "name"))
				}
				break
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
