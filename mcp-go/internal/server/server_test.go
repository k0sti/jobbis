package server

import (
	"os"
	"testing"

	"github.com/yourusername/tyomarkkinatori-mcp/internal/config"
)

func TestNew_Success(t *testing.T) {
	// Set required environment variables
	os.Setenv("CLIENT_ID", "test-client-id")
	os.Setenv("CLIENT_SECRET", "test-client-secret")
	os.Setenv("TENANT_ID", "test-tenant-id")
	defer func() {
		os.Unsetenv("CLIENT_ID")
		os.Unsetenv("CLIENT_SECRET")
		os.Unsetenv("TENANT_ID")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server")
	}
}

func TestNew_WithCustomConfig(t *testing.T) {
	cfg := &config.Config{
		ClientID:       "test-id",
		ClientSecret:   "test-secret",
		TenantID:       "test-tenant",
		APIBaseURL:     "https://test-api.example.com",
		RateLimitRPS:   5.0,
		RateLimitBurst: 20,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server")
	}
}
