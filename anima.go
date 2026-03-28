// Package anima provides a Go client for the Anima API.
//
// Create a client with your API key:
//
//	client := anima.NewClient("ak_live_...",
//	    anima.WithBaseURL("https://api.anima.com"),
//	)
//
// Resource services (agents, emails, cards, etc.) are available as fields on the
// Client struct. They will be populated in Task 8.
package anima

import (
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production Anima API endpoint.
	DefaultBaseURL = "https://api.anima.com"
	// DefaultTimeout is the default per-request timeout.
	DefaultTimeout = 30 * time.Second
	// DefaultMaxRetries is the default number of retries for failed requests.
	DefaultMaxRetries = 3
	// SDKVersion is the current version of this SDK.
	SDKVersion = "0.1.0"
)

// Client is the Anima API client. Create one with NewClient.
type Client struct {
	// httpClient handles low-level HTTP requests.
	httpClient *httpClient

	// Service fields — populated in Task 8.
	// Agents        *AgentsService
	// Cards         *CardsService
	// Domains       *DomainsService
	// Emails        *EmailsService
	// Messages      *MessagesService
	// Organizations *OrganizationsService
	// Phones        *PhonesService
	// Security      *SecurityService
	// Vault         *VaultService
	// Webhooks      *WebhooksService
}

// NewClient creates a new Anima API client.
//
//	client := anima.NewClient("ak_live_...",
//	    anima.WithTimeout(10 * time.Second),
//	    anima.WithMaxRetries(5),
//	)
func NewClient(apiKey string, opts ...Option) *Client {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	// Strip trailing slashes from base URL.
	cfg.baseURL = strings.TrimRight(cfg.baseURL, "/")

	// Build the underlying HTTP client.
	hc := cfg.httpClient
	if hc == nil {
		hc = &http.Client{
			Timeout: cfg.timeout,
		}
	}

	c := &Client{
		httpClient: &httpClient{
			apiKey:     apiKey,
			baseURL:    cfg.baseURL,
			maxRetries: cfg.maxRetries,
			client:     hc,
		},
	}

	return c
}

// HTTPClient returns the internal httpClient for use by service implementations.
// This is not part of the public API surface.
func (c *Client) HTTPClient() *httpClient {
	return c.httpClient
}
