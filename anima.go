// Package anima provides a Go client for the Anima API.
//
// Create a client with your API key:
//
//	client := anima.NewClient("ak_live_...",
//	    anima.WithBaseURL("https://api.useanima.sh"),
//	)
//
// Resource services are available as fields on the Client struct:
//
//	client.Agents.Create(ctx, params)
//	client.Messages.SendEmail(ctx, params)
//	client.Cards.Create(ctx, params)
package anima

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production Anima API endpoint.
	DefaultBaseURL = "https://api.useanima.sh"
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

	// Anomaly provides methods for anomaly detection alerts, rules, baselines, and quarantine.
	Anomaly *AnomalyService
	// Audit provides methods for querying and exporting immutable audit logs.
	Audit *AuditService

	// A2A provides methods for agent-to-agent communication.
	A2A *A2AService
	// Addresses provides methods for managing agent physical addresses.
	Addresses *AddressesService
	// Agents provides methods for managing agents.
	Agents *AgentsService
	// Cards provides methods for managing cards, policies, and transactions.
	Cards *CardsService
	// Compliance provides methods for compliance controls, reports, dashboards, and DSARs.
	Compliance *ComplianceService
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
	// VaultOAuth provides methods for managing OAuth connections.
	VaultOAuth *VaultOAuthService
	// Wallet provides methods for managing agent crypto wallets.
	Wallet *WalletService
	// Voices provides methods for browsing the voice catalog.
	Voices *VoicesService
	// Calls provides methods for managing voice calls.
	Calls *CallsService
	// Webhooks provides methods for managing webhooks.
	Webhooks *WebhooksService
}

// NewClient creates a new Anima API client.
//
// If apiKey is empty, the ANIMA_API_KEY environment variable is used.
// If no key is found, NewClient panics.
//
// The base URL defaults to DefaultBaseURL, but can be overridden with
// WithBaseURL or the ANIMA_API_URL environment variable.
//
//	client := anima.NewClient("",  // uses ANIMA_API_KEY env var
//	    anima.WithTimeout(10 * time.Second),
//	)
func NewClient(apiKey string, opts ...Option) *Client {
	// Environment variable fallback for API key.
	if apiKey == "" {
		apiKey = os.Getenv("ANIMA_API_KEY")
	}
	if apiKey == "" {
		panic("anima: missing API key. Pass it to NewClient or set the ANIMA_API_KEY environment variable.")
	}

	cfg := defaultConfig()

	// Environment variable fallback for base URL.
	if envURL := os.Getenv("ANIMA_API_URL"); envURL != "" {
		cfg.baseURL = envURL
	}

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

	debug := os.Getenv("ANIMA_LOG") == "debug"
	var logger *slog.Logger
	if debug {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	internal := &httpClient{
		apiKey:     apiKey,
		baseURL:    cfg.baseURL,
		maxRetries: cfg.maxRetries,
		client:     hc,
		debug:      debug,
		logger:     logger,
	}

	c := &Client{
		httpClient:    internal,
		Anomaly:       newAnomalyService(internal),
		Audit:         newAuditService(internal),
		A2A:           newA2AService(internal),
		Addresses:     newAddressesService(internal),
		Agents:        newAgentsService(internal),
		Cards:         newCardsService(internal),
		Compliance:    newComplianceService(internal),
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
		VaultOAuth:    newVaultOAuthService(internal),
		Wallet:        newWalletService(internal),
		Voices:        newVoicesService(internal),
		Calls:         newCallsService(internal),
		Webhooks:      newWebhooksService(internal),
	}

	return c
}

// HTTPClient returns the internal httpClient for use by service implementations.
// This is not part of the public API surface.
func (c *Client) HTTPClient() *httpClient {
	return c.httpClient
}
