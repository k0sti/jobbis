package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestConfig_Load_Success(t *testing.T) {
	// Set required environment variables
	os.Setenv("CLIENT_ID", "test-client-id")
	os.Setenv("CLIENT_SECRET", "test-client-secret")
	os.Setenv("TENANT_ID", "test-tenant-id")
	defer func() {
		os.Unsetenv("CLIENT_ID")
		os.Unsetenv("CLIENT_SECRET")
		os.Unsetenv("TENANT_ID")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should succeed: %v", err)
	}

	if cfg.ClientID != "test-client-id" {
		t.Errorf("Expected ClientID='test-client-id', got '%s'", cfg.ClientID)
	}
	if cfg.ClientSecret != "test-client-secret" {
		t.Errorf("Expected ClientSecret='test-client-secret', got '%s'", cfg.ClientSecret)
	}
	if cfg.TenantID != "test-tenant-id" {
		t.Errorf("Expected TenantID='test-tenant-id', got '%s'", cfg.TenantID)
	}
}

func TestConfig_Load_MissingClientID(t *testing.T) {
	// Set only some required variables
	os.Setenv("CLIENT_SECRET", "test-secret")
	os.Setenv("TENANT_ID", "test-tenant")
	defer func() {
		os.Unsetenv("CLIENT_SECRET")
		os.Unsetenv("TENANT_ID")
	}()

	_, err := Load()
	if err == nil {
		t.Error("Expected error when CLIENT_ID is missing")
	}
	if !strings.Contains(err.Error(), "CLIENT_ID") {
		t.Errorf("Error should mention CLIENT_ID: %v", err)
	}
}

func TestConfig_Load_MissingClientSecret(t *testing.T) {
	os.Setenv("CLIENT_ID", "test-id")
	os.Setenv("TENANT_ID", "test-tenant")
	defer func() {
		os.Unsetenv("CLIENT_ID")
		os.Unsetenv("TENANT_ID")
	}()

	_, err := Load()
	if err == nil {
		t.Error("Expected error when CLIENT_SECRET is missing")
	}
	if !strings.Contains(err.Error(), "CLIENT_SECRET") {
		t.Errorf("Error should mention CLIENT_SECRET: %v", err)
	}
}

func TestConfig_Load_MissingTenantID(t *testing.T) {
	os.Setenv("CLIENT_ID", "test-id")
	os.Setenv("CLIENT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("CLIENT_ID")
		os.Unsetenv("CLIENT_SECRET")
	}()

	_, err := Load()
	if err == nil {
		t.Error("Expected error when TENANT_ID is missing")
	}
	if !strings.Contains(err.Error(), "TENANT_ID") {
		t.Errorf("Error should mention TENANT_ID: %v", err)
	}
}

func TestConfig_Load_Defaults(t *testing.T) {
	// Set only required variables
	os.Setenv("CLIENT_ID", "test-id")
	os.Setenv("CLIENT_SECRET", "test-secret")
	os.Setenv("TENANT_ID", "test-tenant")
	defer func() {
		os.Unsetenv("CLIENT_ID")
		os.Unsetenv("CLIENT_SECRET")
		os.Unsetenv("TENANT_ID")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should succeed: %v", err)
	}

	// Check defaults
	if cfg.APIBaseURL != "https://integraatiot.tyomarkkinatori.fi" {
		t.Errorf("Expected default APIBaseURL, got '%s'", cfg.APIBaseURL)
	}
	if cfg.Environment != "production" {
		t.Errorf("Expected default Environment='production', got '%s'", cfg.Environment)
	}
	if cfg.CacheTTL != 15*time.Minute {
		t.Errorf("Expected default CacheTTL=15m, got %v", cfg.CacheTTL)
	}
	if cfg.RateLimitRPS != 2.0 {
		t.Errorf("Expected default RateLimitRPS=2.0, got %f", cfg.RateLimitRPS)
	}
	if cfg.RateLimitBurst != 10 {
		t.Errorf("Expected default RateLimitBurst=10, got %d", cfg.RateLimitBurst)
	}
}

func TestConfig_Load_CustomValues(t *testing.T) {
	// Set all variables including optional ones
	os.Setenv("CLIENT_ID", "test-id")
	os.Setenv("CLIENT_SECRET", "test-secret")
	os.Setenv("TENANT_ID", "test-tenant")
	os.Setenv("API_BASE_URL", "https://custom-api.example.com")
	os.Setenv("NODE_ENV", "development")
	os.Setenv("CACHE_TTL_MS", "60000") // 1 minute
	os.Setenv("RATE_LIMIT_RPS", "5.5")
	os.Setenv("RATE_LIMIT_BURST", "20")
	defer func() {
		os.Unsetenv("CLIENT_ID")
		os.Unsetenv("CLIENT_SECRET")
		os.Unsetenv("TENANT_ID")
		os.Unsetenv("API_BASE_URL")
		os.Unsetenv("NODE_ENV")
		os.Unsetenv("CACHE_TTL_MS")
		os.Unsetenv("RATE_LIMIT_RPS")
		os.Unsetenv("RATE_LIMIT_BURST")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should succeed: %v", err)
	}

	if cfg.APIBaseURL != "https://custom-api.example.com" {
		t.Errorf("Expected custom APIBaseURL, got '%s'", cfg.APIBaseURL)
	}
	if cfg.Environment != "development" {
		t.Errorf("Expected Environment='development', got '%s'", cfg.Environment)
	}
	if cfg.CacheTTL != 1*time.Minute {
		t.Errorf("Expected CacheTTL=1m, got %v", cfg.CacheTTL)
	}
	if cfg.RateLimitRPS != 5.5 {
		t.Errorf("Expected RateLimitRPS=5.5, got %f", cfg.RateLimitRPS)
	}
	if cfg.RateLimitBurst != 20 {
		t.Errorf("Expected RateLimitBurst=20, got %d", cfg.RateLimitBurst)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		expectError bool
		errorText   string
	}{
		{
			name: "Valid configuration",
			cfg: Config{
				ClientID:     "id",
				ClientSecret: "secret",
				TenantID:     "tenant",
			},
			expectError: false,
		},
		{
			name: "Missing ClientID",
			cfg: Config{
				ClientSecret: "secret",
				TenantID:     "tenant",
			},
			expectError: true,
			errorText:   "CLIENT_ID",
		},
		{
			name: "Missing ClientSecret",
			cfg: Config{
				ClientID: "id",
				TenantID: "tenant",
			},
			expectError: true,
			errorText:   "CLIENT_SECRET",
		},
		{
			name: "Missing TenantID",
			cfg: Config{
				ClientID:     "id",
				ClientSecret: "secret",
			},
			expectError: true,
			errorText:   "TENANT_ID",
		},
		{
			name:        "All missing",
			cfg:         Config{},
			expectError: true,
			errorText:   "CLIENT_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error")
				} else if tt.errorText != "" && !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorText, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
