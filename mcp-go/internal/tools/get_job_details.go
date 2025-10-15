package tools

import (
	"context"
	"encoding/json"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/client"
)

// NewGetJobDetailsHandler creates a new get job details handler
func NewGetJobDetailsHandler(apiClient *client.Client) func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse parameters
		var args map[string]interface{}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Failed to parse arguments: " + err.Error()}},
				IsError: true,
			}, nil
		}

		jobID, ok := args["job_id"].(string)
		if !ok || jobID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "job_id is required and must be a string"}},
				IsError: true,
			}, nil
		}

		log.Printf("Fetching job details for ID: %s", jobID)

		// Execute request
		job, err := apiClient.GetJobDetails(ctx, jobID)
		if err != nil {
			log.Printf("Get job details error: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Failed to get job details: " + err.Error()}},
				IsError: true,
			}, nil
		}

		// Convert to JSON
		jsonBytes, err := json.MarshalIndent(job, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Failed to format result: " + err.Error()}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
		}, nil
	}
}
