package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// MessageChannel represents the channel a message was sent through.
type MessageChannel string

const (
	MessageChannelEmail MessageChannel = "EMAIL"
	MessageChannelSMS   MessageChannel = "SMS"
	MessageChannelMMS   MessageChannel = "MMS"
	MessageChannelVoice MessageChannel = "VOICE"
)

// MessageDirection indicates whether a message is inbound or outbound.
type MessageDirection string

const (
	MessageDirectionInbound  MessageDirection = "INBOUND"
	MessageDirectionOutbound MessageDirection = "OUTBOUND"
)

// MessageStatus represents the delivery status of a message.
type MessageStatus string

const (
	MessageStatusQueued          MessageStatus = "QUEUED"
	MessageStatusSent            MessageStatus = "SENT"
	MessageStatusDelivered       MessageStatus = "DELIVERED"
	MessageStatusFailed          MessageStatus = "FAILED"
	MessageStatusBounced         MessageStatus = "BOUNCED"
	MessageStatusBlocked         MessageStatus = "BLOCKED"
	MessageStatusPendingApproval MessageStatus = "PENDING_APPROVAL"
)

// AttachmentOutput represents an attachment on a message.
type AttachmentOutput struct {
	ID         string  `json:"id"`
	Filename   string  `json:"filename"`
	MimeType   string  `json:"mimeType"`
	SizeBytes  int64   `json:"sizeBytes"`
	StorageKey string  `json:"storageKey"`
	URL        *string `json:"url"`
	CreatedAt  string  `json:"createdAt"`
}

// Message represents a message in the Anima platform.
type Message struct {
	ID          string                 `json:"id"`
	AgentID     string                 `json:"agentId"`
	Channel     MessageChannel         `json:"channel"`
	Direction   MessageDirection       `json:"direction"`
	Status      MessageStatus          `json:"status"`
	FromAddress string                 `json:"fromAddress"`
	ToAddress   string                 `json:"toAddress"`
	Subject     *string                `json:"subject"`
	Body        string                 `json:"body"`
	BodyHTML    *string                `json:"bodyHtml"`
	Headers     map[string]interface{} `json:"headers"`
	Metadata    map[string]interface{} `json:"metadata"`
	ThreadID    *string                `json:"threadId"`
	InReplyTo   *string                `json:"inReplyTo"`
	ExternalID  *string                `json:"externalId"`
	SentAt      *string                `json:"sentAt"`
	ReceivedAt  *string                `json:"receivedAt"`
	Attachments []AttachmentOutput     `json:"attachments"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
}

// EmailAttachment is a file attachment for an outbound email. Provide exactly
// one of Content (base64-encoded bytes inline in the request) or URL (the
// server fetches the file). The total size cap is 25MB per email across all
// attachments; max 20 attachments per email.
type EmailAttachment struct {
	// Filename presented to the recipient. Inferred from the URL path when URL
	// is used and this is omitted; falls back to "attachment" when neither is
	// available.
	Filename string `json:"filename,omitempty"`
	// ContentID makes the attachment inline, for images referenced in the HTML
	// body via cid: URIs (e.g. set to "logo" to be referenced as
	// <img src="cid:logo">). When empty the attachment is a regular download.
	ContentID string `json:"contentId,omitempty"`
	// ContentType is the MIME type. Auto-detected from the Filename extension
	// if omitted.
	ContentType string `json:"contentType,omitempty"`
	// Content is the base64-encoded attachment bytes. Provide either Content
	// or URL, not both.
	Content string `json:"content,omitempty"`
	// URL is a public URL the server fetches and attaches. Provide either
	// Content or URL, not both.
	URL string `json:"url,omitempty"`
}

// SendEmailParams contains the parameters for sending an email.
type SendEmailParams struct {
	AgentID  string   `json:"agentId"`
	To       []string `json:"to"`
	CC       []string `json:"cc,omitempty"`
	BCC      []string `json:"bcc,omitempty"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	BodyHTML string   `json:"bodyHtml,omitempty"`
	// Attachments are optional file attachments (max 20, 25MB total).
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	// InReplyTo is the Message-ID of the email this is replying to, used for
	// threading.
	InReplyTo string `json:"inReplyTo,omitempty"`
	// References is the list of Message-IDs forming the email thread chain.
	References []string               `json:"references,omitempty"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SendSMSParams contains the parameters for sending an SMS.
type SendSMSParams struct {
	AgentID   string                 `json:"agentId"`
	To        string                 `json:"to"`
	Body      string                 `json:"body"`
	MediaURLs []string               `json:"mediaUrls,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DateRange represents a time range filter.
type DateRange struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// MessageListParams contains parameters for listing messages.
type MessageListParams struct {
	ListParams
	AgentID   string
	ThreadID  string
	Channel   MessageChannel
	Direction MessageDirection
	DateRange *DateRange
}

// ToQuery converts MessageListParams to URL query values.
func (p MessageListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.ThreadID != "" {
		q.Set("threadId", p.ThreadID)
	}
	if p.Channel != "" {
		q.Set("channel", string(p.Channel))
	}
	if p.Direction != "" {
		q.Set("direction", string(p.Direction))
	}
	if p.DateRange != nil {
		if p.DateRange.From != "" {
			q.Set("dateRange.from", p.DateRange.From)
		}
		if p.DateRange.To != "" {
			q.Set("dateRange.to", p.DateRange.To)
		}
	}
	return q
}

// MessageSearchFilters contains filters for searching messages.
type MessageSearchFilters struct {
	AgentID   string           `json:"agentId,omitempty"`
	Channel   MessageChannel   `json:"channel,omitempty"`
	Direction MessageDirection `json:"direction,omitempty"`
	Status    MessageStatus    `json:"status,omitempty"`
	DateRange *DateRange       `json:"dateRange,omitempty"`
}

// MessageSearchParams contains parameters for searching messages.
type MessageSearchParams struct {
	Query      string                `json:"query"`
	Filters    *MessageSearchFilters `json:"filters,omitempty"`
	Pagination *ListParams           `json:"pagination,omitempty"`
}

// SemanticSearchParams contains parameters for semantic (meaning-based)
// message search. Query is required; the other fields are optional.
type SemanticSearchParams struct {
	// Query is a natural-language description of what to find (1-1000 chars).
	Query string `json:"query"`
	// AgentID filters results to a specific agent.
	AgentID string `json:"agentId,omitempty"`
	// Limit is the maximum number of results (1-50, server default 10).
	Limit int `json:"limit,omitempty"`
	// Threshold is the minimum cosine-similarity score (0-1, server default
	// 0.7). It is a pointer so an explicit 0 ("return everything") is
	// distinguishable from unset.
	Threshold *float64 `json:"threshold,omitempty"`
}

// SemanticSearchResult is a single message matched by semantic search.
type SemanticSearchResult struct {
	ID string `json:"id"`
	// Content is the message content text.
	Content string `json:"content"`
	// Similarity is the cosine similarity score between 0 and 1.
	Similarity float64          `json:"similarity"`
	Channel    MessageChannel   `json:"channel"`
	Direction  MessageDirection `json:"direction"`
	CreatedAt  string           `json:"createdAt"`
	AgentID    string           `json:"agentId"`
}

// SemanticSearchResults contains messages ranked by semantic similarity.
type SemanticSearchResults struct {
	Results []SemanticSearchResult `json:"results"`
}

// MessagesService provides methods for sending and managing messages.
type MessagesService struct {
	client *httpClient
}

// newMessagesService creates a new MessagesService.
func newMessagesService(c *httpClient) *MessagesService {
	return &MessagesService{client: c}
}

// SendEmail sends an email message.
func (s *MessagesService) SendEmail(ctx context.Context, params SendEmailParams) (*Message, error) {
	msg, err := Do[Message](ctx, s.client, http.MethodPost, "/messages/email", params, nil)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// SendSMS sends an SMS message.
func (s *MessagesService) SendSMS(ctx context.Context, params SendSMSParams) (*Message, error) {
	msg, err := Do[Message](ctx, s.client, http.MethodPost, "/phone/send-sms", params, nil)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// Get retrieves a message by ID.
func (s *MessagesService) Get(ctx context.Context, id string) (*Message, error) {
	msg, err := Do[Message](ctx, s.client, http.MethodGet, fmt.Sprintf("/messages/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// List returns a paginated list of messages.
func (s *MessagesService) List(ctx context.Context, params *MessageListParams) (*Page[Message], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[Message]](ctx, s.client, http.MethodGet, "/messages", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging returns an iterator that automatically paginates through all messages.
func (s *MessagesService) ListAutoPaging(params *MessageListParams) *ListIterator[Message] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[Message], error) {
		p := &MessageListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Search searches messages with full-text search and filters.
func (s *MessagesService) Search(ctx context.Context, params MessageSearchParams) (*Page[Message], error) {
	page, err := Do[Page[Message]](ctx, s.client, http.MethodPost, "/messages/search", params, nil)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// SemanticSearch searches messages by meaning rather than keywords, using
// vector embeddings server-side. Results are ranked by cosine similarity.
//
// An empty Results slice means nothing matched above the threshold — it is
// not an error. An embedding-provider outage surfaces as an *APIError with
// a 5xx status (matching ErrInternalServer), so callers can distinguish
// "no results" from "search unavailable".
func (s *MessagesService) SemanticSearch(ctx context.Context, params SemanticSearchParams) (*SemanticSearchResults, error) {
	res, err := Do[SemanticSearchResults](ctx, s.client, http.MethodPost, "/messages/search/semantic", params, nil)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
