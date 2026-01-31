# JSON-to-Terraform Parser

A Go tool that parses architecture-diagram JSON into deployable Terraform (AWS) configurations. Built for performance, concurrency, and cloud-native deployment.

## Features

- **Flat Terraform output**: Generates `main.tf`, `variables.tf`, `versions.tf`, `outputs.tf`, and optional `terraform.tfvars`
- **Concurrent processing**: Validates and generates HCL for independent nodes in parallel (by dependency tier)
- **Extensible handlers**: Registry-based handlers for EC2, Lambda, VPC, Subnet, Security Group, S3, RDS
- **Dependency resolution**: Topological sort from diagram edges (`contains`, `connects_to`, `depends_on`); cycle detection
- **Structured errors**: AGENTS.md-style errors and warnings (JSON or human-readable)

## Build

```bash
go mod tidy
go build -o json2tf ./cmd/parser
```

## Usage

```bash
# Generate Terraform into ./output (with terraform.tfvars)
./json2tf -input diagram.json -o output

# From stdin, no tfvars
./json2tf -input - -o out -no-tfvars

# Output errors as JSON
./json2tf -input diagram.json -o out -json
```

### Flags

| Flag        | Description                                      |
|------------|--------------------------------------------------|
| `-input`   | Path to diagram JSON file, or `-` for stdin      |
| `-o`       | Output directory (default: `output`)            |
| `-no-tfvars` | Do not generate `terraform.tfvars`             |
| `-parallel`  | Max parallel nodes per tier (0 = auto)         |
| `-json`    | Emit errors/warnings as JSON                     |

## Input format

See [AGENTS.md](AGENTS.md) for the full JSON schema. Minimal example:

```json
{
  "metadata": { "version": "1.0", "name": "my-infra", "environment": "production" },
  "nodes": [
    {
      "id": "node-1",
      "type": "ec2_instance",
      "label": "Web Server",
      "position": { "x": 100, "y": 200 },
      "properties": {
        "instance_type": "t3.micro",
        "ami": "ami-xxx",
        "key_name": "my-key",
        "tags": { "Name": "WebServer" }
      }
    }
  ],
  "edges": []
}
```

Supported node types: `vpc`, `subnet`, `security_group`, `ec2_instance`, `lambda_function`, `s3_bucket`, `rds_instance`.

## Project structure

- `cmd/parser` – CLI entry point
- `internal/diagram` – Diagram structs and schema validation
- `internal/parser` – Parser orchestrator and options
- `internal/registry` – Handler registry
- `internal/handler` – Resource handlers (VPC, EC2, Lambda, etc.)
- `internal/dependency` – Graph and topological sort
- `internal/terraform` – Terraform builder and HCL helpers
- `internal/result` – Parse result and error types
- `internal/logger` – Structured logging
- `testdata/` – Sample diagram JSON files

## Programmatic use

```go
import (
    "encoding/json"
    "github.com/json-to-terraform/parser/internal/diagram"
    "github.com/json-to-terraform/parser/internal/parser"
)

data, _ := os.ReadFile("diagram.json")
var d diagram.Diagram
json.Unmarshal(data, &d)

p := parser.New(parser.DefaultOptions())
result, _ := p.Parse(&d)
if result.Success {
    for name, content := range result.TerraformFiles {
        os.WriteFile(name, content, 0644)
    }
} else {
    for _, e := range result.Errors {
        fmt.Println(e.Message)
    }
}
```

## License

See repository license.
