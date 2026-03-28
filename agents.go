package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AgentStatus represents the status of an agent.
type AgentStatus string

const (
	AgentStatusActive    AgentStatus = "ACTIVE"
	AgentStatusSuspended AgentStatus = "SUSPENDED"
	AgentStatusDeleted   AgentStatus = "DELETED"
)

// EmailIdentity represents an email identity attached to an agent.
type EmailIdentity struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Domain    string `json:"domain"`
	LocalPart string `json:"localPart"`
	IsPrimary bool   `json:"isPrimary"`
	Verified  bool   `json:"verified"`
	CreatedAt string `json:"createdAt"`
}

// PhoneCapabilities describes what a phone number can do.
type PhoneCapabilities struct {
	SMS   bool `json:"sms"`
	MMS   bool `json:"mms"`
	Voice bool `json:"voice"`
}

// PhoneProvider is the telephony provider for a phone number.
type PhoneProvider string

const (
	PhoneProviderTelnyx PhoneProvider = "TELNYX"
	PhoneProviderTwilio PhoneProvider = "TWILIO"
)

// TenDLCStatus represents the 10DLC registration status.
type TenDLCStatus string

const (
	TenDLCStatusPending      TenDLCStatus = "PENDING"
	TenDLCStatusRegistered   TenDLCStatus = "REGISTERED"
	TenDLCStatusRejected     TenDLCStatus = "REJECTED"
	TenDLCStatusNotRequired  TenDLCStatus = "NOT_REQUIRED"
)

// PhoneIdentity represents a phone identity attached to an agent.
type PhoneIdentity struct {
	ID           string            `json:"id"`
	PhoneNumber  string            `json:"phoneNumber"`
	Provider     PhoneProvider     `json:"provider"`
	ProviderID   *string           `json:"providerId"`
	Capabilities PhoneCapabilities `json:"capabilities"`
	TenDLCStatus TenDLCStatus      `json:"tenDlcStatus"`
	IsPrimary    bool              `json:"isPrimary"`
	CreatedAt    string            `json:"createdAt"`
}

// Agent represents an Anima agent.
type Agent struct {
	ID              string                 `json:"id"`
	OrgID           string                 `json:"orgId"`
	Name            string                 `json:"name"`
	Slug            string                 `json:"slug"`
	Status          AgentStatus            `json:"status"`
	APIKeyPrefix    *string                `json:"apiKeyPrefix"`
	Metadata        map[string]interface{} `json:"metadata"`
	EmailIdentities []EmailIdentity        `json:"emailIdentities"`
	PhoneIdentities []PhoneIdentity        `json:"phoneIdentities"`
	CreatedAt       string                 `json:"createdAt"`
	UpdatedAt       string                 `json:"updatedAt"`
}

// CreateAgentParams contains the parameters for creating an agent.
type CreateAgentParams struct {
	OrgID          string                 `json:"orgId"`
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	Email          string                 `json:"email,omitempty"`
	ProvisionPhone bool                   `json:"provisionPhone,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateAgentParams contains the parameters for updating an agent.
type UpdateAgentParams struct {
	Name     string                 `json:"name,omitempty"`
	Slug     string                 `json:"slug,omitempty"`
	Status   AgentStatus            `json:"status,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentListParams contains parameters for listing agents.
type AgentListParams struct {
	ListParams
	OrgID  string
	Status AgentStatus
	Query  string
}

// ToQuery converts AgentListParams to URL query values.
func (p AgentListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.OrgID != "" {
		q.Set("orgId", p.OrgID)
	}
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	if p.Query != "" {
		q.Set("query", p.Query)
	}
	return q
}

// RotateKeyResult contains the result of an API key rotation.
type RotateKeyResult struct {
	APIKey       string `json:"apiKey"`
	APIKeyPrefix string `json:"apiKeyPrefix"`
}

// AgentsService provides methods for managing agents.
type AgentsService struct {
	client *httpClient
}

// newAgentsService creates a new AgentsService.
func newAgentsService(c *httpClient) *AgentsService {
	return &AgentsService{client: c}
}

// Create creates a new agent.
func (s *AgentsService) Create(ctx context.Context, params CreateAgentParams) (*Agent, error) {
	agent, err := Do[Agent](ctx, s.client, http.MethodPost, "/agents", params, nil)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Get retrieves an agent by ID.
func (s *AgentsService) Get(ctx context.Context, id string) (*Agent, error) {
	agent, err := Do[Agent](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Update updates an existing agent.
func (s *AgentsService) Update(ctx context.Context, id string, params UpdateAgentParams) (*Agent, error) {
	agent, err := Do[Agent](ctx, s.client, http.MethodPatch, fmt.Sprintf("/agents/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Delete deletes an agent.
func (s *AgentsService) Delete(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/agents/%s", id), nil, nil)
	return err
}

// List returns a paginated list of agents.
func (s *AgentsService) List(ctx context.Context, params *AgentListParams) (*Page[Agent], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[Agent]](ctx, s.client, http.MethodGet, "/agents", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging returns an iterator that automatically paginates through all agents.
func (s *AgentsService) ListAutoPaging(params *AgentListParams) *ListIterator[Agent] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[Agent], error) {
		p := &AgentListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// RotateKey rotates an agent's API key and returns the new key.
func (s *AgentsService) RotateKey(ctx context.Context, id string) (*RotateKeyResult, error) {
	result, err := Do[RotateKeyResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/rotate-key", id), map[string]string{"id": id}, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
