package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// EmailDraft represents a composed-but-not-sent email owned by an agent.
//
// Distinct from Message: drafts can be incomplete (no recipients, no
// subject), carry no threadId/status/delivery state, and the Send
// operation converts a draft into a real Message and deletes the draft
// row atomically.
type EmailDraft struct {
	ID      string `json:"id"`
	AgentID string `json:"agentId"`
	OrgID   string `json:"orgId"`
	// FromIdentityID is the EmailIdentity used as the sender, or nil to use
	// the agent's primary identity at send time.
	FromIdentityID *string  `json:"fromIdentityId"`
	To             []string `json:"to"`
	CC             []string `json:"cc"`
	BCC            []string `json:"bcc"`
	// Subject is nil if not yet written.
	Subject *string `json:"subject"`
	// Body is the plain-text body, nil if not yet written.
	Body *string `json:"body"`
	// BodyHTML is the HTML body, nil if not provided.
	BodyHTML *string `json:"bodyHtml"`
	// InReplyTo is an optional In-Reply-To Message-ID applied for threading
	// on send.
	InReplyTo *string `json:"inReplyTo"`
	// References is an optional References Message-ID chain applied for
	// threading on send.
	References []string               `json:"references"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

// CreateDraftParams contains the parameters for creating an email draft.
// Only AgentID is required — drafts may be created incomplete and sent
// once they have at least one recipient, a subject, and a body.
type CreateDraftParams struct {
	AgentID string `json:"agentId"`
	// FromIdentityID optionally selects the EmailIdentity to send from. It
	// must belong to this agent and be verified.
	FromIdentityID string   `json:"fromIdentityId,omitempty"`
	To             []string `json:"to,omitempty"`
	CC             []string `json:"cc,omitempty"`
	BCC            []string `json:"bcc,omitempty"`
	Subject        string   `json:"subject,omitempty"`
	Body           string   `json:"body,omitempty"`
	BodyHTML       string   `json:"bodyHtml,omitempty"`
	// InReplyTo is the Message-ID of the email the draft replies to, used
	// for threading when the draft is sent.
	InReplyTo string `json:"inReplyTo,omitempty"`
	// References is the list of Message-IDs forming the email thread chain.
	References []string               `json:"references,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DraftListParams contains parameters for listing email drafts.
type DraftListParams struct {
	ListParams
	// AgentID filters drafts to a single agent. Agent-scoped keys are always
	// limited to their own drafts regardless of this filter.
	AgentID string
}

// ToQuery converts DraftListParams to URL query values.
func (p DraftListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	return q
}

// DraftsService provides methods for managing email drafts.
type DraftsService struct {
	client *httpClient
}

// newDraftsService creates a new DraftsService.
func newDraftsService(c *httpClient) *DraftsService {
	return &DraftsService{client: c}
}

// Create creates a new email draft.
func (s *DraftsService) Create(ctx context.Context, params CreateDraftParams) (*EmailDraft, error) {
	d, err := Do[EmailDraft](ctx, s.client, http.MethodPost, "/email/drafts", params, nil)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Get retrieves an email draft by ID.
func (s *DraftsService) Get(ctx context.Context, id string) (*EmailDraft, error) {
	d, err := Do[EmailDraft](ctx, s.client, http.MethodGet, fmt.Sprintf("/email/drafts/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// List returns a paginated list of email drafts, newest first.
func (s *DraftsService) List(ctx context.Context, params *DraftListParams) (*Page[EmailDraft], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[EmailDraft]](ctx, s.client, http.MethodGet, "/email/drafts", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging returns an iterator that automatically paginates through all drafts.
func (s *DraftsService) ListAutoPaging(params *DraftListParams) *ListIterator[EmailDraft] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[EmailDraft], error) {
		p := &DraftListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Send sends the draft, converting it into a real Message with full
// email-send semantics (threading, scanning, plan limits) and deleting the
// draft row atomically. It returns the new Message, not the draft; the
// draft ID 404s afterwards.
//
// The draft must have at least one recipient, a subject, and a body —
// sending an incomplete draft fails with a validation error and the draft
// survives so it can be completed and retried.
func (s *DraftsService) Send(ctx context.Context, id string) (*Message, error) {
	body := struct {
		ID string `json:"id"`
	}{ID: id}
	msg, err := Do[Message](ctx, s.client, http.MethodPost, fmt.Sprintf("/email/drafts/%s/send", id), body, nil)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// Delete deletes the draft without sending it. It returns the deleted draft.
func (s *DraftsService) Delete(ctx context.Context, id string) (*EmailDraft, error) {
	d, err := Do[EmailDraft](ctx, s.client, http.MethodDelete, fmt.Sprintf("/email/drafts/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
