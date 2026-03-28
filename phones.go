package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ProvisionPhoneParams contains the parameters for provisioning a phone number.
type ProvisionPhoneParams struct {
	AgentID      string   `json:"agentId"`
	CountryCode  string   `json:"countryCode,omitempty"`
	AreaCode     string   `json:"areaCode,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// PhoneConfigUpdateParams contains the parameters for updating a phone's configuration.
type PhoneConfigUpdateParams struct {
	IsPrimary    *bool                  `json:"isPrimary,omitempty"`
	TenDLCStatus TenDLCStatus           `json:"tenDlcStatus,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PhoneListParams contains parameters for listing phone numbers.
type PhoneListParams struct {
	ListParams
	AgentID string
}

// ToQuery converts PhoneListParams to URL query values.
func (p PhoneListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	return q
}

// PhoneList wraps a list of phone identities.
type PhoneList struct {
	Items []PhoneIdentity `json:"items"`
}

// PhonesService provides methods for managing phone numbers.
type PhonesService struct {
	client *httpClient
}

// newPhonesService creates a new PhonesService.
func newPhonesService(c *httpClient) *PhonesService {
	return &PhonesService{client: c}
}

// Provision provisions a new phone number for an agent.
func (s *PhonesService) Provision(ctx context.Context, params ProvisionPhoneParams) (*PhoneIdentity, error) {
	phone, err := Do[PhoneIdentity](ctx, s.client, http.MethodPost, "/phone/provision", params, nil)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// Get retrieves a phone identity by ID.
func (s *PhonesService) Get(ctx context.Context, id string) (*PhoneIdentity, error) {
	phone, err := Do[PhoneIdentity](ctx, s.client, http.MethodGet, fmt.Sprintf("/phones/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// List returns a list of phone numbers. If AgentID is provided, returns numbers
// for that specific agent.
func (s *PhonesService) List(ctx context.Context, params *PhoneListParams) (*PhoneList, error) {
	if params != nil && params.AgentID != "" {
		q := url.Values{}
		q.Set("agentId", params.AgentID)
		list, err := Do[PhoneList](ctx, s.client, http.MethodGet, "/phone/numbers", nil, q)
		if err != nil {
			return nil, err
		}
		return &list, nil
	}

	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	list, err := Do[PhoneList](ctx, s.client, http.MethodGet, "/phones", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Release releases (deletes) a phone number.
func (s *PhonesService) Release(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/phones/%s", id), nil, nil)
	return err
}

// UpdateConfig updates the configuration of a phone number.
func (s *PhonesService) UpdateConfig(ctx context.Context, id string, params PhoneConfigUpdateParams) (*PhoneIdentity, error) {
	phone, err := Do[PhoneIdentity](ctx, s.client, http.MethodPatch, fmt.Sprintf("/phones/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}
