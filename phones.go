package anima

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// SearchPhonesParams contains parameters for searching available phone numbers.
type SearchPhonesParams struct {
	CountryCode  string   `json:"countryCode,omitempty"`
	AreaCode     string   `json:"areaCode,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Limit        int      `json:"limit,omitempty"`
}

// AvailableNumber represents a phone number available for provisioning.
type AvailableNumber struct {
	PhoneNumber  string           `json:"phoneNumber"`
	Region       string           `json:"region,omitempty"`
	Capabilities PhoneCapabilities `json:"capabilities"`
	MonthlyCost  float64          `json:"monthlyCost,omitempty"`
}

// AvailableNumberList wraps a list of available phone numbers.
type AvailableNumberList struct {
	Items []AvailableNumber `json:"items"`
}

// ProvisionPhoneParams contains the parameters for provisioning a phone number.
type ProvisionPhoneParams struct {
	AgentID      string   `json:"agentId"`
	CountryCode  string   `json:"countryCode,omitempty"`
	AreaCode     string   `json:"areaCode,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// ReleasePhoneParams contains the parameters for releasing a phone number.
type ReleasePhoneParams struct {
	AgentID     string `json:"agentId"`
	PhoneNumber string `json:"phoneNumber"`
}

// PhoneListParams contains parameters for listing phone numbers.
type PhoneListParams struct {
	AgentID string
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

// Search searches for available phone numbers.
func (s *PhonesService) Search(ctx context.Context, params SearchPhonesParams) (*AvailableNumberList, error) {
	q := url.Values{}
	if params.CountryCode != "" {
		q.Set("countryCode", params.CountryCode)
	}
	if params.AreaCode != "" {
		q.Set("areaCode", params.AreaCode)
	}
	for _, cap := range params.Capabilities {
		q.Add("capabilities[]", cap)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	list, err := Do[AvailableNumberList](ctx, s.client, http.MethodGet, "/phone/search", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Provision provisions a new phone number for an agent.
func (s *PhonesService) Provision(ctx context.Context, params ProvisionPhoneParams) (*PhoneIdentity, error) {
	phone, err := Do[PhoneIdentity](ctx, s.client, http.MethodPost, "/phone/provision", params, nil)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// List returns a list of phone numbers for the specified agent.
func (s *PhonesService) List(ctx context.Context, params PhoneListParams) (*PhoneList, error) {
	q := url.Values{}
	q.Set("agentId", params.AgentID)
	list, err := Do[PhoneList](ctx, s.client, http.MethodGet, "/phone/numbers", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Release releases a phone number from an agent.
func (s *PhonesService) Release(ctx context.Context, params ReleasePhoneParams) error {
	_, err := Do[struct{ Success bool }](ctx, s.client, http.MethodPost, "/phone/release", params, nil)
	return err
}

