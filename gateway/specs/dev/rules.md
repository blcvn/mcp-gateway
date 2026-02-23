# Developer Rules & Standards

## 📦 Required Go Packages

### MCP & Transport
- `github.com/modelcontextprotocol/go-sdk`: Core
- `github.com/sourcegraph/jsonrpc2`: Core JSON-RPC implementation.
- `github.com/gorilla/websocket`: WebSocket transport for MCP.

### Configuration & Tooling
- `github.com/spf13/viper`: Configuration management.
- `github.com/fsnotify/fsnotify`: File system watching for tool hot reloading.

### HTTP & Networking
- `github.com/go-resty/resty/v2`: Elegant HTTP client.
- `github.com/hashicorp/go-retryablehttp`: Retries and backoff for downstream APIs.

### Security & Reliability
- `github.com/sony/gobreaker`: Circuit breaker implementation.
- `golang.org/x/time/rate`: Token bucket rate limiting.

### Logging & Observability
- `go.uber.org/zap`: High-performance structured logging.
- `github.com/prometheus/client_golang`: Prometheus metrics.
- `go.opentelemetry.io/otel`: OpenTelemetry integration.

### Validation
- `github.com/go-playground/validator/v10`: Struct validation.

## 🧪 Testing Strategy

### Unit Tests
- **Goal**: 80%+ coverage for business logic in `internal/core`.
- **Approach**: Mock all external ports (DB, Caching, External APIs) using `uber-go/mock`.

### Integration Tests
- **Goal**: Verify interaction between components (e.g., API Layer to Data Plane).
- **Approach**: 
    - Use `testcontainers-go` for Redis and PostgreSQL.
    - Spin up mock HTTP servers (`httptest`) for downstream tool calls.

### Contract & Schema Tests
- **Goal**: Ensure tool definitions match MCP/OpenAPI specifications.
- **Approach**: Validate JSON/YAML schemas against the registry's expectations.

### Security Testing
- **Goal**: Verify sandbox isolation and auth enforcement.
- **Approach**: Automated scripts to attempt unprivileged tool access and prompt injection scenarios.


