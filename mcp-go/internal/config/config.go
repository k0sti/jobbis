package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
			strings.Join(errs, "\n  - "))
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
