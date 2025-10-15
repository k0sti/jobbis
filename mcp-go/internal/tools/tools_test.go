package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mockCallToolRequest creates a mock CallToolRequest for testing
func mockCallToolRequest(argsJSON string) *mcp.CallToolRequest {
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Arguments: json.RawMessage(argsJSON),
		},
	}
}

func TestNewListCategoriesHandler(t *testing.T) {
	handler := NewListCategoriesHandler()
	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}

	ctx := context.Background()
	req := mockCallToolRequest(`{}`)

	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("Handler should not return error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.IsError {
		t.Errorf("Result should not be an error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}

	// Verify content is valid JSON
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &jsonData); err != nil {
		t.Errorf("Result content should be valid JSON: %v", err)
	}

	// Check that expected fields exist
	expectedFields := []string{"occupation_groups", "employer_types", "working_hours", "durations", "regions", "languages"}
	for _, field := range expectedFields {
		if _, exists := jsonData[field]; !exists {
			t.Errorf("Expected field '%s' in result", field)
		}
	}
}

func TestListCategoriesHandler_Content(t *testing.T) {
	handler := NewListCategoriesHandler()
	ctx := context.Background()
	req := mockCallToolRequest(`{}`)

	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	var data struct {
		OccupationGroups []struct {
			Code   string `json:"code"`
			NameFI string `json:"name_fi"`
		} `json:"occupation_groups"`
		EmployerTypes []struct {
			Code  string `json:"code"`
			Value string `json:"value"`
		} `json:"employer_types"`
		WorkingHours []string `json:"working_hours"`
		Durations    []string `json:"durations"`
		Regions      []string `json:"regions"`
		Languages    []string `json:"languages"`
	}

	if err := json.Unmarshal([]byte(textContent.Text), &data); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Verify we have some occupation groups
	if len(data.OccupationGroups) == 0 {
		t.Error("Expected at least one occupation group")
	}

	// Verify occupation group structure
	if len(data.OccupationGroups) > 0 {
		og := data.OccupationGroups[0]
		if og.Code == "" {
			t.Error("Expected occupation group to have code")
		}
		if og.NameFI == "" {
			t.Error("Expected occupation group to have name_fi")
		}
	}

	// Verify employer types
	if len(data.EmployerTypes) != 3 {
		t.Errorf("Expected 3 employer types, got %d", len(data.EmployerTypes))
	}

	// Verify working hours
	if len(data.WorkingHours) != 2 {
		t.Errorf("Expected 2 working hours options, got %d", len(data.WorkingHours))
	}

	// Verify durations
	if len(data.Durations) != 2 {
		t.Errorf("Expected 2 duration options, got %d", len(data.Durations))
	}

	// Verify regions
	if len(data.Regions) == 0 {
		t.Error("Expected at least one region")
	}

	// Verify languages
	if len(data.Languages) != 3 {
		t.Errorf("Expected 3 languages, got %d", len(data.Languages))
	}
}

func TestGetJobDetailsHandler_MissingJobID(t *testing.T) {
	// Create a mock API client (nil is okay for this test since we're testing error handling)
	handler := NewGetJobDetailsHandler(nil)
	ctx := context.Background()

	// Test with missing job_id
	req := mockCallToolRequest(`{}`)

	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("Handler should not return error, but use IsError flag: %v", err)
	}

	if !result.IsError {
		t.Error("Expected IsError to be true when job_id is missing")
	}

	if len(result.Content) == 0 {
		t.Error("Expected error message in content")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	if textContent.Text == "" {
		t.Error("Expected error message text")
	}
}

func TestGetJobDetailsHandler_InvalidJSON(t *testing.T) {
	handler := NewGetJobDetailsHandler(nil)
	ctx := context.Background()

	// Test with invalid JSON
	req := mockCallToolRequest(`{invalid json`)

	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("Handler should not return error, but use IsError flag: %v", err)
	}

	if !result.IsError {
		t.Error("Expected IsError to be true for invalid JSON")
	}
}

func TestSearchJobsHandler_InvalidJSON(t *testing.T) {
	handler := NewSearchJobsHandler(nil)
	ctx := context.Background()

	// Test with invalid JSON - this should fail at JSON parsing stage
	req := mockCallToolRequest(`{invalid json`)

	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("Handler should not return error, but use IsError flag: %v", err)
	}

	if !result.IsError {
		t.Error("Expected IsError to be true for invalid JSON")
	}
}

func TestNewSearchJobsHandler(t *testing.T) {
	// Test that handler can be created
	handler := NewSearchJobsHandler(nil)
	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
}

func TestNewGetJobDetailsHandler(t *testing.T) {
	// Test that handler can be created
	handler := NewGetJobDetailsHandler(nil)
	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
}
