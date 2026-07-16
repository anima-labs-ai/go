package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Inbox represents an email inbox in the Anima platform.
type Inbox struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Domain      string  `json:"domain"`
	LocalPart   string  `json:"localPart"`
	DisplayName *string `json:"displayName"`
	AgentID     *string `json:"agentId"`
	CreatedAt   string  `json:"createdAt"`
}

// CreateInboxParams contains the parameters for creating an inbox.
// All fields are optional: the server generates a username when omitted and
// falls back to the default domain.
type CreateInboxParams struct {
	// Username is the local part of the inbox email address (letters, numbers,
	// dots, hyphens, underscores). Normalized to lowercase server-side.
	Username string `json:"username,omitempty"`
	// Domain for the inbox email address; the default domain is used if omitted.
	Domain string `json:"domain,omitempty"`
	// DisplayName is a human-readable display name (max 128 characters).
	DisplayName string `json:"displayName,omitempty"`
	// AgentID associates the inbox with an agent.
	AgentID string `json:"agentId,omitempty"`
}

// UpdateInboxParams contains the parameters for updating an inbox.
// Nil fields are left unchanged.
type UpdateInboxParams struct {
	// DisplayName updates the human-readable display name (max 128 characters).
	DisplayName *string `json:"displayName,omitempty"`
	// AgentID updates the agent association.
	AgentID *string `json:"agentId,omitempty"`
}

// InboxListParams contains parameters for listing inboxes.
type InboxListParams struct {
	ListParams
	// Query is a free-text search filtering inboxes by email or display name.
	Query string
}

// ToQuery converts InboxListParams to URL query values.
func (p InboxListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Query != "" {
		q.Set("query", p.Query)
	}
	return q
}

// InboxesService provides methods for managing email inboxes.
type InboxesService struct {
	client *httpClient
}

// newInboxesService creates a new InboxesService.
func newInboxesService(c *httpClient) *InboxesService {
	return &InboxesService{client: c}
}

// Create creates a new inbox. The inbox address can receive mail immediately.
func (s *InboxesService) Create(ctx context.Context, params CreateInboxParams) (*Inbox, error) {
	inbox, err := Do[Inbox](ctx, s.client, http.MethodPost, "/inboxes", params, nil)
	if err != nil {
		return nil, err
	}
	return &inbox, nil
}

// Get retrieves an inbox by ID.
func (s *InboxesService) Get(ctx context.Context, id string) (*Inbox, error) {
	inbox, err := Do[Inbox](ctx, s.client, http.MethodGet, fmt.Sprintf("/inboxes/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &inbox, nil
}

// List returns a paginated list of inboxes.
func (s *InboxesService) List(ctx context.Context, params *InboxListParams) (*Page[Inbox], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[Inbox]](ctx, s.client, http.MethodGet, "/inboxes", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging returns an iterator that automatically paginates through all inboxes.
func (s *InboxesService) ListAutoPaging(params *InboxListParams) *ListIterator[Inbox] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[Inbox], error) {
		p := &InboxListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Update updates an inbox's display name or agent association.
func (s *InboxesService) Update(ctx context.Context, id string, params UpdateInboxParams) (*Inbox, error) {
	inbox, err := Do[Inbox](ctx, s.client, http.MethodPatch, fmt.Sprintf("/inboxes/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &inbox, nil
}

// Delete removes an inbox.
func (s *InboxesService) Delete(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/inboxes/%s", id), nil, nil)
	return err
}
