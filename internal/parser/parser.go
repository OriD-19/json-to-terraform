package parser

import (
	"runtime"
	"sync"

	"github.com/json-to-terraform/parser/internal/dependency"
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/registry"
	"github.com/json-to-terraform/parser/internal/result"
	"github.com/json-to-terraform/parser/internal/terraform"
)

// InfrastructureParser parses diagram JSON into Terraform files.
type InfrastructureParser struct {
	opts    Options
	reg     *registry.Registry
	builder *terraform.TerraformBuilder
}

// New returns a new parser with the given options.
func New(opts Options) *InfrastructureParser {
	if opts.MaxParallel <= 0 {
		opts.MaxParallel = runtime.NumCPU()
	}
	if opts.MaxParallel > 32 {
		opts.MaxParallel = 32
	}
	return &InfrastructureParser{
		opts: opts,
		reg:  registry.Default,
	}
}

// Parse validates the diagram, resolves dependencies, and generates Terraform files.
func (p *InfrastructureParser) Parse(d *diagram.Diagram) (*result.ParseResult, error) {
	out := &result.ParseResult{Success: true}

	// 1. Diagram-level validation
	diagErrs := diagram.Validate(d)
	for _, e := range diagErrs {
		out.Errors = append(out.Errors, result.Error{
			Type: e.Type, Severity: e.Severity, NodeID: e.NodeID,
			Message: e.Message, Suggestion: e.Suggestion,
		})
	}
	if len(out.Errors) > 0 {
		out.Success = false
		return out, nil
	}

	// 2. Resolve dependency order and tiers
	ordered, tiers, err := dependency.Resolve(d)
	if err != nil {
		out.Success = false
		out.Errors = append(out.Errors, result.Error{
			Type: "dependency_error", Severity: "error",
			Message: err.Error(), Suggestion: "Remove circular edges or fix node references",
		})
		return out, nil
	}

	// 3. Build ref map and collect resource blocks in order
	refs := make(registry.RefMap)
	resourceBlocks := make([][]byte, 0, len(ordered))
	nodeByID := make(map[string]*diagram.Node)
	for i := range d.Nodes {
		nodeByID[d.Nodes[i].ID] = &d.Nodes[i]
	}

	// Process tier by tier; within each tier run handlers in parallel
	for _, tier := range tiers {
		var wg sync.WaitGroup
		var mu sync.Mutex
		type nodeResult struct {
			nodeID string
			hcl    []byte
			errs   []result.Error
			warns  []result.Warning
		}
		results := make([]nodeResult, 0, len(tier))

		for _, nodeID := range tier {
			node := nodeByID[nodeID]
			if node == nil {
				continue
			}
			h, ok := p.reg.Get(node.Type)
			if !ok {
				mu.Lock()
				out.Errors = append(out.Errors, result.Error{
					Type: "validation_error", Severity: "error", NodeID: nodeID,
					Message: "unsupported resource type: " + node.Type,
					Suggestion: "Use one of: vpc, subnet, security_group, ec2_instance, lambda_function, s3_bucket, rds_instance",
				})
				out.Success = false
				mu.Unlock()
				continue
			}

			wg.Add(1)
			go func(n *diagram.Node) {
				defer wg.Done()
				verrs, vwarns := h.Validate(n)
				hcl, genErr := h.GenerateHCL(n, d, refs)
				mu.Lock()
				res := nodeResult{nodeID: n.ID, errs: verrs, warns: vwarns}
				if genErr != nil {
					res.errs = append(res.errs, result.Error{
						Type: "generation_error", Severity: "error", NodeID: n.ID,
						Message: genErr.Error(),
					})
				} else {
					res.hcl = hcl
				}
				results = append(results, res)
				mu.Unlock()
			}(node)
		}
		wg.Wait()

		resultsByID := make(map[string]nodeResult)
		for _, res := range results {
			out.Errors = append(out.Errors, res.errs...)
			out.Warnings = append(out.Warnings, res.warns...)
			if len(res.errs) > 0 {
				out.Success = false
			}
			resultsByID[res.nodeID] = res
		}
		// Append blocks in tier order so main.tf stays in dependency order
		for _, nodeID := range tier {
			res := resultsByID[nodeID]
			if len(res.hcl) > 0 {
				resourceBlocks = append(resourceBlocks, res.hcl)
				node := nodeByID[nodeID]
				if node != nil {
					tfType := terraformResourceType(node.Type)
					name := terraform.SanitizeName(nodeID)
					refs[nodeID] = tfType + "." + name
				}
			}
		}
	}

	if !out.Success {
		return out, nil
	}

	// 4. Build Terraform files
	b := terraform.NewBuilder(p.opts.EmitTfvars)
	b.SetVersions(terraform.VersionsTF())
	b.SetVariables(terraform.VariablesTF())
	b.SetOutputs(terraform.OutputsTF())
	for _, block := range resourceBlocks {
		b.AddResource(block)
	}
	if p.opts.EmitTfvars {
		b.SetTfvars(terraform.TfvarsFromMetadata(&d.Metadata))
	}
	out.TerraformFiles = b.Build()
	return out, nil
}

func terraformResourceType(diagramType string) string {
	m := map[string]string{
		"vpc":             "aws_vpc",
		"subnet":          "aws_subnet",
		"security_group":  "aws_security_group",
		"ec2_instance":    "aws_instance",
		"lambda_function": "aws_lambda_function",
		"s3_bucket":       "aws_s3_bucket",
		"rds_instance":    "aws_db_instance",
	}
	if t, ok := m[diagramType]; ok {
		return t
	}
	return "aws_" + diagramType
}
