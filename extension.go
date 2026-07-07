package anima

import (
	"context"
	"net/http"
)

// ConnectExtensionParams contains the parameters for issuing a browser
// extension connect handshake.
//
// Authentication determines how AgentID is used: with a master key, AgentID
// is required to select which agent the extension connects as; with an agent
// key, AgentID must be omitted (the key already identifies the agent). Unset
// fields are dropped from the request via omitempty.
type ConnectExtensionParams struct {
	// AgentID selects the agent to connect as. Required with a master key,
	// omitted with an agent key.
	AgentID string `json:"agentId,omitempty"`
	// TTL controls how long the resulting connection is valid. One of
	// "15m", "1h", or "session". Optional.
	TTL string `json:"ttl,omitempty"`
}

// ConnectExtensionResult is returned by Connect. It carries the URL the
// extension should open to complete the handshake, along with expiry and the
// policy that will govern the session. It never contains a secret.
type ConnectExtensionResult struct {
	// AgentID is the agent the extension is connecting as.
	AgentID string `json:"agentId"`
	// ConnectURL is the URL the extension opens to complete the handshake.
	ConnectURL string `json:"connectUrl"`
	// ExpiresAt is when the resulting connection expires. Nullable — a
	// "session" TTL yields no fixed expiry.
	ExpiresAt *string `json:"expiresAt"`
	// ExchangeExpiresAt is when the one-time exchange window closes.
	ExchangeExpiresAt string `json:"exchangeExpiresAt"`
	// Policy is the session policy: "session" or "pre_approved".
	Policy string `json:"policy"`
}

// ExtensionService provides methods for connecting browser extensions to an
// agent for headless (worker-driven) sessions.
type ExtensionService struct {
	client *httpClient
}

// newExtensionService creates a new ExtensionService.
func newExtensionService(c *httpClient) *ExtensionService {
	return &ExtensionService{client: c}
}

// Connect issues a browser extension connect handshake for an agent.
//
// With a master key, set params.AgentID to the target agent. With an agent
// key, leave params.AgentID empty. The returned result contains the connect
// URL and exchange window; no secret is returned.
func (s *ExtensionService) Connect(ctx context.Context, params ConnectExtensionParams) (*ConnectExtensionResult, error) {
	result, err := Do[ConnectExtensionResult](ctx, s.client, http.MethodPost, "/extension/connect", params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
