package auth

import (
	"context"
	"fmt"
	"log"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Client handles OAuth 2.0 authentication
type Client struct {
	config      *clientcredentials.Config
	tokenSource oauth2.TokenSource
	mu          sync.RWMutex
}

// NewClient creates a new OAuth client with Azure AD B2C
func NewClient(clientID, clientSecret, tokenURL, scope string) *Client {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{scope},
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
