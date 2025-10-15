package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/config"
	"github.com/yourusername/tyomarkkinatori-mcp/internal/server"
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

	// Run server with stdio transport
	fmt.Fprintln(os.Stderr, "Työmarkkinatori MCP server running on stdio")
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
