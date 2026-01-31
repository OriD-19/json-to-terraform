# Component Library — Parser Input Reference

This document describes **every component (node type) and relationship (edge type)** that the JSON-to-Terraform parser currently supports. It is intended for agents or developers building the **frontend drag-and-drop interface**: use it to know which node types exist, which properties each accepts, and how to connect them with edges so the generated JSON is valid and produces the expected Terraform.

---

## 1. Document structure

The diagram is a single JSON object with three top-level keys:

| Key        | Type     | Description |
|-----------|----------|-------------|
| `metadata` | object   | Diagram-level info (version, name, description, environment). |
| `nodes`    | array    | List of resource nodes (VPC, subnet, EC2, Lambda, etc.). |
| `edges`    | array    | List of directed relationships between nodes (e.g. “VPC contains subnet”). |

### 1.1 Metadata (required)

```json
{
  "metadata": {
    "version": "1.0",
    "name": "my-infrastructure",
    "description": "Optional description",
    "environment": "production"
  }
}
```

- **`version`** (required): e.g. `"1.0"`.
- **`name`**, **`description`**, **`environment`**: optional strings; used in generated Terraform (e.g. variables, tags).

### 1.2 Node (common shape)

Every node in `nodes` has this shape:

```json
{
  "id": "unique-node-id",
  "type": "resource_type",
  "label": "Display name",
  "position": { "x": 100, "y": 200 },
  "properties": { }
}
```

- **`id`** (required): Unique string in the diagram; used in `edges` as `source` / `target`. Must be unique across all nodes.
- **`type`** (required): One of the supported component types (see below).
- **`label`** (optional): Human-readable name; often used as default for Terraform `Name` tag or resource naming.
- **`position`** (optional): `{ "x": number, "y": number }` for canvas layout; not used by the parser for Terraform.
- **`properties`** (optional): Object; keys depend on `type`. Can be `{}` if no properties are needed.

### 1.3 Edge (relationship)

Every edge in `edges` has this shape:

```json
{
  "id": "edge-id",
  "source": "source-node-id",
  "target": "target-node-id",
  "type": "contains | connects_to | depends_on",
  "properties": {}
}
```

- **`source`** (required): Node `id` that the edge comes from.
- **`target`** (required): Node `id` that the edge goes to.
- **`type`** (required): Semantics below. **`properties`** is optional and currently used only where noted.

**Edge semantics:**

| Edge type      | Meaning | Typical use |
|----------------|--------|-------------|
| **`contains`** | Parent contains child (e.g. VPC → subnet, subnet → EC2). | Sets `vpc_id`, `subnet_id`, etc. in Terraform. |
| **`connects_to`** | Resource is attached to another (e.g. security group → EC2). | Sets `vpc_security_group_ids`, etc. |
| **`depends_on`** | Creation order only; no direct Terraform reference from this edge. | Dependency resolution / ordering. |

---

## 2. Supported node types and properties

Below, **required** properties must be set (or the parser will report a validation error). **Optional** properties can be omitted; defaults are noted where applicable.

**Tags:** For every component, `properties.tags` is optional: a flat object of string key-value pairs (e.g. `"Name": "my-resource"`). If `tags` is omitted but `label` is set, the parser often uses `label` as the `Name` tag.

---

### 2.1 VPC — `type: "vpc"`

Represents an AWS VPC.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `cidr_block` | string | **Yes** | VPC CIDR (e.g. `"10.0.0.0/16"`). |
| `enable_dns_hostnames` | boolean | No | Default: not set. |
| `enable_dns_support` | boolean | No | Default: not set. |
| `tags` | object | No | String key-value pairs. |

**Edges:** None required. Subnets and security groups link *to* the VPC with **`contains`** (VPC is **source**).

**Sample node:**

```json
{
  "id": "vpc-main",
  "type": "vpc",
  "label": "Main VPC",
  "position": { "x": 400, "y": 80 },
  "properties": {
    "cidr_block": "10.0.0.0/16",
    "enable_dns_hostnames": true,
    "enable_dns_support": true,
    "tags": {
      "Name": "main-vpc",
      "Environment": "production"
    }
  }
}
```

---

### 2.2 Subnet — `type: "subnet"`

Represents an AWS subnet. Must be contained in a VPC via a **`contains`** edge (VPC → subnet).

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `cidr_block` | string | **Yes** | Subnet CIDR (e.g. `"10.0.1.0/24"`). |
| `availability_zone` | string | No | AZ (e.g. `"us-east-1a"`). |
| `map_public_ip_on_launch` | boolean | No | Default: not set. |
| `tags` | object | No | String key-value pairs. |

**Edges:** One **`contains`** edge from a **vpc** node (source = VPC, target = this subnet) so the parser can set `vpc_id`.

**Sample node:**

```json
{
  "id": "subnet-public-1a",
  "type": "subnet",
  "label": "Public Subnet 1a",
  "position": { "x": 200, "y": 220 },
  "properties": {
    "cidr_block": "10.0.1.0/24",
    "availability_zone": "us-east-1a",
    "map_public_ip_on_launch": true,
    "tags": { "Name": "public-1a", "Type": "public" }
  }
}
```

**Sample edge (VPC contains subnet):**

```json
{ "id": "e1", "source": "vpc-main", "target": "subnet-public-1a", "type": "contains" }
```

---

### 2.3 Security group — `type: "security_group"`

Represents an AWS security group. Must be in a VPC via **`contains`** (VPC → security group).

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | **Yes*** | Security group name. *Can be omitted if `label` is set. |
| `description` | string | No | SG description. |
| `ingress` | array | No | List of ingress rule objects (see below). |
| `egress` | array | No | List of egress rule objects. If omitted, a default “all outbound” egress is generated. |
| `tags` | object | No | String key-value pairs. |

**Ingress / egress rule object:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `from_port` | number | Yes | Start port (e.g. 80, 22). Use 0 with protocol `"-1"` for “all”. |
| `to_port` | number | Yes | End port. |
| `protocol` | string | Yes | e.g. `"tcp"`, `"udp"`, `"-1"` (all). |
| `cidr_blocks` | array of strings | Yes | e.g. `["0.0.0.0/0"]`, `["10.0.0.0/16"]`. |
| `description` | string | No | Rule description. |

**Edges:** One **`contains`** edge from a **vpc** node. Other resources (e.g. EC2, RDS) attach via **`connects_to`** (security group → instance).

**Sample node:**

```json
{
  "id": "sg-web",
  "type": "security_group",
  "label": "Web Security Group",
  "position": { "x": 400, "y": 320 },
  "properties": {
    "name": "web-sg",
    "description": "Allow HTTP/HTTPS and SSH",
    "ingress": [
      {
        "from_port": 80,
        "to_port": 80,
        "protocol": "tcp",
        "cidr_blocks": ["0.0.0.0/0"],
        "description": "HTTP"
      },
      {
        "from_port": 22,
        "to_port": 22,
        "protocol": "tcp",
        "cidr_blocks": ["10.0.0.0/16"],
        "description": "SSH from VPC"
      }
    ],
    "egress": [
      {
        "from_port": 0,
        "to_port": 0,
        "protocol": "-1",
        "cidr_blocks": ["0.0.0.0/0"],
        "description": "All outbound"
      }
    ],
    "tags": { "Name": "web-sg" }
  }
}
```

**Sample edges:**

```json
{ "id": "e4", "source": "vpc-main", "target": "sg-web", "type": "contains" },
{ "id": "e8", "source": "sg-web", "target": "ec2-web-1", "type": "connects_to" }
```

---

### 2.4 EC2 instance — `type: "ec2_instance"`

Represents an AWS EC2 instance. Typically in a subnet and attached to one or more security groups.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `ami` | string | **Yes** | AMI ID (e.g. `"ami-0c55b159cbfafe1f0"`). |
| `instance_type` | string | **Yes** | e.g. `"t3.micro"`, `"t3.small"`. |
| `key_name` | string | No | SSH key pair name. |
| `tags` | object | No | String key-value pairs. |

**Edges:**

- **`contains`** from a **subnet** node (subnet → EC2): parser sets `subnet_id`.
- **`connects_to`** from **security_group** node(s) (SG → EC2): parser sets `vpc_security_group_ids`.

**Sample node:**

```json
{
  "id": "ec2-web-1",
  "type": "ec2_instance",
  "label": "Web Server 1",
  "position": { "x": 200, "y": 280 },
  "properties": {
    "ami": "ami-0c55b159cbfafe1f0",
    "instance_type": "t3.small",
    "key_name": "my-key",
    "tags": {
      "Name": "web-server-1",
      "Role": "web",
      "Environment": "production"
    }
  }
}
```

**Sample edges:**

```json
{ "id": "e6", "source": "subnet-public-1a", "target": "ec2-web-1", "type": "contains" },
{ "id": "e8", "source": "sg-web", "target": "ec2-web-1", "type": "connects_to" }
```

---

### 2.5 Lambda function — `type: "lambda_function"`

Represents an AWS Lambda function. No edges are required for Terraform generation.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `runtime` | string | **Yes** | e.g. `"python3.9"`, `"nodejs18.x"`. |
| `handler` | string | **Yes** | e.g. `"index.handler"`. |
| `memory_size` | number | No | Default: 128 (MB). |
| `timeout` | number | No | Default: 3 (seconds). |
| `filename` | string | No | Path to deployment package (often set at deploy time). |
| `function_name` | string | No | If omitted, parser can use `label`. |
| `environment_variables` | object | No | String key-value pairs for Lambda env. |
| `tags` | object | No | String key-value pairs. |

**Edges:** None required.

**Sample node:**

```json
{
  "id": "lambda-processor",
  "type": "lambda_function",
  "label": "Data Processor",
  "position": { "x": 300, "y": 200 },
  "properties": {
    "runtime": "python3.9",
    "handler": "index.handler",
    "memory_size": 256,
    "timeout": 60,
    "function_name": "data-processor",
    "environment_variables": {
      "STAGE": "prod",
      "LOG_LEVEL": "info"
    },
    "tags": {
      "Name": "data-processor",
      "Environment": "production"
    }
  }
}
```

---

### 2.6 S3 bucket — `type: "s3_bucket"`

Represents an AWS S3 bucket. No edges required.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `bucket` | string | **Yes*** | Globally unique bucket name. *Can be omitted if `label` is set (used as bucket name). |
| `versioning` | boolean | No | Enable versioning. |
| `block_public_acls` | boolean | No | Block public ACLs. |
| `block_public_policy` | boolean | No | Block public bucket policy. |
| `force_destroy` | boolean | No | Allow non-empty bucket destroy. |
| `tags` | object | No | String key-value pairs. |

**Edges:** None required.

**Sample node:**

```json
{
  "id": "s3-app-data",
  "type": "s3_bucket",
  "label": "Application Data Bucket",
  "position": { "x": 400, "y": 600 },
  "properties": {
    "bucket": "my-app-data-bucket-unique-12345",
    "versioning": true,
    "block_public_acls": true,
    "block_public_policy": true,
    "force_destroy": false,
    "tags": {
      "Name": "app-data",
      "Purpose": "uploads-and-assets",
      "Environment": "production"
    }
  }
}
```

---

### 2.7 RDS instance — `type: "rds_instance"`

Represents an AWS RDS DB instance. Typically attached to security group(s) via **`connects_to`**.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `engine` | string | **Yes** | e.g. `"postgres"`, `"mysql"`. |
| `instance_class` | string | **Yes** | e.g. `"db.t3.micro"`. |
| `allocated_storage` | number | **Yes** | Storage in GB. |
| `engine_version` | string | No | e.g. `"15.4"`. |
| `storage_type` | string | No | e.g. `"gp3"`, `"gp2"`. |
| `db_name` | string | No | Initial DB name. |
| `username` | string | No | Master username. |
| `password` | string | No | Master password. If omitted, not written to Terraform (use variables in practice). |
| `skip_final_snapshot` | boolean | No | Set true for dev/test to allow destroy. |
| `backup_retention_period` | number | No | Days (e.g. 7). |
| `multi_az` | boolean | No | Multi-AZ deployment. |
| `tags` | object | No | String key-value pairs. |

**Edges:** **`connects_to`** from **security_group** node(s) (SG → RDS): parser sets `vpc_security_group_ids`. Optional: **`contains`** from a **db_subnet_group** node if that type is added in the future.

**Sample node:**

```json
{
  "id": "rds-main",
  "type": "rds_instance",
  "label": "PostgreSQL Database",
  "position": { "x": 400, "y": 440 },
  "properties": {
    "engine": "postgres",
    "engine_version": "15.4",
    "instance_class": "db.t3.micro",
    "allocated_storage": 20,
    "storage_type": "gp3",
    "db_name": "appdb",
    "username": "dbadmin",
    "password": "CHANGE_ME_USE_VARIABLE",
    "skip_final_snapshot": true,
    "backup_retention_period": 7,
    "multi_az": false,
    "tags": {
      "Name": "main-postgres",
      "Role": "database",
      "Environment": "production"
    }
  }
}
```

**Sample edge:**

```json
{ "id": "e10", "source": "sg-db", "target": "rds-main", "type": "connects_to" }
```

---

## 3. Edge summary by resource

| Source node type   | Edge type     | Target node type   | Effect in Terraform |
|--------------------|---------------|--------------------|----------------------|
| **vpc**            | contains      | subnet             | Subnet’s `vpc_id` = VPC |
| **vpc**            | contains      | security_group     | SG’s `vpc_id` = VPC |
| **subnet**         | contains      | ec2_instance       | Instance’s `subnet_id` = Subnet |
| **security_group** | connects_to   | ec2_instance       | Instance’s `vpc_security_group_ids` includes SG |
| **security_group** | connects_to   | rds_instance       | RDS’s `vpc_security_group_ids` includes SG |

---

## 4. Full diagram example

Minimal valid diagram with one VPC, one subnet, one security group, and one EC2 instance:

```json
{
  "metadata": {
    "version": "1.0",
    "name": "simple-web",
    "description": "VPC with one web server",
    "environment": "production"
  },
  "nodes": [
    {
      "id": "vpc-1",
      "type": "vpc",
      "label": "Main VPC",
      "position": { "x": 200, "y": 100 },
      "properties": {
        "cidr_block": "10.0.0.0/16",
        "enable_dns_hostnames": true,
        "enable_dns_support": true
      }
    },
    {
      "id": "subnet-1",
      "type": "subnet",
      "label": "Public Subnet",
      "position": { "x": 200, "y": 220 },
      "properties": {
        "cidr_block": "10.0.1.0/24",
        "availability_zone": "us-east-1a",
        "map_public_ip_on_launch": true
      }
    },
    {
      "id": "sg-1",
      "type": "security_group",
      "label": "Web SG",
      "position": { "x": 400, "y": 220 },
      "properties": {
        "name": "web-sg",
        "description": "HTTP and SSH",
        "ingress": [
          { "from_port": 80, "to_port": 80, "protocol": "tcp", "cidr_blocks": ["0.0.0.0/0"] },
          { "from_port": 22, "to_port": 22, "protocol": "tcp", "cidr_blocks": ["10.0.0.0/16"] }
        ],
        "egress": [
          { "from_port": 0, "to_port": 0, "protocol": "-1", "cidr_blocks": ["0.0.0.0/0"] }
        ]
      }
    },
    {
      "id": "web-1",
      "type": "ec2_instance",
      "label": "Web Server",
      "position": { "x": 200, "y": 340 },
      "properties": {
        "ami": "ami-0c55b159cbfafe1f0",
        "instance_type": "t3.micro",
        "key_name": "my-key"
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "vpc-1", "target": "subnet-1", "type": "contains" },
    { "id": "e2", "source": "vpc-1", "target": "sg-1", "type": "contains" },
    { "id": "e3", "source": "subnet-1", "target": "web-1", "type": "contains" },
    { "id": "e4", "source": "sg-1", "target": "web-1", "type": "connects_to" }
  ]
}
```

---

## 5. Frontend implementation notes

1. **IDs:** Generate unique, stable `id`s (e.g. UUID or `type` + short id). Use them exactly in `edges.source` and `edges.target`.
2. **Position:** Persist `position` for drag-and-drop layout; the parser ignores it but the frontend can use it for rendering.
3. **Labels:** Provide a default `label` (e.g. from component type + id) so generated Terraform has readable names/tags.
4. **Validation:** The parser validates required fields and edge references. Emit the exact `type` strings (e.g. `"ec2_instance"`, `"security_group"`) and edge types (`"contains"`, `"connects_to"`, `"depends_on"`) as in this document.
5. **Optional properties:** Omit optional keys rather than sending `null` when the user has not set them; the parser uses documented defaults.

This component library reflects the **current** parser behaviour. New node types or properties may be added in future; the parser will ignore unknown node types and unknown properties.
