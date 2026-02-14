package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/blcvn/backend/services/ba-mcp-server/pkg/mcp"
)

func main() {
	server := mcp.NewServer("figma-mcp-go", "1.0.0")

	// Define tool: figma_get_file
	fileTool := mcp.Tool{
		Name:        "figma_get_file",
		Description: "Get Figma file nodes and structure",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_key": map[string]interface{}{
					"type":        "string",
					"description": "The Figma file key",
				},
				"node_ids": map[string]interface{}{
					"type":        "string",
					"description": "Comma separated node IDs",
				},
			},
			"required": []string{"file_key"},
		},
	}

	fileHandler := func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		fileKey, ok := args["file_key"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'file_key' argument")
		}

		// Mock Implementation
		log.Printf("Fetching Figma file: %s", fileKey)

		nodes := map[string]interface{}{
			"document": map[string]string{"id": "0:0", "name": "Document"},
			"page1":    map[string]string{"id": "0:1", "name": "Page 1"},
		}

		responsePayload := map[string]interface{}{
			"file_key": fileKey,
			"nodes":    nodes,
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

	server.RegisterTool(fileTool, fileHandler)

	// Start Server
	log.Println("Figma MCP Server (Go) running on Stdio")
	server.ServeStdio()
}
