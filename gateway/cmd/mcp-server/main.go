package main

import (
	"log"
	"net/http"

	"vnp-network-backend/mcp/internal/adapters/http_adapter"
	"vnp-network-backend/mcp/internal/config"
	"vnp-network-backend/mcp/internal/core/tools"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Core Components
	toolRegistry := tools.NewRegistry()

	// 3. Initialize Adapters
	// TODO: Add Auth and RateLimit middleware
	handler := http_adapter.NewHandler(toolRegistry)

	// 4. Start Server
	log.Printf("Starting MCP Server on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler.Router()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
