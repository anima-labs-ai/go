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
	CredentialTypeLogin       CredentialType = "login"
	CredentialTypeSecureNote  CredentialType = "secure_note"
	CredentialTypeCard        CredentialType = "card"
	CredentialTypeIdentity    CredentialType = "identity"
	CredentialTypeOAuthToken  CredentialType = "oauth_token"
	CredentialTypeAPIKey      CredentialType = "api_key"
	CredentialTypeCertificate CredentialType = "certificate"
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
	// GeneratePassword asks the vault to generate the login password
	// server-side (login type only, mutually exclusive with
	// Login.Password). Defaults: 24 characters, all character classes.
	// The password is stored in the vault and never returned — the
	// response carries only the masked credential.
	GeneratePassword *GeneratePasswordParams `json:"generatePassword,omitempty"`
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

// VaultIdentityListItem is a provisioned vault plus the name and slug of the
// agent that owns it, as returned by ListIdentities.
type VaultIdentityListItem struct {
	VaultIdentity
	AgentName string `json:"agentName"`
	AgentSlug string `json:"agentSlug"`
}

// ListVaultIdentitiesParams filters the org's vault identities.
type ListVaultIdentitiesParams struct {
	ListParams
	// Status is one of ACTIVE, LOCKED, ERROR. Empty means no filter.
	Status string
}

// ToQuery converts ListVaultIdentitiesParams into URL query values.
func (p ListVaultIdentitiesParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	return q
}

// VaultAuditLogEntry records one access or modification of a credential. It
// never contains secret material.
type VaultAuditLogEntry struct {
	ID           string `json:"id"`
	CredentialID string `json:"credentialId"`
	AgentID      string `json:"agentId"`
	OrgID        string `json:"orgId"`
	// Action is e.g. access, access_reveal, store, delete, refresh, refresh_failed.
	Action string `json:"action"`
	// Actor is a user ID, or "system" for platform-initiated actions.
	Actor     string                 `json:"actor"`
	IPAddress *string                `json:"ipAddress"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt string                 `json:"createdAt"`
}

// VaultAuditParams filters the credential audit trail. Since and Until are
// RFC 3339 timestamps.
type VaultAuditParams struct {
	ListParams
	CredentialID string
	AgentID      string
	Action       string
	Since        string
	Until        string
}

// ToQuery converts VaultAuditParams into URL query values.
func (p VaultAuditParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.CredentialID != "" {
		q.Set("credentialId", p.CredentialID)
	}
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.Action != "" {
		q.Set("action", p.Action)
	}
	if p.Since != "" {
		q.Set("since", p.Since)
	}
	if p.Until != "" {
		q.Set("until", p.Until)
	}
	return q
}

// UseCredentialParams describes an outbound call to make with a credential
// attached server-side.
type UseCredentialParams struct {
	// AgentID is optional when authenticating with an agent API key.
	AgentID string `json:"agentId,omitempty"`
	// Method is one of GET, POST, PUT, PATCH, DELETE, HEAD.
	Method string `json:"method"`
	// URL must be absolute https:// and its host must be on the credential's
	// allowlist.
	URL string `json:"url"`
	// Headers are extra request headers. Any auth header here is dropped and
	// replaced by the credential.
	Headers map[string]string `json:"headers,omitempty"`
	// Body is the raw request body — encode JSON yourself.
	Body string `json:"body,omitempty"`
}

// UseCredentialResult is the upstream response, scrubbed of the injected
// credential.
type UseCredentialResult struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	// Truncated is true when the body exceeded the size cap and was cut off.
	Truncated bool `json:"truncated"`
}

// CredentialRequestStatus is the lifecycle status of a human-in-the-loop
// credential request.
type CredentialRequestStatus string

const (
	// CredentialRequestStatusPending means the request is awaiting the human.
	CredentialRequestStatusPending CredentialRequestStatus = "PENDING"
	// CredentialRequestStatusFulfilled means the secret was submitted and the
	// credential reference is ready.
	CredentialRequestStatusFulfilled CredentialRequestStatus = "FULFILLED"
	CredentialRequestStatusExpired   CredentialRequestStatus = "EXPIRED"
	CredentialRequestStatusDeclined  CredentialRequestStatus = "DECLINED"
	CredentialRequestStatusCancelled CredentialRequestStatus = "CANCELLED"
)

// CreateCredentialRequestParams asks a human for a secret out-of-band. The
// secret is never passed as an argument here.
type CreateCredentialRequestParams struct {
	AgentID string         `json:"agentId,omitempty"`
	Type    CredentialType `json:"type"`
	// Name is the display name for the credential to be created.
	Name string `json:"name"`
	// Reason is shown to the owner to explain why the credential is needed.
	Reason string `json:"reason"`
	// TTLSeconds is 60-3600; zero means the server default of 900 (15 minutes).
	TTLSeconds int `json:"ttlSeconds,omitempty"`
	// NotifyOwner emails the fill URL to the org owner.
	NotifyOwner bool `json:"notifyOwner,omitempty"`
}

// CredentialRequest is a created credential request.
type CredentialRequest struct {
	RequestID string `json:"requestId"`
	// FillURL is the token-gated URL where the human submits the secret.
	FillURL   string                  `json:"fillUrl"`
	Status    CredentialRequestStatus `json:"status"`
	ExpiresAt string                  `json:"expiresAt"`
	EmailSent bool                    `json:"emailSent"`
	// CredentialID is set when the request resolved synchronously; nil
	// otherwise, in which case poll GetCredentialRequestStatus.
	CredentialID *string `json:"credentialId"`
}

// CredentialRequestStatusResult is a polled credential request.
type CredentialRequestStatusResult struct {
	Status CredentialRequestStatus `json:"status"`
	// CredentialID is set once the request is FULFILLED, nil otherwise.
	CredentialID *string `json:"credentialId"`
	// MaskedPreview (e.g. ****1234) confirms which secret arrived without
	// revealing it. Nil until FULFILLED.
	MaskedPreview *string `json:"maskedPreview"`
}

// CancelCredentialRequestResult is the outcome of cancelling a request.
type CancelCredentialRequestResult struct {
	Status CredentialRequestStatus `json:"status"`
}

// SharePermission represents the permission level for a shared credential.
type SharePermission string

const (
	SharePermissionRead   SharePermission = "READ"
	SharePermissionUse    SharePermission = "USE"
	SharePermissionManage SharePermission = "MANAGE"
)

// TokenScope represents the scope of an ephemeral vault token.
type TokenScope string

const (
	TokenScopeAutofill TokenScope = "autofill"
	TokenScopeProxy    TokenScope = "proxy"
	TokenScopeExport   TokenScope = "export"
)

// VaultShare represents a credential share between agents.
type VaultShare struct {
	ID            string  `json:"id"`
	CredentialID  string  `json:"credentialId"`
	SourceAgentID string  `json:"sourceAgentId"`
	TargetAgentID string  `json:"targetAgentId"`
	Permission    string  `json:"permission"`
	ExpiresAt     *string `json:"expiresAt"`
	CreatedAt     string  `json:"createdAt"`
}

// VaultShareList wraps a list of vault shares.
type VaultShareList struct {
	Items []VaultShare `json:"items"`
}

// ShareCredentialParams contains parameters for sharing a credential.
type ShareCredentialParams struct {
	AgentID          string `json:"agentId"`
	CredentialID     string `json:"credentialId"`
	TargetAgentID    string `json:"targetAgentId"`
	Permission       string `json:"permission"`
	ExpiresInSeconds *int   `json:"expiresInSeconds,omitempty"`
}

// RevokeShareParams contains parameters for revoking a share.
type RevokeShareParams struct {
	ShareID string `json:"shareId"`
	AgentID string `json:"agentId,omitempty"`
}

// VaultToken represents an ephemeral vault token.
type VaultToken struct {
	Token        string `json:"token"`
	CredentialID string `json:"credentialId"`
	Scope        string `json:"scope"`
	ExpiresAt    string `json:"expiresAt"`
}

// CreateTokenParams contains parameters for creating an ephemeral token.
type CreateTokenParams struct {
	CredentialID string `json:"credentialId"`
	Scope        string `json:"scope"`
	AgentID      string `json:"agentId,omitempty"`
	TTLSeconds   *int   `json:"ttlSeconds,omitempty"`
}

// RevokeTokensParams contains parameters for revoking tokens.
type RevokeTokensParams struct {
	CredentialID string `json:"credentialId"`
	AgentID      string `json:"agentId,omitempty"`
}

// RevokeTokensResult contains the result of revoking tokens.
type RevokeTokensResult struct {
	Success bool `json:"success"`
	Revoked int  `json:"revoked"`
}

// VaultService provides methods for managing the agent credential vault.
type VaultService struct {
	client *httpClient
}

// newVaultService creates a new VaultService.
func newVaultService(c *httpClient) *VaultService {
	return &VaultService{client: c}
}

// ListIdentities lists the provisioned vaults across the organization.
func (s *VaultService) ListIdentities(ctx context.Context, params *ListVaultIdentitiesParams) (*Page[VaultIdentityListItem], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[VaultIdentityListItem]](ctx, s.client, http.MethodGet, "/vault/identities", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListIdentitiesAutoPaging iterates every vault identity, fetching pages as needed.
func (s *VaultService) ListIdentitiesAutoPaging(params *ListVaultIdentitiesParams) *ListIterator[VaultIdentityListItem] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[VaultIdentityListItem], error) {
		p := &ListVaultIdentitiesParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListIdentities(ctx, p)
	})
}

// Audit queries the credential access and modification trail.
func (s *VaultService) Audit(ctx context.Context, params *VaultAuditParams) (*Page[VaultAuditLogEntry], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[VaultAuditLogEntry]](ctx, s.client, http.MethodGet, "/vault/audit", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// AuditAutoPaging iterates every audit entry, fetching pages as needed.
func (s *VaultService) AuditAutoPaging(params *VaultAuditParams) *ListIterator[VaultAuditLogEntry] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[VaultAuditLogEntry], error) {
		p := &VaultAuditParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.Audit(ctx, p)
	})
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

// UseCredential makes an outbound HTTPS call with the credential attached
// server-side and returns the upstream response. The plaintext secret is never
// returned — this is how a brokered credential is used without revealing it.
func (s *VaultService) UseCredential(ctx context.Context, id string, params UseCredentialParams) (*UseCredentialResult, error) {
	result, err := Do[UseCredentialResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/vault/credentials/%s/use", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCredentialRequest asks a human to supply a secret out-of-band. Hand the
// returned FillURL to the human; the secret never passes through the agent.
//
// Named credentialRequestCreate in the node and python SDKs.
func (s *VaultService) CreateCredentialRequest(ctx context.Context, params CreateCredentialRequestParams) (*CredentialRequest, error) {
	req, err := Do[CredentialRequest](ctx, s.client, http.MethodPost, "/vault/credential-requests", params, nil)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// GetCredentialRequestStatus polls a credential request until it is FULFILLED.
//
// Named credentialRequestStatus in the node and python SDKs.
func (s *VaultService) GetCredentialRequestStatus(ctx context.Context, requestID string) (*CredentialRequestStatusResult, error) {
	result, err := Do[CredentialRequestStatusResult](ctx, s.client, http.MethodGet, fmt.Sprintf("/vault/credential-requests/%s", requestID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelCredentialRequest invalidates a pending request and its fill URL.
//
// Named credentialRequestCancel in the node and python SDKs.
func (s *VaultService) CancelCredentialRequest(ctx context.Context, requestID string) (*CancelCredentialRequestResult, error) {
	result, err := Do[CancelCredentialRequestResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/vault/credential-requests/%s/cancel", requestID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
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

// --- Sharing ---

// ShareCredential shares a vault credential with another agent.
func (s *VaultService) ShareCredential(ctx context.Context, params ShareCredentialParams) (*VaultShare, error) {
	share, err := Do[VaultShare](ctx, s.client, http.MethodPost, "/vault/share", params, nil)
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// ListShares lists credential shares granted by or received by an agent.
func (s *VaultService) ListShares(ctx context.Context, direction string, agentID string) (*VaultShareList, error) {
	q := url.Values{}
	q.Set("direction", direction)
	if agentID != "" {
		q.Set("agentId", agentID)
	}
	list, err := Do[VaultShareList](ctx, s.client, http.MethodGet, "/vault/shares", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// RevokeShare revokes a previously granted credential share.
func (s *VaultService) RevokeShare(ctx context.Context, params RevokeShareParams) error {
	_, err := Do[SuccessResult](ctx, s.client, http.MethodPost, "/vault/share/revoke", params, nil)
	return err
}

// --- Ephemeral Tokens ---

// CreateToken creates a short-lived ephemeral token for a credential.
func (s *VaultService) CreateToken(ctx context.Context, params CreateTokenParams) (*VaultToken, error) {
	token, err := Do[VaultToken](ctx, s.client, http.MethodPost, "/vault/token", params, nil)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// ExchangeToken exchanges an ephemeral vtk_ token for the underlying credential.
func (s *VaultService) ExchangeToken(ctx context.Context, token string) (*VaultCredential, error) {
	body := struct {
		Token string `json:"token"`
	}{Token: token}
	cred, err := Do[VaultCredential](ctx, s.client, http.MethodPost, "/vault/token/exchange", body, nil)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// RevokeTokens revokes all active ephemeral tokens for a credential.
func (s *VaultService) RevokeTokens(ctx context.Context, params RevokeTokensParams) (*RevokeTokensResult, error) {
	result, err := Do[RevokeTokensResult](ctx, s.client, http.MethodPost, "/vault/token/revoke", params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
