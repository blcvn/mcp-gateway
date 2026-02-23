module github.com/blcvn/mcp-gateway/mcp-server

go 1.24.0

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/mark3labs/mcp-go v0.1.0 // Hypothetical Go MCP SDK, or implement minimal protocol manually
	github.com/spf13/cobra v1.8.0
)

replace github.com/blcvn/ba-shared-libs/pkg => ../../ba-shared-libs/pkg

replace github.com/blcvn/ba-shared-libs/proto => ../../ba-shared-libs/proto
