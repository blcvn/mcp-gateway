# Technical Design Document (TDD): MCP Platform

## 1. Internal Architecture

### 1.1 Hexagonal Design Detail
The system follows a Ports and Adapters pattern to ensure the core MCP logic remains decoupled from specific transport (HTTP/gRPC) or storage (PostgreSQL/Redis) implementations.

- **Internal Domain**: `internal/core/domain` contains the MCP protocol handlers, session management, and sandbox orchestration logic.
- **Port Interfaces**: `internal/core/ports` defines `ToolRegistry`, `ExecutionEngine`, `TenantStore`, and `AuditBuffer`.
- **Adapters**:
    - `adapters/http`: Implements the public MCP API and OAuth2 middleware.
    - `adapters/storage`: PostgreSQL implementation for `TenantStore`.
    - `adapters/sandbox`: WASM/Firecracker implementation for `ExecutionEngine`.

## 2. Data Models (Database Schema)

### 2.1 Tenants
```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 2.2 Tool Registry
```sql
CREATE TABLE tool_registry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    is_recommended BOOLEAN DEFAULT FALSE,
    schema JSONB NOT NULL,
    metadata JSONB, -- Stores category, description, tags (FR18)
    risk_level VARCHAR(20) DEFAULT 'unclassified',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, name, version)
);
```

### 2.3 Resource Usage (Metering)
```sql
CREATE TABLE resource_usage (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),
    month_year VARCHAR(7), -- e.g., '2026-02'
    invocation_count BIGINT DEFAULT 0,
    total_execution_time_ms BIGINT DEFAULT 0,
    hard_quota BIGINT,
    UNIQUE(tenant_id, month_year)
);
```

### 2.4 Audit Logs
```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),
    agent_id VARCHAR(255),
    tool_id UUID REFERENCES tool_registry(id),
    input_payload JSONB,
    output_payload JSONB,
    risk_classification JSONB,
    execution_time_ms INTEGER,
    status_code VARCHAR(50),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 2.5 API Keys
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL, -- SHA-256 hash of the API key
    scopes JSONB, -- list of permitted tools/actions
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE
);
```

### 2.6 Access Policies (RBAC/ABAC)
```sql
CREATE TABLE access_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    rule_type VARCHAR(20) NOT NULL, -- 'RBAC' or 'ABAC'
    rule_definition JSONB NOT NULL, -- OPA policy or standard permission set
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 2.7 DLP Configuration
```sql
CREATE TABLE dlp_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    patterns JSONB NOT NULL, -- list of regex patterns or dictionary matchers
    severity VARCHAR(20) DEFAULT 'medium',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 2.8 Tool Execution Sessions
```sql
CREATE TABLE tool_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    agent_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'running', 'completed', 'failed'
    context_data JSONB, -- Stores conversation history or stateful variables
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    finished_at TIMESTAMP WITH TIME ZONE
);
```

## 3. Component Interaction flows

### 3.1 Handshake & Transport (FR17, FR22, FR23, FR26)
1.  **Transport Initiation**: `adapters/http` upgrades connection (HTTP to SSE/WS) (FR17).
2.  **Negotiation**: `domain/initialize` processes the handshake, verifying protocol version (FR22).
3.  **Serialization Setup**: Gateway confirms JSON-RPC 2.0 as the wire format (FR23).

### 3.2 Security & Authorization (FR4, FR5, FR26)
1.  **Identity**: Middleware intercepts request, terminates TLS 1.3, and verifies OIDC/mTLS (FR4, FR26).
2.  **Context**: `adapters/http` builds `SecurityContext` containing user/tenant claims.
3.  **Policy**: `domain/policy` delegates to OPA/Go-validator to enforce RBAC rules (FR5).

### 3.3 Tool Discovery & Management (FR1, FR18, FR19, FR24)
1.  **Polling**: `adapters/storage` or file-watcher detects changes (FR1).
2.  **Metadata Loading**: Registry loads schemas and enhanced metadata (category, tags) (FR18).
3.  **Listing**: `domain/discover` filters tools based on context (FR19) and applies cursor-based pagination (FR24).

### 3.4 Invocation & Isolation (FR6, FR7, FR8)
1.  **Validation**: `domain/validator` executes schema check (FR6).
2.  **Isolation**: `domain/executor` injects `tenant_id` into the runtime environment (FR7).
3.  **Environment Injector**: Vault adapter populates environment variables with tool secrets (FR8).

### 3.5 Privacy & Observability (FR9-15, FR25)
1.  **DLP Redaction**: Stream output passes through `domain/redactor` for PII filtering (FR12, FR13).
2.  **Metrics Export**: `adapters/observability` increments Prometheus counters asynchronously (FR10, FR15).
3.  **Tracing**: Trace headers are propagated to all downstream tool sub-processes (FR11).

## 4. Security Implementation Details

### 4.1 Zero-Trust Middleware
- **Context Injection**: Each request context is enriched with a `SecurityContext` containing identity and permissions.
- **ABAC Engine**: Policy evaluations use the Open Policy Agent (OPA) or a custom Go-based logic to check requirements like "Tool X requires Senior Admin role for Tenant Y".

### 4.2 Sandbox Runtime
- **WASM Implementation**: Use `wasmer-go` for low-latency, stateless function execution. Memory (512MB) and CPU quotas are enforced per invocation.
- **DLP Redaction Pipeline**: A stream processor in the Data Plane scans outputs for PII patterns. It uses a high-performance regex engine for real-time masking before data transfer.

## 5. Deployment Schema
- **Stateless MCP Nodes**: Horizontally scalable via K8s HPA based on CPU/Memory/Request throughput.
- **Sidecars**: Envoy or Istio for mTLS and advanced traffic management.
- **Secrets Management**: Integration with HashiCorp Vault using Go's `vault/api` for dynamic tool secret injection.
