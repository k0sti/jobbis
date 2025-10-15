package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/models"
)

// NewListCategoriesHandler creates a new list categories handler
func NewListCategoriesHandler() func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Return static filter options
		options := models.FilterOptions{
			OccupationGroups: []models.OccupationGroup{
				{Code: "2512", NameFI: "Ohjelmistokehittäjät", NameEN: "Software developers"},
				{Code: "2221", NameFI: "Sairaanhoitajat", NameEN: "Nursing professionals"},
				{Code: "2141", NameFI: "Teollisuusinsinöörit", NameEN: "Industrial engineers"},
				{Code: "5223", NameFI: "Myyjät", NameEN: "Shop salespersons"},
				{Code: "2330", NameFI: "Opettajat", NameEN: "Secondary education teachers"},
			},
			EmployerTypes: []models.EmployerType{
				{Code: "01", Name: "Yritykset", Value: "company"},
				{Code: "02", Name: "Julkinen sektori", Value: "public"},
				{Code: "03", Name: "Järjestöt", Value: "nonprofit"},
			},
			WorkingHours: []string{"full-time", "part-time"},
			Durations:    []string{"permanent", "temporary"},
			Regions: []string{
				"Uusimaa",
				"Pirkanmaa",
				"Varsinais-Suomi",
				"Pohjois-Pohjanmaa",
				"Keski-Suomi",
				"Pohjois-Savo",
				"Kanta-Häme",
				"Satakunta",
				"Päijät-Häme",
				"Lappi",
			},
			Languages: []string{"fi", "sv", "en"},
		}

		// Convert to JSON
		jsonBytes, err := json.MarshalIndent(options, "", "  ")
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
