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

// SendEmailParams contains the parameters for sending an email.
type SendEmailParams struct {
	AgentID  string                 `json:"agentId"`
	To       []string               `json:"to"`
	CC       []string               `json:"cc,omitempty"`
	BCC      []string               `json:"bcc,omitempty"`
	Subject  string                 `json:"subject"`
	Body     string                 `json:"body"`
	BodyHTML string                 `json:"bodyHtml,omitempty"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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
	msg, err := Do[Message](ctx, s.client, http.MethodPost, "/messages/sms", params, nil)
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
