package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ProvisionableResource is what an agent may ask its owner to provision.
//
// A closed set on purpose: these are the two things an agent can be given that
// cost money or grant reach.
type ProvisionableResource string

const (
	ProvisionableResourceVault       ProvisionableResource = "VAULT"
	ProvisionableResourcePhoneNumber ProvisionableResource = "PHONE_NUMBER"
)

// ProvisioningRequestStatus is the lifecycle state of a provisioning request.
type ProvisioningRequestStatus string

const (
	ProvisioningRequestPending   ProvisioningRequestStatus = "PENDING"
	ProvisioningRequestApproved  ProvisioningRequestStatus = "APPROVED"
	ProvisioningRequestDeclined  ProvisioningRequestStatus = "DECLINED"
	ProvisioningRequestExpired   ProvisioningRequestStatus = "EXPIRED"
	ProvisioningRequestCancelled ProvisioningRequestStatus = "CANCELLED"
)

// ProvisioningOptions carries per-resource options. Only PHONE_NUMBER takes any.
type ProvisioningOptions struct {
	CountryCode string `json:"countryCode,omitempty"`
	AreaCode    string `json:"areaCode,omitempty"`
}

// ProvisioningRequest is an agent's ask and its current state.
type ProvisioningRequest struct {
	RequestID string `json:"requestId"`
	AgentID   string `json:"agentId"`
	// AgentName is present so the owner knows who is asking, not just an id.
	AgentName string                `json:"agentName"`
	Resource  ProvisionableResource `json:"resource"`
	Reason    string                `json:"reason"`
	// Status is lazily expired: a request past its TTL reads as EXPIRED here
	// even though nothing wrote that transition.
	Status    ProvisioningRequestStatus `json:"status"`
	Options   *ProvisioningOptions      `json:"options"`
	ExpiresAt string                    `json:"expiresAt"`
	DecidedAt *string                   `json:"decidedAt"`
	// DecidedNote is the owner's note, typically why it was declined —
	// surfaced so a second attempt can address the objection.
	DecidedNote *string `json:"decidedNote"`
	// ProvisionedID is the vault or phone identity id once APPROVED.
	ProvisionedID *string `json:"provisionedId"`
	CreatedAt     string  `json:"createdAt"`
}

// CreateProvisioningRequestResult is what creating a request returns.
type CreateProvisioningRequestResult struct {
	ProvisioningRequest
	// EmailSent false does NOT mean the request failed — it is live and
	// visible in the console either way — but no human was told, so nothing
	// will happen until someone looks.
	EmailSent bool `json:"emailSent"`
}

// CreateProvisioningRequestParams is the body for filing a request.
type CreateProvisioningRequestParams struct {
	// AgentID is required with a master key; omit it when calling as the
	// agent itself.
	AgentID  string                `json:"agentId,omitempty"`
	Resource ProvisionableResource `json:"resource"`
	// Reason is shown verbatim to the owner. An unexplained ask is not a
	// decidable one, and the API requires it.
	Reason  string               `json:"reason"`
	Options *ProvisioningOptions `json:"options,omitempty"`
}

// DecideProvisioningRequestParams is the body for approve and decline.
type DecideProvisioningRequestParams struct {
	RequestID string `json:"requestId"`
	// Note reaches the agent, so a retry can address the objection.
	Note string `json:"note,omitempty"`
}

// ListProvisioningRequestsParams filters the request list.
type ListProvisioningRequestsParams struct {
	ListParams
	AgentID  string
	Status   ProvisioningRequestStatus
	Resource ProvisionableResource
}

// ToQuery converts ListProvisioningRequestsParams into URL query values.
func (p ListProvisioningRequestsParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	if p.Resource != "" {
		q.Set("resource", string(p.Resource))
	}
	return q
}

// ProvisioningRequestsService is how an agent asks its owner to provision
// something it cannot provision itself.
//
// Vault.Provision and Phones.Provision are both master-gated, and an agent key
// is never given master authority — so for an agent these are the only routes
// to a vault (after sign-up) or a phone number at all.
//
// The authority split is enforced server-side, not here: Create, List, Get and
// Cancel work with an agent key; Approve and Decline answer 403 without a
// master credential. An agent can open the conversation and withdraw from it,
// never conclude it.
type ProvisioningRequestsService struct {
	client *httpClient
}

func newProvisioningRequestsService(c *httpClient) *ProvisioningRequestsService {
	return &ProvisioningRequestsService{client: c}
}

// Create files a request and best-effort notifies the owner by email.
//
// Check EmailSent on the result: false does not mean the request failed, but
// it does mean no human was told.
//
// Filing an identical request while one is already pending returns the
// existing one rather than stacking duplicates in the owner's queue, so
// retrying after a timeout is safe.
func (s *ProvisioningRequestsService) Create(ctx context.Context, params CreateProvisioningRequestParams) (*CreateProvisioningRequestResult, error) {
	result, err := Do[CreateProvisioningRequestResult](ctx, s.client, http.MethodPost, "/provisioning-requests", params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns provisioning requests newest first. Agents see only their own;
// org credentials see the whole org.
func (s *ProvisioningRequestsService) List(ctx context.Context, params *ListProvisioningRequestsParams) (*Page[ProvisioningRequest], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[ProvisioningRequest]](ctx, s.client, http.MethodGet, "/provisioning-requests", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging iterates every provisioning request.
func (s *ProvisioningRequestsService) ListAutoPaging(params *ListProvisioningRequestsParams) *ListIterator[ProvisioningRequest] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[ProvisioningRequest], error) {
		p := &ListProvisioningRequestsParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Get returns one provisioning request.
func (s *ProvisioningRequestsService) Get(ctx context.Context, requestID string) (*ProvisioningRequest, error) {
	result, err := Do[ProvisioningRequest](ctx, s.client, http.MethodGet, fmt.Sprintf("/provisioning-requests/%s", url.PathEscape(requestID)), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Approve provisions the resource and marks the request APPROVED. Requires a
// master credential.
//
// Provisioning happens before the request is marked APPROVED, so a failure
// (plan too low, no numbers available, provider down) leaves it PENDING — fix
// the cause and approve again.
func (s *ProvisioningRequestsService) Approve(ctx context.Context, requestID string, note string) (*ProvisioningRequest, error) {
	body := DecideProvisioningRequestParams{RequestID: requestID, Note: note}
	result, err := Do[ProvisioningRequest](ctx, s.client, http.MethodPost, fmt.Sprintf("/provisioning-requests/%s/approve", url.PathEscape(requestID)), body, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Decline refuses the request. Requires a master credential.
//
// Soft — the agent may ask again, so pass a note saying what would change your
// mind.
func (s *ProvisioningRequestsService) Decline(ctx context.Context, requestID string, note string) (*ProvisioningRequest, error) {
	body := DecideProvisioningRequestParams{RequestID: requestID, Note: note}
	result, err := Do[ProvisioningRequest](ctx, s.client, http.MethodPost, fmt.Sprintf("/provisioning-requests/%s/decline", url.PathEscape(requestID)), body, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Cancel withdraws your own pending request. Works with an agent key.
func (s *ProvisioningRequestsService) Cancel(ctx context.Context, requestID string) (*ProvisioningRequest, error) {
	body := DecideProvisioningRequestParams{RequestID: requestID}
	result, err := Do[ProvisioningRequest](ctx, s.client, http.MethodPost, fmt.Sprintf("/provisioning-requests/%s/cancel", url.PathEscape(requestID)), body, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
