package anima

import (
	"context"
	"fmt"
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
	PhoneNumber  string            `json:"phoneNumber"`
	Region       string            `json:"region,omitempty"`
	Capabilities PhoneCapabilities `json:"capabilities"`
	MonthlyCost  float64           `json:"monthlyCost,omitempty"`
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

// PhoneIdentityListItem is a number plus the agent that owns it, as returned
// by ListIdentities. List is scoped to one agent and needs no such joining;
// this list spans agents, so the row has to say whose each number is.
type PhoneIdentityListItem struct {
	PhoneIdentity
	AgentID   string `json:"agentId"`
	AgentName string `json:"agentName"`
	AgentSlug string `json:"agentSlug"`
}

// ListPhoneIdentitiesParams filters the organization's phone numbers.
type ListPhoneIdentitiesParams struct {
	ListParams
	// Query is a free-text search. Digits match the number itself with
	// punctuation and spacing ignored, so "(555) 123" and "555-123" both find
	// +15551234567; the raw term also matches the owning agent's name or slug.
	Query string
	// AgentID restricts the list to one agent's numbers. Empty lists every
	// number in the organization. An agent key is confined to its own numbers
	// whether or not this is set.
	AgentID string
}

// ToQuery converts ListPhoneIdentitiesParams into URL query values.
func (p ListPhoneIdentitiesParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Query != "" {
		q.Set("query", p.Query)
	}
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	return q
}

// SMSMessageDirection is the direction of an SMS message.
type SMSMessageDirection string

// SMS message directions.
const (
	SMSMessageDirectionInbound  SMSMessageDirection = "INBOUND"
	SMSMessageDirectionOutbound SMSMessageDirection = "OUTBOUND"
)

// SMSThread summarises one SMS conversation.
//
// An SMS conversation is the pair (agent, counterparty E.164) — SMS has no
// In-Reply-To chain to thread on, so the pair IS the thread. ThreadID is the id
// of the conversation's first message, so Messages.List with that ThreadID
// returns the whole conversation.
type SMSThread struct {
	ThreadID string `json:"threadId"`
	AgentID  string `json:"agentId"`
	// ParticipantAddress is the counterparty's number (E.164).
	ParticipantAddress string `json:"participantAddress"`
	// AgentAddress is the agent's own number used in this conversation.
	AgentAddress         string              `json:"agentAddress"`
	LastMessageAt        string              `json:"lastMessageAt"`
	LastMessageSnippet   string              `json:"lastMessageSnippet"`
	LastMessageDirection SMSMessageDirection `json:"lastMessageDirection"`
	MessageCount         int                 `json:"messageCount"`
	// UnreadCount counts messages still carrying the unread label. Inbound
	// texts start unread.
	UnreadCount int `json:"unreadCount"`
}

// SMSThreadListParams filters SMS conversations.
//
// Offset-paged, not cursor-paged: this is the one list on this SDK that is, so
// it does not embed ListParams and has no AutoPaging sibling.
type SMSThreadListParams struct {
	// AgentID filters to one agent. Optional for master keys (empty lists the
	// whole org); ignored for agent keys, which always see only their own.
	AgentID string
	// Limit is 1-100. Zero sends none and the server defaults to 20.
	Limit int
	// Offset is the number of conversations to skip.
	Offset int
	// Unread, when non-nil and true, returns only conversations with at least
	// one unread message. A pointer because false is meaningful — it is the
	// explicit "All" filter, not "unset".
	Unread *bool
}

// ToQuery converts SMSThreadListParams into URL query values.
func (p SMSThreadListParams) ToQuery() url.Values {
	q := url.Values{}
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		q.Set("offset", strconv.Itoa(p.Offset))
	}
	if p.Unread != nil {
		q.Set("unread", strconv.FormatBool(*p.Unread))
	}
	return q
}

// SMSThreadList is a page of SMS conversations.
//
// {items, total, hasMore} with offset/limit — neither the {items, pagination}
// nor the {items, nextCursor} envelope, which is why this is a plain struct
// and not a Page[T]: decoding it as one yields HasMore false on the first page
// and silently truncates.
type SMSThreadList struct {
	Items []SMSThread `json:"items"`
	// Total counts every conversation matching the query, not just this page.
	// With Unread set it counts unread conversations.
	Total   int  `json:"total"`
	HasMore bool `json:"hasMore"`
}

// SMSThreadGetParams tunes how much of a conversation is returned.
type SMSThreadGetParams struct {
	// Limit is 1-100; the server defaults to 50. When the conversation is
	// longer, the MOST RECENT Limit messages are returned, still oldest-first
	// among themselves. Reach deeper history via Messages.List with the
	// conversation's ThreadID.
	Limit int
}

// SMSThreadDetail is one SMS conversation with its message history.
type SMSThreadDetail struct {
	ThreadID           string `json:"threadId"`
	AgentID            string `json:"agentId"`
	ParticipantAddress string `json:"participantAddress"`
	AgentAddress       string `json:"agentAddress"`
	// Messages are in chronological (oldest-first) reading order.
	Messages     []Message `json:"messages"`
	MessageCount int       `json:"messageCount"`
	// HasMore is true when the conversation holds more messages than this
	// response returned; the older ones precede Messages.
	HasMore bool `json:"hasMore"`
}

// SMSThreadStatsParams scopes the per-agent conversation totals.
type SMSThreadStatsParams struct {
	// AgentID restricts to one agent. Empty covers every agent in the org;
	// ignored for agent keys, which always see only their own.
	AgentID string
}

// SMSThreadStat holds one agent's SMS conversation totals.
type SMSThreadStat struct {
	AgentID       string `json:"agentId"`
	Conversations int    `json:"conversations"`
	// Unread counts unread inbound messages across those conversations.
	Unread        int     `json:"unread"`
	LastMessageAt *string `json:"lastMessageAt"`
}

// SMSThreadStatList wraps the per-agent SMS totals.
type SMSThreadStatList struct {
	Items []SMSThreadStat `json:"items"`
}

// SMSSuppressionReason is why a number is suppressed.
type SMSSuppressionReason string

// SMS suppression reasons.
const (
	SMSSuppressionReasonStopKeyword SMSSuppressionReason = "STOP_KEYWORD"
	SMSSuppressionReasonManual      SMSSuppressionReason = "MANUAL"
)

// SMSSuppression is a recipient that SMS sends are refused for. Sends to a
// suppressed number fail with RECIPIENT_OPTED_OUT.
type SMSSuppression struct {
	ID string `json:"id"`
	// PhoneNumber is the suppressed recipient (E.164).
	PhoneNumber string `json:"phoneNumber"`
	// AgentID is the agent the suppression is scoped to, nil when
	// workspace-wide.
	AgentID *string              `json:"agentId"`
	Reason  SMSSuppressionReason `json:"reason"`
	// Source is a free-form origin marker, e.g. "inbound-stop-keyword".
	Source    *string `json:"source"`
	CreatedAt string  `json:"createdAt"`
}

// SMSSuppressionListParams filters the suppression list.
type SMSSuppressionListParams struct {
	ListParams
	// PhoneNumber filters to one recipient (E.164 or close; normalized
	// server-side).
	PhoneNumber string
}

// ToQuery converts SMSSuppressionListParams into URL query values.
func (p SMSSuppressionListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.PhoneNumber != "" {
		q.Set("phoneNumber", p.PhoneNumber)
	}
	return q
}

// UnsuppressSMSParams names the recipient to unsuppress.
type UnsuppressSMSParams struct {
	// PhoneNumber is the recipient to unsuppress (E.164 or close; normalized
	// server-side).
	PhoneNumber string `json:"phoneNumber"`
}

// UnsuppressSMSResult reports what the override removed.
type UnsuppressSMSResult struct {
	// PhoneNumber is the normalized number that was unsuppressed.
	PhoneNumber string `json:"phoneNumber"`
	// Removed is how many suppression entries were deleted.
	Removed int `json:"removed"`
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

// ListIdentities returns every number in the organization with the agent that
// owns each.
//
// The org-wide sibling of List, which answers "what does this agent own" and is
// naturally small. This one answers "what does the org own" and is not, so it
// pages.
func (s *PhonesService) ListIdentities(ctx context.Context, params *ListPhoneIdentitiesParams) (*Page[PhoneIdentityListItem], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[PhoneIdentityListItem]](ctx, s.client, http.MethodGet, "/phone/identities", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListIdentitiesAutoPaging iterates every number in the organization, fetching
// pages as needed.
func (s *PhonesService) ListIdentitiesAutoPaging(params *ListPhoneIdentitiesParams) *ListIterator[PhoneIdentityListItem] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[PhoneIdentityListItem], error) {
		p := &ListPhoneIdentitiesParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListIdentities(ctx, p)
	})
}

// ListSMSThreads returns SMS conversations, newest activity first.
//
// Offset-paged, so this returns a plain page rather than a Page[T]: advance
// with SMSThreadListParams.Offset and stop when HasMore is false.
func (s *PhonesService) ListSMSThreads(ctx context.Context, params *SMSThreadListParams) (*SMSThreadList, error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	list, err := Do[SMSThreadList](ctx, s.client, http.MethodGet, "/phone/sms/threads", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// GetSMSThread returns one conversation with its message history. The id is a
// ThreadID from ListSMSThreads or Message.ThreadID.
func (s *PhonesService) GetSMSThread(ctx context.Context, id string, params *SMSThreadGetParams) (*SMSThreadDetail, error) {
	q := url.Values{}
	if params != nil && params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	thread, err := Do[SMSThreadDetail](ctx, s.client, http.MethodGet, fmt.Sprintf("/phone/sms/threads/%s", url.PathEscape(id)), nil, q)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

// SMSThreadStats returns per-agent conversation totals for an SMS overview.
//
// An aggregate, so it is both correct and cheaper than counting a page of
// ListSMSThreads client-side — that approach makes every number a lower bound
// once the org exceeds one page, with nothing saying so.
func (s *PhonesService) SMSThreadStats(ctx context.Context, params *SMSThreadStatsParams) (*SMSThreadStatList, error) {
	q := url.Values{}
	if params != nil && params.AgentID != "" {
		q.Set("agentId", params.AgentID)
	}
	stats, err := Do[SMSThreadStatList](ctx, s.client, http.MethodGet, "/phone/sms/stats", nil, q)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// ListSMSSuppressions returns recipients that SMS sends are refused for.
// Master-key only.
func (s *PhonesService) ListSMSSuppressions(ctx context.Context, params *SMSSuppressionListParams) (*Page[SMSSuppression], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[SMSSuppression]](ctx, s.client, http.MethodGet, "/phone/sms-suppressions", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListSMSSuppressionsAutoPaging iterates every SMS suppression, fetching pages
// as needed.
func (s *PhonesService) ListSMSSuppressionsAutoPaging(params *SMSSuppressionListParams) *ListIterator[SMSSuppression] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[SMSSuppression], error) {
		p := &SMSSuppressionListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListSMSSuppressions(ctx, p)
	})
}

// UnsuppressSMS removes every SMS suppression for a number in this org.
// Master-key only.
//
// Use sparingly: a suppression records the recipient's own STOP, so reversing
// one is an org-owner decision, never an agent's. The recipient-driven lift is
// texting START, which needs no call here.
func (s *PhonesService) UnsuppressSMS(ctx context.Context, params UnsuppressSMSParams) (*UnsuppressSMSResult, error) {
	result, err := Do[UnsuppressSMSResult](ctx, s.client, http.MethodPost, "/phone/sms-unsuppress", params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
