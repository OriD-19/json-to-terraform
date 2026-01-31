package handler

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/registry"
	"github.com/json-to-terraform/parser/internal/result"
	"github.com/json-to-terraform/parser/internal/terraform"
	"github.com/zclconf/go-cty/cty"
)

type lambdaHandler struct{}

func init() {
	registry.Default.Register("lambda_function", &lambdaHandler{})
}

func (lambdaHandler) ResourceType() string { return "lambda_function" }

func (lambdaHandler) Validate(node *diagram.Node) ([]result.Error, []result.Warning) {
	var errs []result.Error
	var warns []result.Warning
	p := node.Properties
	if diagram.GetStr(p, "runtime") == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "runtime is required", Suggestion: "Set properties.runtime (e.g. python3.9)",
		})
	}
	if diagram.GetStr(p, "handler") == "" {
		errs = append(errs, result.Error{
			Type: "validation_error", Severity: "error", NodeID: node.ID,
			Message: "handler is required", Suggestion: "Set properties.handler (e.g. index.handler)",
		})
	}
	return errs, warns
}

func (lambdaHandler) GenerateHCL(node *diagram.Node, d *diagram.Diagram, refs RefMap) ([]byte, error) {
	name := terraform.SanitizeName(node.ID)
	block := terraform.ResourceBlock("aws_lambda_function", name)
	body := block.Body()

	p := node.Properties
	terraform.SetAttributeStr(body, "runtime", diagram.GetStr(p, "runtime"))
	terraform.SetAttributeStr(body, "handler", diagram.GetStr(p, "handler"))
	mem := diagram.GetInt(p, "memory_size")
	if mem == 0 {
		mem = 128
	}
	terraform.SetAttributeInt(body, "memory_size", mem)
	timeout := diagram.GetInt(p, "timeout")
	if timeout == 0 {
		timeout = 3
	}
	terraform.SetAttributeInt(body, "timeout", timeout)
	terraform.SetAttributeStr(body, "filename", diagram.GetStr(p, "filename"))
	fnName := diagram.GetStr(p, "function_name")
	if fnName == "" && node.Label != "" {
		fnName = node.Label
	}
	terraform.SetAttributeStr(body, "function_name", fnName)

	env := diagram.GetStrMap(p, "environment_variables")
	if len(env) > 0 {
		envBlock := body.AppendNewBlock("environment", nil)
		envBody := envBlock.Body()
		ctyVars := make(map[string]cty.Value)
		for k, v := range env {
			ctyVars[k] = cty.StringVal(v)
		}
		envBody.SetAttributeValue("variables", cty.MapVal(ctyVars))
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
