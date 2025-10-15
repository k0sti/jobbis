package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/yourusername/tyomarkkinatori-mcp/internal/models"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/utils"
)

// OAuthClient is the interface for OAuth authentication
type OAuthClient interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Client handles API requests to Työmarkkinatori
type Client struct {
	httpClient  *http.Client
	oauthClient OAuthClient
	rateLimiter *utils.RateLimiter
	cache       *utils.Cache
	baseURL     string
}

// NewClient creates a new API client
func NewClient(oauthClient OAuthClient, baseURL string, rateLimitRPS float64, rateLimitBurst int, cacheTTL time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		oauthClient: oauthClient,
		rateLimiter: utils.NewRateLimiter(rateLimitRPS, rateLimitBurst),
		cache:       utils.NewCache(cacheTTL),
		baseURL:     baseURL,
	}
}

// SearchJobs searches for job listings
func (c *Client) SearchJobs(ctx context.Context, params models.SearchParams) ([]models.JobListing, error) {
	cacheKey := fmt.Sprintf("search:%+v", params)

	// Check cache
	if cached, found := c.cache.Get(cacheKey); found {
		if jobs, ok := cached.([]models.JobListing); ok {
			return jobs, nil
		}
	}

	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Get OAuth token
	token, err := c.oauthClient.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Build request URL
	apiURL := fmt.Sprintf("%s/jobpostingprovider/v1/tyopaikat", c.baseURL)
	reqURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	query := buildAPIParams(params)
	reqURL.RawQuery = query.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp struct {
		Tulokset []map[string]interface{} `json:"tulokset"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to JobListing
	jobs := parseJobListings(apiResp.Tulokset)

	// Cache results
	c.cache.Set(cacheKey, jobs)

	return jobs, nil
}

// GetJobDetails retrieves details for a specific job
func (c *Client) GetJobDetails(ctx context.Context, jobID string) (*models.JobListing, error) {
	cacheKey := fmt.Sprintf("job:%s", jobID)

	// Check cache
	if cached, found := c.cache.Get(cacheKey); found {
		if job, ok := cached.(*models.JobListing); ok {
			return job, nil
		}
	}

	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Get OAuth token
	token, err := c.oauthClient.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Build request URL
	apiURL := fmt.Sprintf("%s/jobpostingprovider/v1/tyopaikka/%s", c.baseURL, jobID)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to JobListing
	job := parseJobDetails(data)

	// Cache with longer TTL (1 hour)
	c.cache.SetWithTTL(cacheKey, job, time.Hour)

	return job, nil
}

// buildAPIParams converts SearchParams to URL query parameters
func buildAPIParams(params models.SearchParams) url.Values {
	query := url.Values{}

	// Pagination
	query.Set("sivu", fmt.Sprintf("%d", params.Page))
	if params.PageSize > 0 {
		query.Set("maara", fmt.Sprintf("%d", params.PageSize))
	} else {
		query.Set("maara", "100")
	}

	// Language
	if params.Language != "" {
		query.Set("kieli", params.Language)
	} else {
		query.Set("kieli", "fi")
	}

	// Search query
	if params.Query != "" {
		query.Set("hakusana", params.Query)
	}

	// Location
	if params.Location != "" {
		query.Set("sijainti", params.Location)
	}

	// Occupation group
	if params.OccupationGroup != "" {
		query.Set("ammattiryhmä", params.OccupationGroup)
	}

	// Employer type mapping
	if params.EmployerType != "" {
		employerTypeMap := map[string]string{
			"company":   "01",
			"public":    "02",
			"nonprofit": "03",
		}
		if code, ok := employerTypeMap[params.EmployerType]; ok {
			query.Set("tyollistaja", code)
		}
	}

	// Working hours mapping
	if params.WorkingHours != "" {
		workingHoursMap := map[string]string{
			"full-time": "01",
			"part-time": "02",
		}
		if code, ok := workingHoursMap[params.WorkingHours]; ok {
			query.Set("työaika", code)
		}
	}

	// Duration mapping
	if params.Duration != "" {
		durationMap := map[string]string{
			"permanent": "01",
			"temporary": "02",
		}
		if code, ok := durationMap[params.Duration]; ok {
			query.Set("työn jatkuvuus", code)
		}
	}

	// Published after
	if params.PublishedAfter != "" {
		query.Set("julkaisupvm", params.PublishedAfter)
	}

	return query
}

// parseJobListings converts API response to JobListing slice
func parseJobListings(results []map[string]interface{}) []models.JobListing {
	jobs := make([]models.JobListing, 0, len(results))

	for _, item := range results {
		job := models.JobListing{
			ID:             getStringOrEmpty(item, "id", "ilmoitusId"),
			Title:          getStringOrEmpty(item, "otsikko", "ammattinimike"),
			Employer:       getStringOrEmpty(item, "tyonantaja"),
			Location:       formatLocation(item),
			EmploymentType: formatEmploymentType(item),
			PublishedDate:  parseDate(getStringOrEmpty(item, "julkaisupvm", "luontipvm")),
			Summary:        getStringOrEmpty(item, "kuvaus", "tehtavakuvaus"),
		}

		// Parse application deadline
		if hakuaika, ok := item["hakuaika"].(map[string]interface{}); ok {
			if paattyy, ok := hakuaika["paattyy"].(string); ok {
				deadline := parseDate(paattyy)
				job.ApplicationDeadline = &deadline
			}
		}

		// Build URL
		job.URL = fmt.Sprintf("https://tyomarkkinatori.fi/tyopaikka/%s", job.ID)

		jobs = append(jobs, job)
	}

	return jobs
}

// parseJobDetails converts API response to JobListing
func parseJobDetails(data map[string]interface{}) *models.JobListing {
	job := &models.JobListing{
		ID:             getStringOrEmpty(data, "id", "ilmoitusId"),
		Title:          getStringOrEmpty(data, "otsikko", "ammattinimike"),
		Employer:       getStringOrEmpty(data, "tyonantaja"),
		Location:       formatLocation(data),
		EmploymentType: formatEmploymentType(data),
		PublishedDate:  parseDate(getStringOrEmpty(data, "julkaisupvm", "luontipvm")),
		Summary:        getStringOrEmpty(data, "kuvaus", "tehtavakuvaus"),
		Description:    getStringOrEmpty(data, "kuvausTaysi", "taysKuvaus"),
		SalaryInfo:     getStringOrEmpty(data, "palkka"),
	}

	// Parse application deadline
	if hakuaika, ok := data["hakuaika"].(map[string]interface{}); ok {
		if paattyy, ok := hakuaika["paattyy"].(string); ok {
			deadline := parseDate(paattyy)
			job.ApplicationDeadline = &deadline
		}
	}

	// Build URL
	job.URL = fmt.Sprintf("https://tyomarkkinatori.fi/tyopaikka/%s", job.ID)

	return job
}

// Helper functions

func getStringOrEmpty(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key].(string); ok && val != "" {
			return val
		}
	}
	return ""
}

func formatLocation(data map[string]interface{}) string {
	if sijainti, ok := data["sijainti"].(map[string]interface{}); ok {
		if kunta, ok := sijainti["kunta"].(string); ok {
			return kunta
		}
		if maakunta, ok := sijainti["maakunta"].(string); ok {
			return maakunta
		}
	}
	if tyopaikka, ok := data["tyopaikka"].(string); ok {
		return tyopaikka
	}
	return "Finland"
}

func formatEmploymentType(data map[string]interface{}) string {
	typeMap := map[string]string{
		"01": "kokoaikainen",
		"02": "osa-aikainen",
		"03": "määräaikainen",
	}

	typeStr := getStringOrEmpty(data, "tyosuhteenLaji", "tyonLaji")
	if mapped, ok := typeMap[typeStr]; ok {
		return mapped
	}
	if typeStr != "" {
		return typeStr
	}
	return "kokoaikainen"
}

func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	// Try multiple formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Now()
}
