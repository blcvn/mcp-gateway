package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/blcvn/backend/services/ba-mcp-server/pkg/mcp"
)

func main() {
	server := mcp.NewServer("jira-mcp-go", "1.0.0")

	// Define tool: jira_search_issues
	searchTool := mcp.Tool{
		Name:        "jira_search_issues",
		Description: "Search Jira issues using JQL",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"jql": map[string]interface{}{
					"type":        "string",
					"description": "The JQL query string",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Max results to return",
				},
			},
			"required": []string{"jql"},
		},
	}

	searchHandler := func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		jql, ok := args["jql"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid 'jql' argument")
		}

		// Mock Implementation
		log.Printf("Searching Jira with JQL: %s", jql)

		results := []map[string]string{
			{"key": "BA-101", "summary": "Implement Login Flow", "status": "In Progress"},
			{"key": "BA-102", "summary": "Design Database Schema", "status": "Done"},
		}

		responsePayload := map[string]interface{}{
			"issues": results,
			"jql":    jql,
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
	log.Println("Jira MCP Server (Go) running on Stdio")
	server.ServeStdio()
}
