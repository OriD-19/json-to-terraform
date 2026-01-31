# AGENTS.md - Visual Infrastructure Builder

## Project Overview

This project is an interactive drag-and-drop infrastructure deployment tool that enables non-technical users to design and deploy AWS infrastructure through visual diagrams. The system converts custom JSON diagram representations into executable Terraform scripts.

## Current Phase: JSON-to-Terraform Parser

The immediate goal is to build the core parsing engine that translates visual diagram data into valid Terraform configurations.

---

## Architecture Principles (Priority Order)

### 1. Extensibility (HIGHEST PRIORITY)
- The parser must support easy addition of new AWS resource types without major refactoring
- Use plugin/provider architecture for resource handlers
- Each resource type should be independent and self-contained
- Abstract interfaces should define contracts for all resource parsers
- New components should be addable by simply creating new handler classes/modules

### 2. Configurability
- Resource properties must be fully configurable through the JSON schema
- Support for environment-specific configurations (dev, staging, prod)
- Terraform variable generation for dynamic values
- Template-based output generation for flexibility

### 3. Availability
- Robust error handling with meaningful error messages
- Validation at multiple stages (JSON schema, resource relationships, Terraform syntax)
- Graceful degradation when optional components fail
- Comprehensive logging for debugging and monitoring

---

## Technical Specifications

### Input Format: Custom JSON Schema

The parser receives a JSON structure representing the infrastructure diagram:

```json
{
  "metadata": {
    "version": "1.0",
    "name": "my-infrastructure",
    "description": "Sample infrastructure diagram",
    "environment": "production"
  },
  "nodes": [
    {
      "id": "node-1",
      "type": "ec2_instance",
      "label": "Web Server",
      "position": {"x": 100, "y": 200},
      "properties": {
        "instance_type": "t3.micro",
        "ami": "ami-0c55b159cbfafe1f0",
        "key_name": "my-key",
        "tags": {
          "Name": "WebServer",
          "Environment": "production"
        }
      }
    },
    {
      "id": "node-2",
      "type": "lambda_function",
      "label": "Data Processor",
      "position": {"x": 300, "y": 200},
      "properties": {
        "runtime": "python3.9",
        "handler": "index.handler",
        "memory_size": 256,
        "timeout": 60,
        "environment_variables": {
          "STAGE": "prod"
        }
      }
    },
    {
      "id": "node-3",
      "type": "vpc",
      "label": "Main VPC",
      "position": {"x": 200, "y": 100},
      "properties": {
        "cidr_block": "10.0.0.0/16",
        "enable_dns_hostnames": true,
        "enable_dns_support": true
      }
    }
  ],
  "edges": [
    {
      "id": "edge-1",
      "source": "node-3",
      "target": "node-1",
      "type": "contains",
      "properties": {
        "subnet_id": "subnet-xxx"
      }
    }
  ]
}
```

### Output Format: Terraform HCL

The parser generates valid Terraform configuration files:

```hcl
# main.tf
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

resource "aws_vpc" "node_3" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = {
    Name = "Main VPC"
  }
}

resource "aws_instance" "node_1" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.micro"
  key_name      = "my-key"
  
  tags = {
    Name        = "WebServer"
    Environment = "production"
  }
}

# ... more resources
```

---

## Required Components

### 1. Core Parser Module
- **Responsibility**: Main orchestration and workflow management
- **Key Functions**:
  - Load and validate input JSON
  - Coordinate resource handlers
  - Manage dependency resolution
  - Generate output Terraform files
  - Error aggregation and reporting

### 2. Resource Handler Registry (EXTENSIBILITY FOCUS)
- **Responsibility**: Dynamic registration and lookup of resource parsers
- **Requirements**:
  - Registry pattern for resource type handlers
  - Auto-discovery of handler modules/plugins
  - Version compatibility checking
  - Handler metadata (supported properties, required fields, validation rules)

### 3. Individual Resource Handlers (EXTENSIBILITY FOCUS)
Start with these AWS resource types:
- **EC2 Instance** (`ec2_instance`)
- **Lambda Function** (`lambda_function`)
- **VPC** (`vpc`)
- **Subnet** (`subnet`)
- **Security Group** (`security_group`)
- **S3 Bucket** (`s3_bucket`)
- **RDS Instance** (`rds_instance`)

Each handler must implement:
- Property validation
- Terraform resource block generation
- Dependency identification
- Default value handling
- Custom validation rules

### 4. Relationship/Dependency Resolver
- **Responsibility**: Process edges and establish Terraform dependencies
- **Requirements**:
  - Parse edge relationships (contains, connects_to, depends_on)
  - Generate Terraform resource references
  - Detect circular dependencies
  - Topological sorting for resource creation order

### 5. Validation Engine
- **Responsibility**: Multi-stage validation
- **Stages**:
  - JSON schema validation
  - Resource property validation
  - Relationship validation
  - Terraform syntax validation (using `terraform validate`)
  - AWS-specific validation (CIDR blocks, resource limits, etc.)

### 6. Template Generator (CONFIGURABILITY FOCUS)
- **Responsibility**: Generate Terraform files with proper structure
- **Outputs**:
  - `main.tf` - Resource definitions
  - `variables.tf` - Input variables
  - `outputs.tf` - Output values
  - `versions.tf` - Provider requirements
  - `terraform.tfvars` - Variable values (optional)

---

## Implementation Guidelines

### Language Recommendation
Python is recommended for this phase due to:
- Excellent JSON handling
- Rich ecosystem for validation (jsonschema, pydantic)
- HCL generation libraries (python-hcl2)
- Easy plugin/module system
- Strong AWS SDK (boto3) for future validation

### Project Structure
```
parser/
├── core/
│   ├── __init__.py
│   ├── parser.py          # Main parser orchestrator
│   ├── registry.py        # Resource handler registry
│   └── validator.py       # Validation engine
├── handlers/
│   ├── __init__.py
│   ├── base.py            # Abstract base handler class
│   ├── ec2.py             # EC2 instance handler
│   ├── lambda_fn.py       # Lambda function handler
│   ├── vpc.py             # VPC handler
│   ├── subnet.py          # Subnet handler
│   ├── security_group.py  # Security group handler
│   └── ...                # Additional handlers
├── templates/
│   ├── main.tf.j2         # Jinja2 template for main.tf
│   ├── variables.tf.j2
│   └── outputs.tf.j2
├── schemas/
│   ├── diagram_schema.json # JSON schema for input validation
│   └── resource_schemas/   # Individual resource schemas
├── utils/
│   ├── dependency.py      # Dependency resolution
│   ├── terraform.py       # Terraform utilities
│   └── logger.py          # Logging configuration
├── tests/
│   ├── test_parser.py
│   ├── test_handlers.py
│   └── fixtures/          # Test JSON diagrams
└── main.py                # CLI entry point
```

### Key Design Patterns

#### 1. Strategy Pattern (for Resource Handlers)
Each resource type implements a common interface:

```python
from abc import ABC, abstractmethod

class ResourceHandler(ABC):
    @abstractmethod
    def validate(self, properties: dict) -> tuple[bool, list[str]]:
        """Validate resource properties"""
        pass
    
    @abstractmethod
    def generate_terraform(self, node: dict) -> str:
        """Generate Terraform HCL for this resource"""
        pass
    
    @abstractmethod
    def get_dependencies(self, node: dict, edges: list) -> list[str]:
        """Identify dependencies for this resource"""
        pass
    
    @property
    @abstractmethod
    def resource_type(self) -> str:
        """Return the AWS resource type identifier"""
        pass
```

#### 2. Registry Pattern (for Handler Discovery)
```python
class HandlerRegistry:
    _handlers = {}
    
    @classmethod
    def register(cls, resource_type: str, handler_class):
        """Register a handler for a resource type"""
        cls._handlers[resource_type] = handler_class
    
    @classmethod
    def get_handler(cls, resource_type: str) -> ResourceHandler:
        """Get handler instance for a resource type"""
        if resource_type not in cls._handlers:
            raise ValueError(f"No handler registered for {resource_type}")
        return cls._handlers[resource_type]()
    
    @classmethod
    def list_supported_types(cls) -> list[str]:
        """List all supported resource types"""
        return list(cls._handlers.keys())
```

#### 3. Builder Pattern (for Terraform Generation)
```python
class TerraformBuilder:
    def __init__(self):
        self.resources = []
        self.variables = []
        self.outputs = []
    
    def add_resource(self, resource_block: str):
        self.resources.append(resource_block)
        return self
    
    def add_variable(self, var_block: str):
        self.variables.append(var_block)
        return self
    
    def build(self) -> dict[str, str]:
        """Build complete Terraform configuration"""
        return {
            'main.tf': '\n\n'.join(self.resources),
            'variables.tf': '\n\n'.join(self.variables),
            # ... other files
        }
```

---

## Error Handling Requirements (AVAILABILITY FOCUS)

### Error Categories
1. **Schema Errors**: Invalid JSON structure
2. **Validation Errors**: Invalid property values
3. **Dependency Errors**: Circular dependencies, missing references
4. **Generation Errors**: Terraform syntax issues

### Error Response Format
```json
{
  "success": false,
  "errors": [
    {
      "type": "validation_error",
      "severity": "error",
      "node_id": "node-1",
      "message": "Invalid instance_type: 't3.invalid'",
      "suggestion": "Valid instance types: t3.micro, t3.small, t3.medium, ..."
    }
  ],
  "warnings": [
    {
      "type": "best_practice",
      "severity": "warning",
      "node_id": "node-2",
      "message": "Lambda function timeout is high (300s)",
      "suggestion": "Consider reducing timeout to avoid unnecessary costs"
    }
  ]
}
```

### Logging Requirements
- Structured logging (JSON format preferred)
- Log levels: DEBUG, INFO, WARNING, ERROR, CRITICAL
- Include context: node_id, resource_type, operation
- Performance metrics: parsing time, validation time, generation time

---

## Testing Requirements

### Unit Tests
- Each resource handler must have dedicated tests
- Test valid and invalid property combinations
- Test edge cases (missing required fields, extra fields, etc.)

### Integration Tests
- End-to-end parsing of complete diagrams
- Multi-resource dependencies
- Complex network topologies

### Test Fixtures
Create sample JSON diagrams for:
- Simple single EC2 instance
- VPC with subnets and instances
- Serverless architecture (Lambda + API Gateway + DynamoDB)
- Complex multi-tier application

### Validation Tests
- Valid Terraform output (run `terraform validate`)
- Idempotency (parsing same JSON produces same output)

---

## Performance Considerations

### Current Phase
- Target: Parse diagrams with 50+ nodes in < 2 seconds
- Memory: Keep memory usage under 500MB for typical diagrams
- Scalability: Design for eventual support of 200+ node diagrams

### Future Optimization Areas
- Parallel resource handler execution
- Caching of validation results
- Incremental parsing (only re-parse changed nodes)

---

## Security Considerations

### Input Validation
- Sanitize all user inputs
- Prevent command injection in Terraform generation
- Validate AWS ARNs and resource identifiers
- Limit JSON size to prevent DoS attacks

### Output Security
- Never include sensitive data in generated Terraform
- Use Terraform variables for secrets
- Include warnings for insecure configurations (e.g., public S3 buckets)

---

## Future Extensibility Hooks

Design with these future features in mind:

1. **Multi-Cloud Support**
   - Abstract resource handlers to support Azure, GCP
   - Provider registry pattern

2. **Custom Resource Types**
   - User-defined resource handlers
   - Plugin system for third-party handlers

3. **Validation Plugins**
   - Custom validation rules
   - Organization-specific policy enforcement

4. **Output Formats**
   - CloudFormation support
   - Pulumi support
   - AWS CDK support

---

## Success Criteria

### Milestone 1: Core Parser (Current)
- [ ] Parse basic JSON diagram into Terraform
- [ ] Support EC2, Lambda, VPC, Subnet, Security Group
- [ ] Basic validation (schema + properties)
- [ ] Handle simple dependencies
- [ ] Generate valid `main.tf` file

### Milestone 2: Enhanced Validation
- [ ] Multi-stage validation pipeline
- [ ] Comprehensive error messages
- [ ] Circular dependency detection
- [ ] Terraform syntax validation

### Milestone 3: Production Ready
- [ ] Support 10+ AWS resource types
- [ ] Full test coverage (>80%)
- [ ] Performance benchmarks met
- [ ] Documentation complete
- [ ] CLI interface for testing

---

## Development Workflow

1. **Setup Phase**
   - Initialize project structure
   - Set up testing framework
   - Create base classes and interfaces

2. **Core Development**
   - Implement HandlerRegistry
   - Create base ResourceHandler class
   - Build validation engine
   - Develop TerraformBuilder

3. **Handler Implementation**
   - Start with EC2 (simplest)
   - Add VPC and networking components
   - Implement serverless resources
   - Test each handler individually

4. **Integration**
   - Build dependency resolver
   - Implement main parser orchestrator
   - Create end-to-end tests
   - Performance testing

5. **Refinement**
   - Error message improvements
   - Documentation
   - Code cleanup and optimization
   - Security review

---

## Example Usage (Target API)

```python
from parser import InfrastructureParser

# Initialize parser
parser = InfrastructureParser()

# Load diagram JSON
with open('diagram.json') as f:
    diagram = json.load(f)

# Parse and validate
result = parser.parse(diagram)

if result.success:
    # Write Terraform files
    for filename, content in result.terraform_files.items():
        with open(f'output/{filename}', 'w') as f:
            f.write(content)
    print("Terraform files generated successfully!")
else:
    # Display errors
    for error in result.errors:
        print(f"ERROR: {error.message}")
```

---

## Questions for Clarification

Before implementation, consider these questions:

1. Should the parser support Terraform modules or only resources?
2. How should we handle Terraform state management (local, S3 backend)?
3. Should we generate `terraform.tfvars` or expect external configuration?
4. Do you want support for Terraform workspaces (dev/staging/prod)?
5. Should handlers validate against AWS quotas/limits?
6. Do you need support for Terraform data sources (existing infrastructure)?

---

## Additional Resources

- Terraform AWS Provider Documentation: https://registry.terraform.io/providers/hashicorp/aws/latest/docs
- Terraform Language Documentation: https://www.terraform.io/language
- JSON Schema Specification: https://json-schema.org/
- Python HCL2 Library: https://github.com/amplify-education/python-hcl2

---

## Notes for the Coding Agent

- **Prioritize extensibility first** - make it easy to add new resource types
- Use type hints throughout the codebase (Python 3.9+)
- Write docstrings for all public methods
- Include inline comments for complex logic
- Keep functions small and focused (single responsibility)
- Prefer composition over inheritance
- Make validation strict but informative
- Generate human-readable error messages
- Test each component independently before integration