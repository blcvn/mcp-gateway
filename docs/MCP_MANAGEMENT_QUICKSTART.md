# MCP Management System — Developer Quickstart Guide

**Project Code:** MCP-MGT  
**Version:** 1.0  
**Date:** 2026-02-18  

---

## 1. Repository Structure

```
mcp-management/
├── cmd/
│   ├── control-plane/          # Control plane entry point
│   │   └── main.go
│   └── mcp-server/             # MCP server entry point (per-group)
│       └── main.go
├── internal/
│   ├── controlplane/           # Control plane services
│   │   ├── lifecycle/          # Server lifecycle management
│   │   ├── registry/           # Tool & server registry
│   │   ├── policy/             # OPA policy integration
│   │   ├── identity/           # Keycloak integration
│   │   ├── audit/              # Audit event pipeline
│   │   └── metering/           # Cost & usage metering
│   ├── mcpserver/              # MCP server runtime
│   │   ├── engine/             # Tool execution engine
│   │   ├── sandbox/            # WASM sandbox runtime
│   │   ├── dlp/                # DLP pipeline
│   │   └── streaming/          # SSE/WebSocket support
│   ├── connectors/             # Service connectors (one per integration)
│   │   ├── base/               # Base connector interface + utilities
│   │   ├── jira/               # Jira Cloud connector
│   │   ├── confluence/         # Confluence connector
│   │   ├── trello/             # Trello connector
│   │   ├── ansible_tower/      # Ansible Tower/AWX connector
│   │   ├── terraform/          # Terraform Cloud connector
│   │   ├── google_workspace/   # Google Workspace (Mail, Drive, Chat, etc.)
│   │   ├── servicenow/         # ServiceNow connector
│   │   ├── cmdb/               # CMDB connector
│   │   ├── pagerduty/          # PagerDuty connector
│   │   └── slack/              # Slack connector
│   ├── core/                   # Shared domain models & interfaces
│   │   ├── models/             # Domain entities
│   │   ├── ports/              # Port interfaces (hexagonal architecture)
│   │   └── events/             # Domain events
│   ├── adapters/               # Infrastructure adapters
│   │   ├── postgres/           # PostgreSQL adapter
│   │   ├── redis/              # Redis adapter
│   │   ├── vault/              # Vault adapter
│   │   ├── etcd/               # etcd adapter
│   │   └── nats/               # NATS JetStream adapter
│   └── config/                 # Configuration management
├── api/
│   ├── openapi/                # OpenAPI 3.0 specs
│   │   ├── management.yaml     # Management API spec
│   │   └── mcp-protocol.yaml   # MCP protocol spec
│   └── graphql/                # GraphQL schema (for dashboard)
│       └── schema.graphql
├── dashboard/                  # Next.js management dashboard
│   ├── app/
│   ├── components/
│   └── lib/
├── deploy/
│   ├── k8s/                    # Kubernetes manifests
│   │   ├── base/               # Base manifests (Kustomize)
│   │   ├── overlays/           # Per-environment overlays
│   │   │   ├── dev/
│   │   │   ├── staging/
│   │   │   └── production/
│   │   └── operator/           # K8s operator CRDs
│   ├── helm/                   # Helm charts
│   │   ├── mcp-control-plane/
│   │   ├── mcp-server/
│   │   └── mcp-observability/
│   └── terraform/              # Infrastructure as Code
│       ├── modules/
│       └── environments/
├── configs/                    # Configuration files
│   ├── groups/                 # Service group definitions
│   │   ├── group_a_devops.yaml
│   │   ├── group_b_collaboration.yaml
│   │   ├── group_c_itsm.yaml
│   │   ├── group_d_google.yaml
│   │   └── group_e_custom.yaml
│   ├── policies/               # OPA policy files (Rego)
│   │   ├── rbac.rego
│   │   ├── abac.rego
│   │   └── dlp.rego
│   └── connectors/             # Connector configurations
├── plugins/                    # Custom tool plugins
├── scripts/                    # Build & operations scripts
│   ├── setup.sh                # Local dev environment setup
│   ├── seed.sh                 # Seed test data
│   └── migrate.sh              # Database migrations
├── tests/
│   ├── integration/            # Integration tests
│   ├── e2e/                    # End-to-end tests
│   └── load/                   # Load tests (k6)
├── docs/                       # Additional documentation
├── docker-compose.yaml         # Local development stack
├── Makefile                    # Build automation
└── go.mod
```

---

## 2. Local Development Setup

### Prerequisites

```bash
# Required tools
go install golang.org/dl/go1.22.0@latest  # Go 1.22+
brew install docker docker-compose         # Docker Desktop
brew install kubectl helm                  # Kubernetes tools
brew install vault                         # Vault CLI
brew install opa                           # OPA CLI (policy testing)
brew install node@20                       # Node.js 20+ (for dashboard)
brew install k6                            # Load testing
```

### Step 1: Start Infrastructure (Docker Compose)

```yaml
# docker-compose.yaml (key services)
services:
  postgres:
    image: postgres:16-alpine
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: mcp_management
      POSTGRES_USER: mcp
      POSTGRES_PASSWORD: dev-password

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  keycloak:
    image: quay.io/keycloak/keycloak:24.0
    ports: ["8080:8080"]
    environment:
      KEYCLOAK_ADMIN: admin
      KEYCLOAK_ADMIN_PASSWORD: admin
    command: start-dev

  vault:
    image: hashicorp/vault:1.15
    ports: ["8200:8200"]
    environment:
      VAULT_DEV_ROOT_TOKEN_ID: dev-root-token

  etcd:
    image: bitnami/etcd:3.5
    ports: ["2379:2379"]
    environment:
      ALLOW_NONE_AUTHENTICATION: "yes"

  opa:
    image: openpolicyagent/opa:0.62.0
    ports: ["8181:8181"]
    command: run --server --log-level debug

  nats:
    image: nats:2.10-alpine
    ports: ["4222:4222", "8222:8222"]
    command: --jetstream

  timescaledb:
    image: timescale/timescaledb:latest-pg16
    ports: ["5433:5432"]
    environment:
      POSTGRES_DB: mcp_audit
      POSTGRES_PASSWORD: dev-password

  jaeger:
    image: jaegertracing/all-in-one:1.54
    ports: ["16686:16686", "4317:4317"]

  prometheus:
    image: prom/prometheus:v2.49.0
    ports: ["9090:9090"]

  grafana:
    image: grafana/grafana:10.3.0
    ports: ["3001:3000"]
```

```bash
# Start all infrastructure
docker compose up -d

# Verify all services are running
docker compose ps
```

### Step 2: Initialize Database

```bash
# Run migrations
make migrate-up

# Seed development data (service groups, sample tools, test users)
make seed
```

### Step 3: Run Control Plane

```bash
# Start the control plane service
make run-control-plane

# API available at http://localhost:8000
# Swagger UI at http://localhost:8000/swagger
```

### Step 4: Run MCP Server (Group B — Collaboration)

```bash
# Start a single MCP server for development
MCP_GROUP=group_b make run-mcp-server

# MCP server available at http://localhost:9000
# Tools: jira.create_issue, jira.search, confluence.create_page, etc.
```

### Step 5: Run Dashboard

```bash
cd dashboard
npm install
npm run dev

# Dashboard at http://localhost:3000
```

---

## 3. Building a New Connector

### Step 1: Define the Connector Interface

```go
// internal/connectors/base/connector.go
package base

type Connector interface {
    // Name returns the service name (e.g., "jira")
    Name() string

    // Initialize sets up authentication and connection pooling
    Initialize(cfg ConnectorConfig) error

    // Tools returns the list of MCP tools this connector exposes
    Tools() []ToolDefinition

    // Execute runs a specific tool with given arguments
    Execute(ctx context.Context, toolName string, args map[string]any) (*ToolResult, error)

    // HealthCheck verifies the downstream service is reachable
    HealthCheck(ctx context.Context) error

    // Close gracefully shuts down connections
    Close() error
}

type ConnectorConfig struct {
    BaseURL      string            `yaml:"base_url"`
    AuthType     string            `yaml:"auth_type"`  // oauth2, api_key, basic, mtls
    Credentials  string            `yaml:"credentials"` // Vault path
    RateLimit    int               `yaml:"rate_limit"`  // requests per minute
    Timeout      time.Duration     `yaml:"timeout"`
    RetryPolicy  RetryConfig       `yaml:"retry"`
    CustomConfig map[string]string `yaml:"custom"`
}

type ToolDefinition struct {
    Name        string           `json:"name"`
    Description string           `json:"description"`
    RiskLevel   string           `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
    Parameters  json.RawMessage  `json:"parameters"` // JSON Schema
    Tags        []string         `json:"tags"`
}

type ToolResult struct {
    Data    any    `json:"data"`
    Status  string `json:"status"` // success, error, partial
    Message string `json:"message,omitempty"`
}
```

### Step 2: Implement the Connector (Example: Trello)

```go
// internal/connectors/trello/trello.go
package trello

import (
    "context"
    "mcp-management/internal/connectors/base"
)

type TrelloConnector struct {
    client  *http.Client
    baseURL string
    apiKey  string
    token   string
}

func New() *TrelloConnector {
    return &TrelloConnector{}
}

func (c *TrelloConnector) Name() string { return "trello" }

func (c *TrelloConnector) Initialize(cfg base.ConnectorConfig) error {
    // Load secrets from Vault
    // Set up HTTP client with retry + circuit breaker
    return nil
}

func (c *TrelloConnector) Tools() []base.ToolDefinition {
    return []base.ToolDefinition{
        {
            Name:        "trello.list_boards",
            Description: "List all Trello boards accessible to the authenticated user",
            RiskLevel:   "LOW",
            Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
        },
        {
            Name:        "trello.create_card",
            Description: "Create a new card on a Trello board",
            RiskLevel:   "LOW",
            Parameters: json.RawMessage(`{
                "type": "object",
                "properties": {
                    "board_id": {"type":"string","description":"Board ID"},
                    "list_id": {"type":"string","description":"List ID"},
                    "name": {"type":"string","description":"Card title"},
                    "description": {"type":"string","description":"Card description"}
                },
                "required": ["board_id", "list_id", "name"]
            }`),
        },
    }
}

func (c *TrelloConnector) Execute(ctx context.Context, toolName string, args map[string]any) (*base.ToolResult, error) {
    switch toolName {
    case "trello.list_boards":
        return c.listBoards(ctx)
    case "trello.create_card":
        return c.createCard(ctx, args)
    default:
        return nil, fmt.Errorf("unknown tool: %s", toolName)
    }
}
```

### Step 3: Register the Connector

```yaml
# configs/groups/group_b_collaboration.yaml
group:
  id: group_b
  name: "Collaboration & Project Management"
  description: "Jira, Confluence, Trello, Slack, Teams"

connectors:
  - name: trello
    type: trello
    config:
      base_url: "https://api.trello.com/1"
      auth_type: api_key
      credentials: "vault:secret/data/connectors/trello"
      rate_limit: 100
      timeout: 30s
      retry:
        max_attempts: 3
        backoff: exponential
```

---

## 4. Makefile Commands

```makefile
# Build
make build                  # Build all binaries
make build-control-plane    # Build control plane only
make build-mcp-server       # Build MCP server only
make build-dashboard        # Build Next.js dashboard

# Run
make run-control-plane      # Run control plane locally
make run-mcp-server         # Run MCP server (set MCP_GROUP env)
make run-dashboard          # Run dashboard dev server

# Database
make migrate-up             # Apply database migrations
make migrate-down           # Rollback last migration
make seed                   # Seed development data

# Test
make test                   # Run all unit tests
make test-integration       # Run integration tests (needs docker compose)
make test-e2e               # Run end-to-end tests
make test-load              # Run k6 load tests
make test-coverage          # Generate coverage report

# Policy
make policy-test            # Run OPA policy unit tests
make policy-lint            # Lint Rego files

# Docker
make docker-build           # Build all Docker images
make docker-push            # Push to registry
make compose-up             # Start local infrastructure
make compose-down           # Stop local infrastructure

# Quality
make lint                   # Run golangci-lint
make fmt                    # Format Go code
make vet                    # Go vet
make security-scan          # Run Trivy on Docker images
```

---

## 5. Environment Variables

```bash
# Control Plane
MCP_CONTROL_PLANE_PORT=8000
MCP_DB_HOST=localhost
MCP_DB_PORT=5432
MCP_DB_NAME=mcp_management
MCP_DB_USER=mcp
MCP_DB_PASSWORD=dev-password
MCP_REDIS_URL=redis://localhost:6379
MCP_VAULT_ADDR=http://localhost:8200
MCP_VAULT_TOKEN=dev-root-token
MCP_KEYCLOAK_URL=http://localhost:8080
MCP_KEYCLOAK_REALM=mcp
MCP_OPA_URL=http://localhost:8181
MCP_ETCD_ENDPOINTS=http://localhost:2379
MCP_NATS_URL=nats://localhost:4222
MCP_OTEL_ENDPOINT=http://localhost:4317
MCP_AUDIT_DB_HOST=localhost
MCP_AUDIT_DB_PORT=5433

# MCP Server
MCP_SERVER_PORT=9000
MCP_GROUP=group_b
MCP_CONTROL_PLANE_URL=http://localhost:8000
```

---

## 6. Testing Strategy

```yaml
Unit Tests:
  Coverage Target: ≥ 80%
  Framework: Go testing + testify
  Focus: Business logic, policy evaluation, connector parsing

Integration Tests:
  Framework: Go testing + testcontainers
  Focus: Database operations, Vault integration, Keycloak flows
  Requires: Docker

E2E Tests:
  Framework: Go testing
  Focus: Full request flow (Agent → API → MCP Server → Connector)
  Requires: docker compose up

Load Tests:
  Tool: k6
  Scenarios:
    - Steady state: 100 RPS for 10 minutes
    - Spike: 0 → 1000 RPS in 30 seconds
    - Soak: 200 RPS for 1 hour
  Targets:
    - p95 latency < 500ms
    - Error rate < 0.1%
    - Zero data loss

Policy Tests:
  Framework: OPA test (opa test ./configs/policies/)
  Coverage: All RBAC/ABAC decision paths
```

---

## 7. Access Local Services

| Service | URL | Credentials |
|---------|-----|-------------|
| Management API | http://localhost:8000 | Keycloak token |
| MCP Server | http://localhost:9000 | API key from Vault |
| Dashboard | http://localhost:3000 | Keycloak SSO |
| Keycloak Admin | http://localhost:8080 | admin / admin |
| Vault UI | http://localhost:8200 | Token: dev-root-token |
| Grafana | http://localhost:3001 | admin / admin |
| Jaeger UI | http://localhost:16686 | — |
| Prometheus | http://localhost:9090 | — |
| NATS Monitoring | http://localhost:8222 | — |

---

**Document Version:** 1.0  
**Last Updated:** 2026-02-18  
**Author:** Engineering Team
