package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// OAuthApp represents an OAuth app definition.
type OAuthApp struct {
	ID            string   `json:"id"`
	Slug          string   `json:"slug"`
	Name          string   `json:"name"`
	Description   *string  `json:"description"`
	IconURL       *string  `json:"iconUrl"`
	AuthMethod    string   `json:"authMethod"`
	DefaultScopes []string `json:"defaultScopes"`
	RequiresPKCE  bool     `json:"requiresPkce"`
	Category      *string  `json:"category"`
	IsManaged     bool     `json:"isManaged"`
	IsActive      bool     `json:"isActive"`
}

// OAuthAppList wraps a list of OAuth apps.
type OAuthAppList struct {
	Items []OAuthApp `json:"items"`
}

// ConnectedAccount represents an agent's authenticated connection to a service.
type ConnectedAccount struct {
	ID              string   `json:"id"`
	AgentID         string   `json:"agentId"`
	UserID          *string  `json:"userId"`
	AppDefinitionID string   `json:"appDefinitionId"`
	AppSlug         string   `json:"appSlug"`
	AppName         string   `json:"appName"`
	AppIconURL      *string  `json:"appIconUrl"`
	CustomAppID     *string  `json:"customAppId"`
	GrantedScopes   []string `json:"grantedScopes"`
	AccountLabel    *string  `json:"accountLabel"`
	AccountEmail    *string  `json:"accountEmail"`
	Status          string   `json:"status"`
	StatusMessage   *string  `json:"statusMessage"`
	TokenExpiresAt  *string  `json:"tokenExpiresAt"`
	LastRefreshedAt *string  `json:"lastRefreshedAt"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}

// ConnectedAccountList wraps a list of connected accounts.
type ConnectedAccountList struct {
	Items []ConnectedAccount `json:"items"`
}

// ConnectLinkResult is the response from creating a Connect Link.
type ConnectLinkResult struct {
	LinkURL   string `json:"linkUrl"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

// ConnectLinkStatus is the response from checking a Connect Link's status.
type ConnectLinkStatus struct {
	Status             string  `json:"status"`
	ConnectedAccountID *string `json:"connectedAccountId"`
}

// CreateConnectLinkParams contains parameters for creating a Connect Link.
type CreateConnectLinkParams struct {
	AgentID     string   `json:"agentId,omitempty"`
	AppSlug     string   `json:"appSlug"`
	UserID      string   `json:"userId,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	CallbackURL string   `json:"callbackUrl,omitempty"`
	CustomAppID string   `json:"customAppId,omitempty"`
}

// ListConnectedAccountsParams contains parameters for listing connected accounts.
type ListConnectedAccountsParams struct {
	AgentID string
	UserID  string
	AppSlug string
	Status  string
}

// VaultOAuthService provides methods for managing OAuth connections.
type VaultOAuthService struct {
	client *httpClient
}

func newVaultOAuthService(c *httpClient) *VaultOAuthService {
	return &VaultOAuthService{client: c}
}

// ListApps lists available OAuth app definitions.
func (s *VaultOAuthService) ListApps(ctx context.Context, category string) (*OAuthAppList, error) {
	q := url.Values{}
	if category != "" {
		q.Set("category", category)
	}
	list, err := Do[OAuthAppList](ctx, s.client, http.MethodGet, "/vault/oauth/apps", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// GetApp gets a single OAuth app definition by slug.
func (s *VaultOAuthService) GetApp(ctx context.Context, slug string) (*OAuthApp, error) {
	app, err := Do[OAuthApp](ctx, s.client, http.MethodGet, fmt.Sprintf("/vault/oauth/apps/%s", slug), nil, nil)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// CreateLink creates a Connect Link for zero-code authentication.
func (s *VaultOAuthService) CreateLink(ctx context.Context, params CreateConnectLinkParams) (*ConnectLinkResult, error) {
	result, err := Do[ConnectLinkResult](ctx, s.client, http.MethodPost, "/vault/oauth/link", params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLinkStatus checks the status of a Connect Link.
func (s *VaultOAuthService) GetLinkStatus(ctx context.Context, token string) (*ConnectLinkStatus, error) {
	result, err := Do[ConnectLinkStatus](ctx, s.client, http.MethodGet, fmt.Sprintf("/vault/oauth/link/%s", token), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListAccounts lists connected OAuth accounts.
func (s *VaultOAuthService) ListAccounts(ctx context.Context, params ListConnectedAccountsParams) (*ConnectedAccountList, error) {
	q := url.Values{}
	if params.AgentID != "" {
		q.Set("agentId", params.AgentID)
	}
	if params.UserID != "" {
		q.Set("userId", params.UserID)
	}
	if params.AppSlug != "" {
		q.Set("appSlug", params.AppSlug)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	list, err := Do[ConnectedAccountList](ctx, s.client, http.MethodGet, "/vault/oauth/accounts", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Disconnect disconnects an OAuth account.
func (s *VaultOAuthService) Disconnect(ctx context.Context, accountID string) error {
	_, err := Do[SuccessResult](ctx, s.client, http.MethodDelete, fmt.Sprintf("/vault/oauth/accounts/%s", accountID), nil, nil)
	return err
}
