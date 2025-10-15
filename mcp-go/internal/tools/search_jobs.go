package tools

import (
	"context"
	"encoding/json"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/client"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/models"
)

// NewSearchJobsHandler creates a new search jobs handler
func NewSearchJobsHandler(apiClient *client.Client) func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse parameters
		var args map[string]interface{}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Failed to parse arguments: " + err.Error()}},
				IsError: true,
			}, nil
		}

		params := models.SearchParams{}

		if query, ok := args["query"].(string); ok {
			params.Query = query
		}
		if location, ok := args["location"].(string); ok {
			params.Location = location
		}
		if occupationGroup, ok := args["occupation_group"].(string); ok {
			params.OccupationGroup = occupationGroup
		}
		if employerType, ok := args["employer_type"].(string); ok {
			params.EmployerType = employerType
		}
		if workingHours, ok := args["working_hours"].(string); ok {
			params.WorkingHours = workingHours
		}
		if duration, ok := args["duration"].(string); ok {
			params.Duration = duration
		}
		if publishedAfter, ok := args["published_after"].(string); ok {
			params.PublishedAfter = publishedAfter
		}
		if language, ok := args["language"].(string); ok {
			params.Language = language
		}
		if page, ok := args["page"].(float64); ok {
			params.Page = int(page)
		}
		if pageSize, ok := args["page_size"].(float64); ok {
			params.PageSize = int(pageSize)
		}

		log.Printf("Searching jobs with params: %+v", params)

		// Execute search
		jobs, err := apiClient.SearchJobs(ctx, params)
		if err != nil {
			log.Printf("Search jobs error: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Failed to search jobs: " + err.Error()}},
				IsError: true,
			}, nil
		}

		// Format response
		result := map[string]interface{}{
			"total": len(jobs),
			"jobs":  jobs,
		}

		// Convert to JSON
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
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
