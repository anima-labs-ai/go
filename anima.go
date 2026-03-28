// Package anima provides a Go client for the Anima API.
//
// Create a client with your API key:
//
//	client := anima.NewClient("ak_live_...",
//	    anima.WithBaseURL("https://api.anima.com"),
//	)
//
// Resource services are available as fields on the Client struct:
//
//	client.Agents.Create(ctx, params)
//	client.Messages.SendEmail(ctx, params)
//	client.Cards.Create(ctx, params)
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

	// A2A provides methods for agent-to-agent communication.
	A2A *A2AService
	// Addresses provides methods for managing agent physical addresses.
	Addresses *AddressesService
	// Agents provides methods for managing agents.
	Agents *AgentsService
	// Cards provides methods for managing cards, policies, and transactions.
	Cards *CardsService
	// Domains provides methods for managing email domains.
	Domains *DomainsService
	// Emails provides methods for listing emails and managing attachments.
	Emails *EmailsService
	// Identity provides methods for managing agent decentralized identity (DID).
	Identity *IdentityService
	// Messages provides methods for sending and listing messages.
	Messages *MessagesService
	// Organizations provides methods for managing organizations.
	Organizations *OrganizationsService
	// Phones provides methods for provisioning and managing phone numbers.
	Phones *PhonesService
	// Pods provides methods for managing agent compute pods.
	Pods *PodsService
	// Registry provides methods for the public agent registry.
	Registry *RegistryService
	// Security provides methods for content scanning and security events.
	Security *SecurityService
	// Vault provides methods for managing the agent credential vault.
	Vault *VaultService
	// Wallet provides methods for managing agent crypto wallets.
	Wallet *WalletService
	// Webhooks provides methods for managing webhooks.
	Webhooks *WebhooksService
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

	internal := &httpClient{
		apiKey:     apiKey,
		baseURL:    cfg.baseURL,
		maxRetries: cfg.maxRetries,
		client:     hc,
	}

	c := &Client{
		httpClient:    internal,
		A2A:           newA2AService(internal),
		Addresses:     newAddressesService(internal),
		Agents:        newAgentsService(internal),
		Cards:         newCardsService(internal),
		Domains:       newDomainsService(internal),
		Emails:        newEmailsService(internal),
		Identity:      newIdentityService(internal),
		Messages:      newMessagesService(internal),
		Organizations: newOrganizationsService(internal),
		Phones:        newPhonesService(internal),
		Pods:          newPodsService(internal),
		Registry:      newRegistryService(internal),
		Security:      newSecurityService(internal),
		Vault:         newVaultService(internal),
		Wallet:        newWalletService(internal),
		Webhooks:      newWebhooksService(internal),
	}

	return c
}

// HTTPClient returns the internal httpClient for use by service implementations.
// This is not part of the public API surface.
func (c *Client) HTTPClient() *httpClient {
	return c.httpClient
}
