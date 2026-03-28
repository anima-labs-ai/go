package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Tier represents the subscription tier of an organization.
type Tier string

const (
	TierFree       Tier = "FREE"
	TierDeveloper  Tier = "DEVELOPER"
	TierGrowth     Tier = "GROWTH"
	TierScale      Tier = "SCALE"
	TierEnterprise Tier = "ENTERPRISE"
)

// Organization represents an Anima organization.
type Organization struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Slug       string                 `json:"slug"`
	ClerkOrgID *string                `json:"clerkOrgId"`
	Tier       Tier                   `json:"tier"`
	MasterKey  string                 `json:"masterKey"`
	Settings   map[string]interface{} `json:"settings"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

// CreateOrganizationParams contains the parameters for creating an organization.
type CreateOrganizationParams struct {
	Name       string                 `json:"name"`
	Slug       string                 `json:"slug"`
	ClerkOrgID string                 `json:"clerkOrgId,omitempty"`
	Tier       Tier                   `json:"tier,omitempty"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// UpdateOrganizationParams contains the parameters for updating an organization.
type UpdateOrganizationParams struct {
	Name       string                 `json:"name,omitempty"`
	Slug       string                 `json:"slug,omitempty"`
	ClerkOrgID *string                `json:"clerkOrgId,omitempty"`
	Tier       Tier                   `json:"tier,omitempty"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// OrganizationListParams contains parameters for listing organizations.
type OrganizationListParams struct {
	ListParams
	Query string
}

// ToQuery converts OrganizationListParams to URL query values.
func (p OrganizationListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Query != "" {
		q.Set("query", p.Query)
	}
	return q
}

// RotateMasterKeyResult contains the result of a master key rotation.
type RotateMasterKeyResult struct {
	MasterKey string `json:"masterKey"`
}

// OrganizationsService provides methods for managing organizations.
type OrganizationsService struct {
	client *httpClient
}

// newOrganizationsService creates a new OrganizationsService.
func newOrganizationsService(c *httpClient) *OrganizationsService {
	return &OrganizationsService{client: c}
}

// Create creates a new organization.
func (s *OrganizationsService) Create(ctx context.Context, params CreateOrganizationParams) (*Organization, error) {
	org, err := Do[Organization](ctx, s.client, http.MethodPost, "/orgs", params, nil)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// Get retrieves an organization by ID.
func (s *OrganizationsService) Get(ctx context.Context, id string) (*Organization, error) {
	org, err := Do[Organization](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// Update updates an existing organization.
func (s *OrganizationsService) Update(ctx context.Context, id string, params UpdateOrganizationParams) (*Organization, error) {
	org, err := Do[Organization](ctx, s.client, http.MethodPatch, fmt.Sprintf("/orgs/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// Delete deletes an organization.
func (s *OrganizationsService) Delete(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/orgs/%s", id), nil, nil)
	return err
}

// List returns a paginated list of organizations.
func (s *OrganizationsService) List(ctx context.Context, params *OrganizationListParams) (*Page[Organization], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[Organization]](ctx, s.client, http.MethodGet, "/orgs", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging returns an iterator that automatically paginates through all organizations.
func (s *OrganizationsService) ListAutoPaging(params *OrganizationListParams) *ListIterator[Organization] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[Organization], error) {
		p := &OrganizationListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// RotateKey rotates an organization's master key.
func (s *OrganizationsService) RotateKey(ctx context.Context, id string) (*RotateMasterKeyResult, error) {
	result, err := Do[RotateMasterKeyResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/orgs/%s/rotate-key", id), map[string]string{"id": id}, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
