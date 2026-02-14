package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/blcvn/backend/services/ba-mcp-server/pkg/mcp"
)

func main() {
	server := mcp.NewServer("confluence-mcp-go", "1.0.0")

	// Define tool: confluence_search
	searchTool := mcp.Tool{
		Name:        "confluence_search",
		Description: "Search Confluence pages by CQL (Confluence Query Language)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The CQL query string",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Max results to return",
				},
			},
			"required": []string{"query"},
		},
	}

	searchHandler := func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		query, ok := args["query"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'query' argument")
		}

		// Mock Implementation
		log.Printf("Searching Confluence with query: %s", query)

		results := []map[string]string{
			{"id": "101", "title": "PRD: User Login", "url": "/wiki/101"},
			{"id": "102", "title": "Meeting Notes: 2024-02-14", "url": "/wiki/102"},
		}

		responsePayload := map[string]interface{}{
			"results": results,
			"query":   query,
		}

		jsonBytes, _ := json.Marshal(responsePayload)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: string(jsonBytes),
				},
			},
		}, nil
	}

	server.RegisterTool(searchTool, searchHandler)

	// Start Server
	log.Println("Confluence MCP Server (Go) running on Stdio")
	server.ServeStdio()
}
