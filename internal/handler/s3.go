package handler

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/registry"
	"github.com/json-to-terraform/parser/internal/result"
	"github.com/json-to-terraform/parser/internal/terraform"
	"github.com/zclconf/go-cty/cty"
)

type s3Handler struct{}

func init() {
	registry.Default.Register("s3_bucket", &s3Handler{})
}

func (s3Handler) ResourceType() string { return "s3_bucket" }

func (s3Handler) Validate(node *diagram.Node) ([]result.Error, []result.Warning) {
	var errs []result.Error
	var warns []result.Warning
	p := node.Properties
	if diagram.GetStr(p, "bucket") == "" && node.Label == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "bucket name or label is required", Suggestion: "Set properties.bucket or node.label",
		})
	}
	return errs, warns
}

func (s3Handler) GenerateHCL(node *diagram.Node, d *diagram.Diagram, refs RefMap) ([]byte, error) {
	name := terraform.SanitizeName(node.ID)
	block := terraform.ResourceBlock("aws_s3_bucket", name)
	body := block.Body()

	p := node.Properties
	bucketName := diagram.GetStr(p, "bucket")
	if bucketName == "" {
		bucketName = node.Label
	}
	terraform.SetAttributeStr(body, "bucket", bucketName)

	if diagram.GetBool(p, "versioning") {
		ver := body.AppendNewBlock("versioning", nil)
		ver.Body().SetAttributeValue("enabled", cty.BoolVal(true))
	}
	if diagram.GetBool(p, "block_public_acls") {
		acl := body.AppendNewBlock("public_access_block", nil)
		acl.Body().SetAttributeValue("block_public_acls", cty.BoolVal(true))
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
