# MCP Management System — Frequently Asked Questions

**Project Code:** MCP-MGT  
**Version:** 1.0  
**Date:** 2026-02-18  

---

## General Questions

### Q1: What is the MCP Management System?
**A:** It is a centralized platform that provisions, orchestrates, monitors, and governs **multiple Model Context Protocol (MCP) servers**. Each MCP server is dedicated to a **service group** (DevOps, Collaboration, ITSM, Google Workspace, Custom) and securely exposes enterprise tools to AI agents and human users through a standardized protocol.

### Q2: Why do we need multiple MCP servers instead of one?
**A:** Separation by service group provides:
- **Security isolation** — a breach in one group cannot access tools in another
- **Independent scaling** — DevOps tools may need 10× capacity during deployments
- **Blast radius containment** — an outage in Google Workspace connectors won't affect Jira
- **Policy granularity** — different risk levels and compliance requirements per group
- **Team ownership** — each team can manage their own group's connectors

### Q3: What services are supported?
**A:** Five service groups at launch:

| Group | Services |
|-------|----------|
| **A — DevOps** | Ansible Tower, Terraform Cloud, Jenkins, ArgoCD, GitLab/GitHub |
| **B — Collaboration** | Jira, Confluence, Trello, Slack, Microsoft Teams |
| **C — ITSM & CMDB** | ServiceNow, CMDB, PagerDuty, Opsgenie |
| **D — Google Workspace** | Gmail, Drive, Chat, Calendar, Sheets |
| **E — Custom** | Internal APIs, database tools, reporting |

### Q4: Can we add new service groups after launch?
**A:** Yes. The system is designed with an extensible **connector framework**. Adding a new group requires:
1. Create a service group configuration YAML
2. Implement connector(s) using the base interface
3. Define OPA policies for the new group
4. Deploy a new MCP server instance via the K8s Operator

Typical effort: 2–4 weeks per connector, depending on downstream API complexity.

### Q5: How does this relate to the existing MCP server code?
**A:** The existing `mcp/` codebase becomes the foundation for individual MCP server instances. The new Management System adds:
- **Control Plane** — lifecycle, registry, policy management
- **Connector Framework** — standardized adapters for each service
- **Multi-Server Orchestration** — K8s Operator for provisioning
- **Management Dashboard** — admin UI for operations
- **Security Layer** — DLP, sandbox, audit trail

---

## Architecture & Technical

### Q6: What is the Control Plane?
**A:** The Control Plane is the "brain" of the system. It manages:
- Server lifecycle (provision, health check, restart, scale, decommission)
- Tool & connector registry (what tools exist, where they run)
- Policy engine (OPA — who can invoke what, under what conditions)
- Identity & access management (Keycloak integration)
- Configuration distribution (etcd-backed config store)
- Audit & metering (compliance logging, cost tracking)

### Q7: How do AI agents authenticate?
**A:** Three methods, in order of recommendation:
1. **OAuth 2.0 Client Credentials** — Agent gets a service account in Keycloak, authenticates via client_id/client_secret, receives a JWT with scoped permissions
2. **mTLS** — Agent presents a client certificate signed by our internal CA (managed by Vault)
3. **API Key** — Scoped key stored in Vault, rotated every 90 days

### Q8: How are tools discovered by agents?
**A:** Two discovery mechanisms:
- **MCP Protocol Discovery** — Standard `tools/list` MCP method returns available tools with JSON Schema parameters
- **Management API** — `GET /api/v1/servers/{id}/tools` returns the full catalog with metadata, risk levels, and documentation

Agents see only tools they are authorized to access (filtered by OPA policy evaluation).

### Q9: What happens if a downstream service (e.g., Jira) goes down?
**A:** Multi-layer resilience:
1. **Circuit Breaker** — After 5 consecutive failures, the connector trips and returns a degraded response
2. **Retry with Backoff** — Exponential backoff (1s, 2s, 4s, max 30s)
3. **Health Check** — Server marks the connector as unhealthy
4. **Alert** — PagerDuty notification to on-call team
5. **Dashboard** — Real-time status visible in management UI
6. **Graceful Degradation** — Other tools in the same group remain operational

### Q10: What is the expected latency for a tool invocation?
**A:** End-to-end (agent → response):

| Component | Budget |
|-----------|--------|
| API Gateway + Auth | < 20ms |
| Policy Evaluation (OPA) | < 5ms |
| MCP Server Processing | < 10ms |
| Connector + Downstream | 100ms–2s (varies) |
| DLP Scanning | < 20ms |
| **Total (p95)** | **< 500ms** (excluding downstream) |

---

## Security & Compliance

### Q11: How is data classified in the system?
**A:** Four levels:
- **PUBLIC** — Marketing materials, public documentation
- **INTERNAL** — Business documents, project plans
- **CONFIDENTIAL** — Customer data, financial records, PII
- **RESTRICTED** — Trade secrets, M&A data, encryption keys

Each tool invocation is tagged with the expected data classification, and DLP policies enforce appropriate handling.

### Q12: What prevents an AI agent from accessing unauthorized data?
**A:** Defense in depth:
1. **Identity** — Every agent has a unique service account with explicit permissions
2. **RBAC** — Agents are scoped to specific service groups and tools
3. **ABAC** — Risk-level gating (CRITICAL tools need human approval)
4. **DLP** — PII and sensitive data are redacted from tool outputs
5. **Rate Limiting** — Prevents runaway agent loops
6. **Audit** — Every invocation is logged for forensic analysis
7. **WASM Sandbox** — Custom tools execute in isolated environments

### Q13: How do we meet SOC 2 Type II requirements?
**A:** The system generates compliance evidence automatically:
- **CC6 (Access Control)** — Keycloak + OPA policy decisions logged
- **CC7 (System Operations)** — Prometheus metrics, health checks, incident records
- **CC8 (Change Management)** — All config changes audited with before/after diffs
- **CC9 (Risk Assessment)** — Tools tagged with risk levels, policy enforcement metrics
- Audit logs retained for 7 years in TimescaleDB → S3 Glacier

### Q14: What about GDPR and data privacy?
**A:** GDPR compliance is built-in:
- **Right to Erasure** — API to purge user data from all logs and caches
- **Data Minimization** — Tool invocations log metadata, not full payloads (configurable)
- **Consent** — Google Workspace tools enforce OAuth consent scopes
- **DLP** — Automatic PII detection and redaction in tool outputs
- **Data Processing Records** — Automated Article 30 record generation

### Q15: Can the system detect prompt injection attacks?
**A:** Yes, the Prompt Injection Firewall:
1. **Pattern Matching** — Known injection patterns (instruction overrides, jailbreaks)
2. **Input Sanitization** — Escape special characters in tool parameters
3. **Output Validation** — Verify tool outputs match expected schema
4. **Anomaly Detection** — Flag unusual parameter patterns (future: ML-based)
5. **Blocking** — CRITICAL risk tools reject suspicious inputs automatically

---

## Operations & Deployment

### Q16: How many environments do we need?
**A:** Three environments:

| Environment | Purpose | Scale |
|-------------|---------|-------|
| **Development** | Local Docker Compose | Single instance |
| **Staging** | Pre-production testing | 1 replica per server |
| **Production** | Live workload | 2–10 replicas per server (HPA) |

### Q17: How do we deploy updates to MCP servers?
**A:** GitOps via ArgoCD:
1. Code merged to `main` → CI builds Docker image
2. Helm values updated in deploy repo
3. ArgoCD detects drift → triggers sync
4. Rolling update (zero-downtime, one pod at a time)
5. Automated smoke tests post-deploy
6. Automatic rollback if health checks fail

### Q18: What monitoring is available?
**A:** Full observability stack:
- **Grafana Dashboards** — Server health, tool metrics, error rates, latency
- **Prometheus Alerts** — Server down, error rate > 5%, latency > 2s, Vault seal
- **Jaeger Traces** — End-to-end request tracing across all services
- **ELK Logs** — Searchable structured logs from all components
- **Audit Dashboard** — Tool invocation history with filters and export

### Q19: What is the expected cost?
**A:** Monthly infrastructure: ~$32,000. Annual total (with 12 FTE team): ~$2.2M in Year 1, ~$900K/year ongoing.

### Q20: What is the team structure?
**A:** 12 FTE: 1 Tech Lead, 4 Backend Engineers (Go), 2 DevOps, 1 Security Engineer, 1 Frontend Engineer, 1 QA, 1 PO, 1 SRE.

---

## Connector Development

### Q21: How long does it take to build a new connector?
**A:** Typical timelines:

| Complexity | Duration | Examples |
|------------|----------|---------|
| Simple (REST, API key) | 1 week | Trello, PagerDuty |
| Medium (OAuth2, pagination) | 2 weeks | Jira, Confluence, Slack |
| Complex (multi-auth, streaming) | 3–4 weeks | Google Workspace, ServiceNow |
| Critical (high security) | 4–6 weeks | Ansible Tower, Terraform |

### Q22: Can connectors be developed by external teams?
**A:** Yes, via the **Plugin SDK** (Phase 4):
- Connector compiled as a Go plugin or WASM module
- Must implement the base `Connector` interface
- Runs in WASM sandbox (no direct system access)
- Code reviewed by security team before deployment
- Published to internal connector registry

### Q23: How are connector secrets managed?
**A:** All secrets stored in **HashiCorp Vault**:
- Static secrets (API keys): `vault:secret/data/connectors/{name}`
- Dynamic secrets (DB credentials): Auto-generated, auto-expired
- OAuth tokens: Stored encrypted, auto-refreshed
- Rotation: Automated 30–90 day cycles depending on sensitivity
- Access: Each MCP server can only read its own group's secrets

---

## Integration & Usage

### Q24: How do AI agents interact with the system?
**A:** Standard MCP protocol flow:
```
1. Agent → MCP Server: tools/list (discover available tools)
2. Agent selects tool based on user intent
3. Agent → MCP Server: tools/call (invoke tool with parameters)
4. MCP Server → OPA: Authorize (check policy)
5. MCP Server → Connector: Execute (call downstream API)
6. MCP Server → DLP: Scan output (redact PII if needed)
7. MCP Server → Agent: Return result
8. All steps → Audit Log: Record event
```

### Q25: Can tools call other tools (cross-group workflows)?
**A:** Yes, via **event choreography** (NATS JetStream):
- Example: "Create Jira ticket" → event → "Update CMDB" → event → "Notify Slack"
- Each step is independently authorized and audited
- Workflow definitions stored in configuration (YAML)
- Future: Visual workflow builder in dashboard

### Q26: What is the SLA?
**A:** Production SLA targets:

| Metric | Target |
|--------|--------|
| System Uptime | 99.99% (52 min downtime/year) |
| Tool Invocation Success Rate | > 99.5% |
| P95 Latency (platform overhead) | < 50ms |
| RTO (Recovery Time Objective) | < 5 minutes |
| RPO (Recovery Point Objective) | < 1 minute |
| Incident Response (P0) | 15 minutes |

---

## Decision Making

### Q27: Why Go instead of Python/Java/Node.js?
**A:** Go is optimal for this platform because:
- **Performance** — Low latency, high throughput (10,000+ RPS per node)
- **Concurrency** — Goroutines for handling many simultaneous tool invocations
- **Existing Codebase** — The MCP server is already in Go
- **K8s Native** — K8s operators are idiomatically written in Go
- **Binary Deployment** — Single static binary, no runtime dependencies
- **Type Safety** — Compile-time error catching for enterprise reliability

### Q28: Why Keycloak over Okta/Auth0?
**A:** For a fintech with strong engineering:
- **Cost** — $0 license vs $120K+/year for Okta Enterprise
- **Control** — Self-hosted, full customization, audit every line of code
- **Compliance** — On-premises data residency (important for fintech regulations)
- **Federation** — Full LDAP/AD, SAML, OIDC support
- **Extensibility** — Custom authenticators, session policies

### Q29: Why separate MCP servers per group vs. namespaces in one server?
**A:** Process-level isolation is stronger than namespace isolation:
- **Security** — Compromise of one server doesn't affect others
- **Scaling** — Each server scales independently based on demand
- **Fault Isolation** — OOM kill or crash affects only one group
- **Deployment** — Update one group without touching others
- **Resource Control** — K8s resource quotas per server

### Q30: What is the go/no-go decision criteria?
**A:** The following must be met before production launch:
- [ ] All 5 service groups operational with ≥ 3 tools each
- [ ] Zero critical findings in penetration test
- [ ] SOC 2 Type I evidence package complete
- [ ] 99.99% uptime demonstrated in staging (30 days)
- [ ] Load test: 10,000 concurrent invocations with < 500ms p95
- [ ] DR failover test passed (RTO < 5 min)
- [ ] Security review signed off by CISO
- [ ] Runbooks and on-call rotation established

---

**Document Version:** 1.0  
**Last Updated:** 2026-02-18  
**Author:** Enterprise Architecture Team
