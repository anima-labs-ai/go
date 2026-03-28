package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CredentialType represents the type of a vault credential.
type CredentialType string

const (
	CredentialTypeLogin      CredentialType = "login"
	CredentialTypeSecureNote CredentialType = "secure_note"
	CredentialTypeCard       CredentialType = "card"
	CredentialTypeIdentity   CredentialType = "identity"
)

// VaultIdentity represents a vault provisioned for an agent.
type VaultIdentity struct {
	ID              string  `json:"id"`
	AgentID         string  `json:"agentId"`
	OrgID           string  `json:"orgId"`
	Status          string  `json:"status"`
	CredentialCount int     `json:"credentialCount"`
	LastSyncAt      *string `json:"lastSyncAt"`
	CreatedAt       string  `json:"createdAt"`
}

// VaultLoginData contains login credential data.
type VaultLoginData struct {
	Username string      `json:"username,omitempty"`
	Password string      `json:"password,omitempty"`
	URIs     []VaultURI  `json:"uris,omitempty"`
	TOTP     string      `json:"totp,omitempty"`
}

// VaultURI represents a URI associated with a login credential.
type VaultURI struct {
	URI   string `json:"uri"`
	Match string `json:"match,omitempty"`
}

// VaultCardData contains card credential data.
type VaultCardData struct {
	CardholderName string `json:"cardholderName,omitempty"`
	Brand          string `json:"brand,omitempty"`
	Number         string `json:"number,omitempty"`
	ExpMonth       string `json:"expMonth,omitempty"`
	ExpYear        string `json:"expYear,omitempty"`
	Code           string `json:"code,omitempty"`
}

// VaultIdentityData contains identity credential data.
type VaultIdentityData struct {
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Address1   string `json:"address1,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	Country    string `json:"country,omitempty"`
	Company    string `json:"company,omitempty"`
}

// VaultCustomField represents a custom field on a vault credential.
type VaultCustomField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"` // "text", "hidden", "boolean"
}

// VaultCredential represents a credential stored in the vault.
type VaultCredential struct {
	ID        string             `json:"id"`
	Type      CredentialType     `json:"type"`
	Name      string             `json:"name"`
	Notes     string             `json:"notes,omitempty"`
	Login     *VaultLoginData    `json:"login,omitempty"`
	Card      *VaultCardData     `json:"card,omitempty"`
	Identity  *VaultIdentityData `json:"identity,omitempty"`
	Fields    []VaultCustomField `json:"fields,omitempty"`
	Favorite  bool               `json:"favorite"`
	CreatedAt string             `json:"createdAt"`
	UpdatedAt string             `json:"updatedAt"`
}

// CreateVaultCredentialParams contains the parameters for creating a credential.
type CreateVaultCredentialParams struct {
	AgentID  string             `json:"agentId"`
	Type     CredentialType     `json:"type"`
	Name     string             `json:"name"`
	Notes    string             `json:"notes,omitempty"`
	Login    *VaultLoginData    `json:"login,omitempty"`
	Card     *VaultCardData     `json:"card,omitempty"`
	Identity *VaultIdentityData `json:"identity,omitempty"`
	Fields   []VaultCustomField `json:"fields,omitempty"`
	Favorite bool               `json:"favorite,omitempty"`
}

// UpdateVaultCredentialParams contains the parameters for updating a credential.
type UpdateVaultCredentialParams struct {
	Name     string             `json:"name,omitempty"`
	Notes    string             `json:"notes,omitempty"`
	Login    *VaultLoginData    `json:"login,omitempty"`
	Card     *VaultCardData     `json:"card,omitempty"`
	Identity *VaultIdentityData `json:"identity,omitempty"`
	Fields   []VaultCustomField `json:"fields,omitempty"`
	Favorite *bool              `json:"favorite,omitempty"`
}

// ListVaultCredentialsParams contains parameters for listing vault credentials.
type ListVaultCredentialsParams struct {
	AgentID string
	Type    CredentialType
}

// SearchVaultParams contains parameters for searching the vault.
type SearchVaultParams struct {
	AgentID string
	Search  string
	Type    CredentialType
}

// GeneratePasswordParams contains parameters for generating a password.
type GeneratePasswordParams struct {
	Length    int  `json:"length,omitempty"`
	Uppercase *bool `json:"uppercase,omitempty"`
	Lowercase *bool `json:"lowercase,omitempty"`
	Number   *bool `json:"number,omitempty"`
	Special  *bool `json:"special,omitempty"`
}

// GeneratePasswordResult contains the generated password.
type GeneratePasswordResult struct {
	Password string `json:"password"`
}

// VaultTOTP contains a TOTP code and its period.
type VaultTOTP struct {
	Code   string `json:"code"`
	Period int    `json:"period"`
}

// VaultStatus contains the vault sync status.
type VaultStatus struct {
	ServerURL string  `json:"serverUrl"`
	LastSync  *string `json:"lastSync"`
	Status    string  `json:"status"`
}

// SuccessResult is a generic success response.
type SuccessResult struct {
	Success bool `json:"success"`
}

// VaultCredentialList wraps a list of vault credentials.
type VaultCredentialList struct {
	Items []VaultCredential `json:"items"`
}

// VaultService provides methods for managing the agent credential vault.
type VaultService struct {
	client *httpClient
}

// newVaultService creates a new VaultService.
func newVaultService(c *httpClient) *VaultService {
	return &VaultService{client: c}
}

// Provision provisions a vault for an agent.
func (s *VaultService) Provision(ctx context.Context, agentID string) (*VaultIdentity, error) {
	body := struct {
		AgentID string `json:"agentId"`
	}{AgentID: agentID}
	vi, err := Do[VaultIdentity](ctx, s.client, http.MethodPost, "/vault/provision", body, nil)
	if err != nil {
		return nil, err
	}
	return &vi, nil
}

// Deprovision removes a vault from an agent.
func (s *VaultService) Deprovision(ctx context.Context, agentID string) error {
	body := struct {
		AgentID string `json:"agentId"`
	}{AgentID: agentID}
	_, err := Do[SuccessResult](ctx, s.client, http.MethodPost, "/vault/deprovision", body, nil)
	return err
}

// ListCredentials lists all credentials for an agent.
func (s *VaultService) ListCredentials(ctx context.Context, params ListVaultCredentialsParams) (*VaultCredentialList, error) {
	q := url.Values{}
	q.Set("agentId", params.AgentID)
	if params.Type != "" {
		q.Set("type", string(params.Type))
	}
	list, err := Do[VaultCredentialList](ctx, s.client, http.MethodGet, "/vault/credentials", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// GetCredential retrieves a credential by ID.
func (s *VaultService) GetCredential(ctx context.Context, id string) (*VaultCredential, error) {
	cred, err := Do[VaultCredential](ctx, s.client, http.MethodGet, fmt.Sprintf("/vault/credentials/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// CreateCredential creates a new credential in the vault.
func (s *VaultService) CreateCredential(ctx context.Context, params CreateVaultCredentialParams) (*VaultCredential, error) {
	cred, err := Do[VaultCredential](ctx, s.client, http.MethodPost, "/vault/credentials", params, nil)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// UpdateCredential updates an existing credential.
func (s *VaultService) UpdateCredential(ctx context.Context, id string, params UpdateVaultCredentialParams) (*VaultCredential, error) {
	cred, err := Do[VaultCredential](ctx, s.client, http.MethodPut, fmt.Sprintf("/vault/credentials/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// DeleteCredential deletes a credential from the vault.
func (s *VaultService) DeleteCredential(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/vault/credentials/%s", id), nil, nil)
	return err
}

// Search searches credentials in the vault.
func (s *VaultService) Search(ctx context.Context, params SearchVaultParams) (*VaultCredentialList, error) {
	q := url.Values{}
	q.Set("agentId", params.AgentID)
	q.Set("search", params.Search)
	if params.Type != "" {
		q.Set("type", string(params.Type))
	}
	list, err := Do[VaultCredentialList](ctx, s.client, http.MethodGet, "/vault/search", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// GeneratePassword generates a random password.
func (s *VaultService) GeneratePassword(ctx context.Context, params *GeneratePasswordParams) (*GeneratePasswordResult, error) {
	result, err := Do[GeneratePasswordResult](ctx, s.client, http.MethodPost, "/vault/generate-password", params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTOTP retrieves the current TOTP code for a credential.
func (s *VaultService) GetTOTP(ctx context.Context, credentialID string) (*VaultTOTP, error) {
	totp, err := Do[VaultTOTP](ctx, s.client, http.MethodGet, fmt.Sprintf("/vault/totp/%s", credentialID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &totp, nil
}

// Status retrieves the vault sync status for an agent.
func (s *VaultService) Status(ctx context.Context, agentID string) (*VaultStatus, error) {
	q := url.Values{}
	q.Set("agentId", agentID)
	status, err := Do[VaultStatus](ctx, s.client, http.MethodGet, "/vault/status", nil, q)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Sync triggers a vault sync for an agent.
func (s *VaultService) Sync(ctx context.Context, agentID string) error {
	body := struct {
		AgentID string `json:"agentId"`
	}{AgentID: agentID}
	_, err := Do[SuccessResult](ctx, s.client, http.MethodPost, "/vault/sync", body, nil)
	return err
}
