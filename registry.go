package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// RegistryAgent represents an agent registered in the public agent registry.
type RegistryAgent struct {
	DID          string   `json:"did"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Category     string   `json:"category,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Endpoints    any      `json:"endpoints,omitempty"`
	Verified     bool     `json:"verified"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}

// RegisterAgentParams contains the parameters for registering an agent in the registry.
type RegisterAgentParams struct {
	DID          string   `json:"did"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Category     string   `json:"category,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Endpoints    any      `json:"endpoints,omitempty"`
}

// UpdateRegistryAgentParams contains the parameters for updating a registry entry.
type UpdateRegistryAgentParams struct {
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	Category     string   `json:"category,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Endpoints    any      `json:"endpoints,omitempty"`
}

// RegistrySearchParams contains parameters for searching the agent registry.
type RegistrySearchParams struct {
	Query    string
	Category string
	Cursor   string
	Limit    int
}

// RegistryAgentList wraps a list of registry agents.
type RegistryAgentList struct {
	Items []RegistryAgent `json:"items"`
}

// RegistryService provides methods for the public agent registry.
type RegistryService struct {
	client *httpClient
}

// newRegistryService creates a new RegistryService.
func newRegistryService(c *httpClient) *RegistryService {
	return &RegistryService{client: c}
}

// Register registers an agent in the public registry.
func (s *RegistryService) Register(ctx context.Context, params RegisterAgentParams) (*RegistryAgent, error) {
	agent, err := Do[RegistryAgent](ctx, s.client, http.MethodPost, "/registry/agents", params, nil)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Search searches the agent registry.
func (s *RegistryService) Search(ctx context.Context, params RegistrySearchParams) (*RegistryAgentList, error) {
	q := url.Values{}
	if params.Query != "" {
		q.Set("q", params.Query)
	}
	if params.Category != "" {
		q.Set("category", params.Category)
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	list, err := Do[RegistryAgentList](ctx, s.client, http.MethodGet, "/registry/agents/search", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Lookup retrieves a registry entry by DID.
func (s *RegistryService) Lookup(ctx context.Context, did string) (*RegistryAgent, error) {
	agent, err := Do[RegistryAgent](ctx, s.client, http.MethodGet, fmt.Sprintf("/registry/agents/%s", did), nil, nil)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Update updates an existing registry entry.
func (s *RegistryService) Update(ctx context.Context, did string, params UpdateRegistryAgentParams) (*RegistryAgent, error) {
	agent, err := Do[RegistryAgent](ctx, s.client, http.MethodPut, fmt.Sprintf("/registry/agents/%s", did), params, nil)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Unlist removes an agent from the public registry.
func (s *RegistryService) Unlist(ctx context.Context, did string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/registry/agents/%s", did), nil, nil)
	return err
}
