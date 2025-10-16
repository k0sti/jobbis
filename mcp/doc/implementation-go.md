# Työmarkkinatori MCP Server - Go Implementation Specification

## Project Setup

### Technology Stack

- **Runtime:** Go 1.21+
- **MCP SDK:** github.com/modelcontextprotocol/go-sdk
- **HTTP Client:** net/http (standard library)
- **OAuth:** golang.org/x/oauth2
- **Build Tool:** go build / just
- **Testing:** testing package (standard library)
- **Linting:** golangci-lint

### Project Structure

```
mcp/
├── cmd/
│   └── tyomarkkinatori-mcp/
│       └── main.go              # Main entry point
├── internal/
│   ├── server/
│   │   └── server.go            # MCP server setup
│   ├── tools/
│   │   ├── search_jobs.go       # Search jobs tool handler
│   │   ├── get_job_details.go   # Job details tool handler
│   │   └── list_categories.go   # Categories tool handler
│   ├── auth/
│   │   └── oauth_client.go      # OAuth 2.0 authentication
│   ├── client/
│   │   ├── api_client.go        # REST API client
│   │   └── parser.go            # JSON response parsing
│   ├── models/
│   │   ├── job.go               # Job models
│   │   └── filters.go           # Filter models
│   ├── utils/
│   │   ├── cache.go             # Caching implementation
│   │   ├── rate_limiter.go      # Rate limiting
│   │   ├── validators.go        # Input validation
│   │   └── logger.go            # Logging utility
│   └── config/
│       └── config.go            # Configuration
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

## Implementation Details

### 1. Main Entry Point (cmd/tyomarkkinatori-mcp/main.go)

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/tyomarkkinatori-mcp/internal/config"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/server"
	mcp "github.com/modelcontextprotocol/go-sdk/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create MCP server
	srv, err := server.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "Shutting down...")
		cancel()
	}()

	// Create stdio transport
	transport := mcp.NewStdioTransport()

	// Run server
	fmt.Fprintln(os.Stderr, "Työmarkkinatori MCP server running on stdio")
	if err := srv.Serve(ctx, transport); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
```

### 2. Configuration (internal/config/config.go)

```go
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// OAuth 2.0 credentials
	ClientID     string
	ClientSecret string
	TenantID     string

	// API configuration
	APIBaseURL string

	// Environment
	Environment string

	// Cache settings
	CacheTTL time.Duration

	// Rate limiting
	RateLimitRPS   float64
	RateLimitBurst int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		ClientID:       os.Getenv("CLIENT_ID"),
		ClientSecret:   os.Getenv("CLIENT_SECRET"),
		TenantID:       os.Getenv("TENANT_ID"),
		APIBaseURL:     getEnvOrDefault("API_BASE_URL", "https://integraatiot.tyomarkkinatori.fi"),
		Environment:    getEnvOrDefault("NODE_ENV", "production"),
		CacheTTL:       getDurationOrDefault("CACHE_TTL_MS", 15*time.Minute),
		RateLimitRPS:   getFloat64OrDefault("RATE_LIMIT_RPS", 2.0),
		RateLimitBurst: getIntOrDefault("RATE_LIMIT_BURST", 10),
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	var errs []string

	if c.ClientID == "" {
		errs = append(errs, "CLIENT_ID is required")
	}
	if c.ClientSecret == "" {
		errs = append(errs, "CLIENT_SECRET is required")
	}
	if c.TenantID == "" {
		errs = append(errs, "TENANT_ID is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("missing required environment variables:\n  - %s\n\nPlease set these in your .env file",
			errs)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getFloat64OrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if ms, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return defaultValue
}
```

### 3. OAuth 2.0 Authentication (internal/auth/oauth_client.go)

```go
package auth

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Client handles OAuth 2.0 authentication
type Client struct {
	config      *clientcredentials.Config
	tokenSource oauth2.TokenSource
	mu          sync.RWMutex
}

// NewClient creates a new OAuth client
func NewClient(clientID, clientSecret, tenantID string) *Client {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{"https://graph.microsoft.com/.default"},
	}

	return &Client{
		config:      config,
		tokenSource: config.TokenSource(context.Background()),
	}
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	log.Println("Requesting OAuth access token")

	token, err := c.tokenSource.Token()
	if err != nil {
		log.Printf("Failed to obtain OAuth token: %v", err)
		return "", fmt.Errorf("authentication failed - check credentials: %w", err)
	}

	log.Println("OAuth token obtained successfully")
	return token.AccessToken, nil
}

// ClearToken clears the cached token (forces refresh on next request)
func (c *Client) ClearToken() {
	// The tokenSource handles caching internally, so we recreate it
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenSource = c.config.TokenSource(context.Background())
}
```

### 4. Data Models (internal/models/job.go)

```go
package models

import "time"

// JobListing represents a job listing
type JobListing struct {
	// Basic information
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Employer       string    `json:"employer"`
	Location       string    `json:"location"`
	EmploymentType string    `json:"employment_type"`

	// Dates
	PublishedDate       time.Time  `json:"published_date"`
	ApplicationDeadline *time.Time `json:"application_deadline,omitempty"`

	// Description
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"` // Full description (only in details)

	// Additional details
	Requirements []string `json:"requirements,omitempty"`
	Benefits     []string `json:"benefits,omitempty"`
	SalaryInfo   string   `json:"salary_info,omitempty"`

	// Link
	URL string `json:"url"`
}

// SearchParams contains parameters for job search
type SearchParams struct {
	// Search & Basic Filters
	Query          string `json:"query,omitempty"`
	Location       string `json:"location,omitempty"`
	OccupationGroup string `json:"occupation_group,omitempty"`

	// Employment Filters
	EmployerType string `json:"employer_type,omitempty"` // company, public, nonprofit
	WorkingHours string `json:"working_hours,omitempty"` // full-time, part-time
	Duration     string `json:"duration,omitempty"`      // permanent, temporary

	// Date & Language
	PublishedAfter string `json:"published_after,omitempty"` // ISO 8601
	Language       string `json:"language,omitempty"`         // fi, sv, en

	// Pagination
	Page     int `json:"page,omitempty"`      // 0-based
	PageSize int `json:"page_size,omitempty"` // 100-500
}

// FilterOptions contains available filter options
type FilterOptions struct {
	OccupationGroups []OccupationGroup `json:"occupation_groups"`
	EmployerTypes    []EmployerType    `json:"employer_types"`
	WorkingHours     []string          `json:"working_hours"`
	Durations        []string          `json:"durations"`
	Regions          []string          `json:"regions"`
	Languages        []string          `json:"languages"`
}

// OccupationGroup represents an ESCO occupation group
type OccupationGroup struct {
	Code   string `json:"code"`
	NameFI string `json:"name_fi"`
	NameEN string `json:"name_en,omitempty"`
	NameSV string `json:"name_sv,omitempty"`
}

// EmployerType represents an employer type option
type EmployerType struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
```

### 5. REST API Client (internal/client/api_client.go)

```go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/yourusername/tyomarkkinatori-mcp/internal/auth"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/models"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/utils"
)

// Client handles API requests to Työmarkkinatori
type Client struct {
	httpClient  *http.Client
	oauthClient *auth.Client
	rateLimiter *utils.RateLimiter
	cache       *utils.Cache
	baseURL     string
}

// NewClient creates a new API client
func NewClient(oauthClient *auth.Client, baseURL string, rateLimitRPS float64, rateLimitBurst int, cacheTTL time.Duration) *Client {
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
			"company":  "01",
			"public":   "02",
			"nonprofit": "03",
		}
		if code, ok := employerTypeMap[params.EmployerType]; ok {
			query.Set("tyollistaja", code)
		}
	}

	// Working hours mapping
	if params.WorkingHours != "" {
		workingHoursMap := map[string]string{
			"full-time":  "01",
			"part-time":  "02",
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
```

### 6. Caching (internal/utils/cache.go)

```go
package utils

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data   interface{}
	expiry time.Time
}

// Cache provides in-memory caching with TTL
type Cache struct {
	store      map[string]cacheEntry
	defaultTTL time.Duration
	mu         sync.RWMutex
}

// NewCache creates a new cache with the specified default TTL
func NewCache(defaultTTL time.Duration) *Cache {
	c := &Cache{
		store:      make(map[string]cacheEntry),
		defaultTTL: defaultTTL,
	}

	// Start cleanup goroutine
	go c.cleanupLoop()

	return c
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.store[key]
	if !found {
		return nil, false
	}

	if time.Now().After(entry.expiry) {
		// Entry expired
		return nil, false
	}

	return entry.data, true
}

// Set stores a value in the cache with the default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = cacheEntry{
		data:   value,
		expiry: time.Now().Add(ttl),
	}
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, key)
}

// Clear removes all values from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]cacheEntry)
}

// cleanupLoop periodically removes expired entries
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.store {
		if now.After(entry.expiry) {
			delete(c.store, key)
		}
	}
}
```

### 7. Rate Limiter (internal/utils/rate_limiter.go)

```go
package utils

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens       float64
	maxTokens    float64
	refillRate   time.Duration
	lastRefill   time.Time
	mu           sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, maxBurst int) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(maxBurst),
		maxTokens:  float64(maxBurst),
		refillRate: time.Duration(float64(time.Second) / requestsPerSecond),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()
		rl.refillTokens()

		if rl.tokens >= 1.0 {
			rl.tokens -= 1.0
			rl.mu.Unlock()
			return nil
		}

		waitTime := rl.refillRate
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			continue
		}
	}
}

func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	timePassed := now.Sub(rl.lastRefill)
	tokensToAdd := float64(timePassed) / float64(rl.refillRate)

	rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
	rl.lastRefill = now
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
```

### 8. Server Setup (internal/server/server.go)

```go
package server

import (
	"context"

	"github.com/yourusername/tyomarkkinatori-mcp/internal/auth"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/client"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/config"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/tools"
	mcp "github.com/modelcontextprotocol/go-sdk/server"
)

// Server is the MCP server
type Server struct {
	mcpServer  *mcp.Server
	apiClient  *client.Client
	config     *config.Config
}

// New creates a new MCP server
func New(cfg *config.Config) (*Server, error) {
	// Create OAuth client
	oauthClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.TenantID)

	// Create API client
	apiClient := client.NewClient(
		oauthClient,
		cfg.APIBaseURL,
		cfg.RateLimitRPS,
		cfg.RateLimitBurst,
		cfg.CacheTTL,
	)

	// Create MCP server
	mcpServer := mcp.NewServer(
		mcp.ServerInfo{
			Name:    "tyomarkkinatori-mcp-server",
			Version: "1.0.0",
		},
		mcp.ServerCapabilities{
			Tools: &mcp.ToolsCapability{},
		},
	)

	srv := &Server{
		mcpServer: mcpServer,
		apiClient: apiClient,
		config:    cfg,
	}

	// Register tool handlers
	srv.registerTools()

	return srv, nil
}

// registerTools registers all tool handlers
func (s *Server) registerTools() {
	// Register search_jobs tool
	s.mcpServer.RegisterTool(
		"search_jobs",
		"Search for job listings on Työmarkkinatori.fi using the official API",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]string{
					"type":        "string",
					"description": "Keywords to search (job title, skills, description)",
				},
				"location": map[string]string{
					"type":        "string",
					"description": "Location or region (e.g., Helsinki, Tampere)",
				},
				"occupation_group": map[string]string{
					"type":        "string",
					"description": "ESCO occupation group code (e.g., 2512 for Software developers)",
				},
				"employer_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of employer",
					"enum":        []string{"company", "public", "nonprofit"},
				},
				"working_hours": map[string]interface{}{
					"type":        "string",
					"description": "Working hours type",
					"enum":        []string{"full-time", "part-time"},
				},
				"duration": map[string]interface{}{
					"type":        "string",
					"description": "Employment duration",
					"enum":        []string{"permanent", "temporary"},
				},
				"published_after": map[string]string{
					"type":        "string",
					"description": "ISO date string (e.g., 2024-01-01)",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Language for results (default: fi)",
					"enum":        []string{"fi", "sv", "en"},
				},
				"page": map[string]interface{}{
					"type":        "number",
					"description": "Page number for pagination (0-based, default: 0)",
				},
				"page_size": map[string]interface{}{
					"type":        "number",
					"description": "Results per page (100-500, default: 100)",
				},
			},
		},
		tools.NewSearchJobsHandler(s.apiClient),
	)

	// Register get_job_details tool
	s.mcpServer.RegisterTool(
		"get_job_details",
		"Get detailed information about a specific job listing",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"job_id": map[string]string{
					"type":        "string",
					"description": "Unique identifier for the job listing",
				},
			},
			"required": []string{"job_id"},
		},
		tools.NewGetJobDetailsHandler(s.apiClient),
	)

	// Register list_categories tool
	s.mcpServer.RegisterTool(
		"list_categories",
		"Get available job categories, employment types, and regions",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		tools.NewListCategoriesHandler(),
	)
}

// Serve starts the server with the given transport
func (s *Server) Serve(ctx context.Context, transport mcp.Transport) error {
	return s.mcpServer.Serve(ctx, transport)
}
```

### 9. Tool Handler Example (internal/tools/search_jobs.go)

```go
package tools

import (
	"context"
	"encoding/json"
	"log"

	"github.com/yourusername/tyomarkkinatori-mcp/internal/client"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/models"
	mcp "github.com/modelcontextprotocol/go-sdk/server"
)

// SearchJobsHandler handles search_jobs tool calls
type SearchJobsHandler struct {
	apiClient *client.Client
}

// NewSearchJobsHandler creates a new search jobs handler
func NewSearchJobsHandler(apiClient *client.Client) mcp.ToolHandler {
	return &SearchJobsHandler{
		apiClient: apiClient,
	}
}

// Handle executes the search_jobs tool
func (h *SearchJobsHandler) Handle(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse parameters
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
	jobs, err := h.apiClient.SearchJobs(ctx, params)
	if err != nil {
		log.Printf("Search jobs error: %v", err)
		return map[string]interface{}{
			"error":   "Failed to search jobs",
			"message": err.Error(),
		}, err
	}

	// Format response
	result := map[string]interface{}{
		"total": len(jobs),
		"jobs":  jobs,
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}

	return string(jsonBytes), nil
}
```

## Configuration

### go.mod

```go
module github.com/yourusername/tyomarkkinatori-mcp

go 1.21

require (
	github.com/modelcontextprotocol/go-sdk v0.1.0
	golang.org/x/oauth2 v0.15.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.19.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
```

### justfile

```justfile
# Build the binary
build:
	go build -o bin/tyomarkkinatori-mcp ./cmd/tyomarkkinatori-mcp

# Run in development mode
run:
	go run ./cmd/tyomarkkinatori-mcp

# Run tests
test:
	go test -v -race -cover ./...

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod download
	go mod verify

# Update dependencies
update:
	go get -u ./...
	go mod tidy

# Build for production (statically linked)
build-prod:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/tyomarkkinatori-mcp ./cmd/tyomarkkinatori-mcp

# Run tests with coverage report
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
```

## Testing

### Example Test (internal/utils/cache_test.go)

```go
package utils

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Test set and get
	cache.Set("key1", "value1")

	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}

func TestCache_Expiry(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	// Set value with short TTL
	cache.Set("key1", "value1")

	// Should be found immediately
	_, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Should not be found after expiry
	_, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	cache.Set("key1", "value1")
	cache.Delete("key1")

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")

	if found1 || found2 {
		t.Error("Expected all keys to be cleared")
	}
}
```

## MCP Client Configuration

See [mcp/README.md](../README.md) for complete MCP client configuration instructions.

