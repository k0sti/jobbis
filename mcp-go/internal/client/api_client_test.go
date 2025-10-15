package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/tyomarkkinatori-mcp/internal/models"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/utils"
)

// mockOAuthClient is a simple mock for testing
type mockOAuthClient struct{}

func (m *mockOAuthClient) GetAccessToken(ctx context.Context) (string, error) {
	return "mock-access-token", nil
}

// loadTestData loads JSON fixture from testdata directory
func loadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test data %s: %v", filename, err)
	}
	return data
}

// mockHTTPServer creates a test server that returns fixture data
func mockHTTPServer(t *testing.T, responseFile string, statusCode int) *httptest.Server {
	t.Helper()
	data := loadTestData(t, responseFile)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header
		if auth := r.Header.Get("Authorization"); auth == "" {
			t.Error("Expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write(data)
	}))
}

func TestClient_SearchJobs_Success(t *testing.T) {
	server := mockHTTPServer(t, "job_search_response.json", http.StatusOK)
	defer server.Close()

	cache := utils.NewCache(time.Minute)
	rateLimiter := utils.NewRateLimiter(10, 10) // High limits for testing
	oauthClient := &mockOAuthClient{} // Mock auth client

	client := &Client{
		baseURL:     server.URL,
		httpClient:  server.Client(),
		oauthClient: oauthClient,
		cache:       cache,
		rateLimiter: rateLimiter,
	}

	tests := []struct {
		name          string
		params        models.SearchParams
		expectedCount int
		expectedFirst string
	}{
		{
			name: "basic search",
			params: models.SearchParams{
				Query: "ohjelmistokehittäjä",
			},
			expectedCount: 2,
			expectedFirst: "Senior Ohjelmistokehittäjä",
		},
		{
			name: "search with location",
			params: models.SearchParams{
				Query:    "developer",
				Location: "Helsinki",
			},
			expectedCount: 2,
			expectedFirst: "Senior Ohjelmistokehittäjä",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			jobs, err := client.SearchJobs(ctx, tt.params)

			if err != nil {
				t.Fatalf("SearchJobs failed: %v", err)
			}

			if len(jobs) != tt.expectedCount {
				t.Errorf("Expected %d jobs, got %d", tt.expectedCount, len(jobs))
			}

			if len(jobs) > 0 && jobs[0].Title != tt.expectedFirst {
				t.Errorf("Expected first job title %q, got %q", tt.expectedFirst, jobs[0].Title)
			}
		})
	}
}

func TestClient_SearchJobs_Caching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		data := loadTestData(t, "job_search_response.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer server.Close()

	cache := utils.NewCache(time.Minute)
	rateLimiter := utils.NewRateLimiter(10, 10)
	oauthClient := &mockOAuthClient{}

	client := &Client{
		baseURL:     server.URL,
		httpClient:  server.Client(),
		oauthClient: oauthClient,
		cache:       cache,
		rateLimiter: rateLimiter,
	}

	ctx := context.Background()
	params := models.SearchParams{Query: "test"}

	// First call - should hit API
	_, err := client.SearchJobs(ctx, params)
	if err != nil {
		t.Fatalf("First search failed: %v", err)
	}

	// Second call - should use cache
	_, err = client.SearchJobs(ctx, params)
	if err != nil {
		t.Fatalf("Second search failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 API call (cached second), got %d", callCount)
	}
}

func TestClient_GetJobDetails_Success(t *testing.T) {
	server := mockHTTPServer(t, "job_details_response.json", http.StatusOK)
	defer server.Close()

	cache := utils.NewCache(time.Minute)
	rateLimiter := utils.NewRateLimiter(10, 10)
	oauthClient := &mockOAuthClient{}

	client := &Client{
		baseURL:     server.URL,
		httpClient:  server.Client(),
		oauthClient: oauthClient,
		cache:       cache,
		rateLimiter: rateLimiter,
	}

	ctx := context.Background()
	job, err := client.GetJobDetails(ctx, "test-job-1")

	if err != nil {
		t.Fatalf("GetJobDetails failed: %v", err)
	}

	// Verify job details match fixture
	if job.ID != "test-job-1" {
		t.Errorf("Expected ID test-job-1, got %s", job.ID)
	}
	if job.Title != "Senior Ohjelmistokehittäjä" {
		t.Errorf("Expected title 'Senior Ohjelmistokehittäjä', got %s", job.Title)
	}
	if job.Employer != "TechFinn Oy" {
		t.Errorf("Expected employer TechFinn Oy, got %s", job.Employer)
	}
	if job.Location != "Helsinki" {
		t.Errorf("Expected location Helsinki, got %s", job.Location)
	}
	if job.SalaryInfo != "4500-6500 EUR/kk" {
		t.Errorf("Expected salary info '4500-6500 EUR/kk', got '%s'", job.SalaryInfo)
	}
}

func TestClient_SearchJobs_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  bool
	}{
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			wantError:  true,
		},
		{
			name:       "500 server error",
			statusCode: http.StatusInternalServerError,
			wantError:  true,
		},
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(map[string]string{"error": "test error"})
			}))
			defer server.Close()

			cache := utils.NewCache(time.Minute)
			rateLimiter := utils.NewRateLimiter(10, 10)
			oauthClient := &mockOAuthClient{}

			client := &Client{
				baseURL:     server.URL,
				httpClient:  server.Client(),
				oauthClient: oauthClient,
				cache:       cache,
				rateLimiter: rateLimiter,
			}

			ctx := context.Background()
			_, err := client.SearchJobs(ctx, models.SearchParams{Query: "test"})

			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		data := loadTestData(t, "job_search_response.json")
		w.Write(data)
	}))
	defer server.Close()

	cache := utils.NewCache(time.Minute)
	rateLimiter := utils.NewRateLimiter(10, 10)
	oauthClient := &mockOAuthClient{}

	client := &Client{
		baseURL:     server.URL,
		httpClient:  server.Client(),
		oauthClient: oauthClient,
		cache:       cache,
		rateLimiter: rateLimiter,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.SearchJobs(ctx, models.SearchParams{Query: "test"})

	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}
