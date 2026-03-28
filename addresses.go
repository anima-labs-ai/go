package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Address represents a physical or mailing address for an agent.
type Address struct {
	ID         string `json:"id"`
	AgentID    string `json:"agentId"`
	Type       string `json:"type"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
	Validated  bool   `json:"validated"`
	Primary    bool   `json:"primary"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// CreateAddressParams contains the parameters for creating an address.
type CreateAddressParams struct {
	AgentID    string `json:"agentId"`
	Type       string `json:"type,omitempty"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
	Primary    bool   `json:"primary,omitempty"`
}

// UpdateAddressParams contains the parameters for updating an address.
type UpdateAddressParams struct {
	AgentID    string `json:"agentId,omitempty"`
	Type       string `json:"type,omitempty"`
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	Country    string `json:"country,omitempty"`
	Primary    *bool  `json:"primary,omitempty"`
}

// AddressListParams contains parameters for listing addresses.
type AddressListParams struct {
	ListParams
	AgentID string
	Type    string
}

// ToQuery converts AddressListParams to URL query values.
func (p AddressListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.Type != "" {
		q.Set("type", p.Type)
	}
	return q
}

// AddressList wraps a list of addresses.
type AddressList struct {
	Items []Address `json:"items"`
}

// ValidateAddressResult contains the result of address validation.
type ValidateAddressResult struct {
	Valid            bool     `json:"valid"`
	Deliverable      bool     `json:"deliverable"`
	Corrections      []string `json:"corrections,omitempty"`
	NormalizedAddress *Address `json:"normalizedAddress,omitempty"`
}

// AddressesService provides methods for managing agent physical addresses.
type AddressesService struct {
	client *httpClient
}

// newAddressesService creates a new AddressesService.
func newAddressesService(c *httpClient) *AddressesService {
	return &AddressesService{client: c}
}

// Create creates a new address.
func (s *AddressesService) Create(ctx context.Context, params CreateAddressParams) (*Address, error) {
	addr, err := Do[Address](ctx, s.client, http.MethodPost, "/addresses", params, nil)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// List lists addresses, optionally filtered by agent or type.
func (s *AddressesService) List(ctx context.Context, params *AddressListParams) (*AddressList, error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	list, err := Do[AddressList](ctx, s.client, http.MethodGet, "/addresses", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Get retrieves an address by ID.
func (s *AddressesService) Get(ctx context.Context, id string, agentID string) (*Address, error) {
	q := url.Values{}
	q.Set("agentId", agentID)
	addr, err := Do[Address](ctx, s.client, http.MethodGet, fmt.Sprintf("/addresses/%s", id), nil, q)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// Update updates an existing address.
func (s *AddressesService) Update(ctx context.Context, id string, params UpdateAddressParams) (*Address, error) {
	addr, err := Do[Address](ctx, s.client, http.MethodPut, fmt.Sprintf("/addresses/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

// Delete deletes an address.
func (s *AddressesService) Delete(ctx context.Context, id string, agentID string) error {
	body := struct {
		AgentID string `json:"agentId"`
	}{AgentID: agentID}
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/addresses/%s", id), body, nil)
	return err
}

// Validate validates an address for deliverability.
func (s *AddressesService) Validate(ctx context.Context, id string, agentID string) (*ValidateAddressResult, error) {
	body := struct {
		AgentID string `json:"agentId"`
	}{AgentID: agentID}
	result, err := Do[ValidateAddressResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/addresses/%s/validate", id), body, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
