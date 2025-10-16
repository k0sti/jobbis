package server

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/auth"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/client"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/config"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/tools"
)

// New creates a new MCP server with registered tools
func New(cfg *config.Config) (*mcp.Server, error) {
	// Create OAuth client for Azure AD B2C
	oauthClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.TokenURL, cfg.OAuthScope)

	// Create API client
	apiClient := client.NewClient(
		oauthClient,
		cfg.APIBaseURL,
		cfg.RateLimitRPS,
		cfg.RateLimitBurst,
		cfg.CacheTTL,
	)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "tyomarkkinatori-mcp-server",
		Version: "1.0.0",
	}, nil)

	// Register search_jobs tool
	searchJobsSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Keywords to search (job title, skills, description)"
			},
			"location": {
				"type": "string",
				"description": "Location or region (e.g., Helsinki, Tampere)"
			},
			"occupation_group": {
				"type": "string",
				"description": "ESCO occupation group code (e.g., 2512 for Software developers)"
			},
			"employer_type": {
				"type": "string",
				"description": "Type of employer",
				"enum": ["company", "public", "nonprofit"]
			},
			"working_hours": {
				"type": "string",
				"description": "Working hours type",
				"enum": ["full-time", "part-time"]
			},
			"duration": {
				"type": "string",
				"description": "Employment duration",
				"enum": ["permanent", "temporary"]
			},
			"published_after": {
				"type": "string",
				"description": "ISO date string (e.g., 2024-01-01)"
			},
			"language": {
				"type": "string",
				"description": "Language for results (default: fi)",
				"enum": ["fi", "sv", "en"]
			},
			"page": {
				"type": "number",
				"description": "Page number for pagination (0-based, default: 0)"
			},
			"page_size": {
				"type": "number",
				"description": "Results per page (100-500, default: 100)"
			}
		}
	}`)

	server.AddTool(&mcp.Tool{
		Name:        "search_jobs",
		Description: "Search for job listings on Työmarkkinatori.fi using the official API",
		InputSchema: searchJobsSchema,
	}, tools.NewSearchJobsHandler(apiClient))

	// Register get_job_details tool
	getJobDetailsSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"job_id": {
				"type": "string",
				"description": "Unique identifier for the job listing"
			}
		},
		"required": ["job_id"]
	}`)

	server.AddTool(&mcp.Tool{
		Name:        "get_job_details",
		Description: "Get detailed information about a specific job listing",
		InputSchema: getJobDetailsSchema,
	}, tools.NewGetJobDetailsHandler(apiClient))

	// Register list_categories tool
	listCategoriesSchema := json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`)

	server.AddTool(&mcp.Tool{
		Name:        "list_categories",
		Description: "Get available job categories, employment types, and regions",
		InputSchema: listCategoriesSchema,
	}, tools.NewListCategoriesHandler())

	return server, nil
}
