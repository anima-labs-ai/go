package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// WebhookEventType represents the type of event a webhook can subscribe to.
type WebhookEventType string

const (
	WebhookEventMessageReceived  WebhookEventType = "message.received"
	WebhookEventMessageSent      WebhookEventType = "message.sent"
	WebhookEventMessageFailed    WebhookEventType = "message.failed"
	WebhookEventMessageBounced   WebhookEventType = "message.bounced"
	WebhookEventAgentCreated     WebhookEventType = "agent.created"
	WebhookEventAgentUpdated     WebhookEventType = "agent.updated"
	WebhookEventAgentDeleted     WebhookEventType = "agent.deleted"
	WebhookEventPhoneProvisioned WebhookEventType = "phone.provisioned"
	WebhookEventPhoneReleased    WebhookEventType = "phone.released"
)

// WebhookAuthType is the auth scheme the platform presents to a webhook
// endpoint, as reported on reads. It never includes the secret itself.
type WebhookAuthType string

const (
	WebhookAuthTypeNone         WebhookAuthType = "NONE"
	WebhookAuthTypeBearer       WebhookAuthType = "BEARER"
	WebhookAuthTypeBasic        WebhookAuthType = "BASIC"
	WebhookAuthTypeCustomHeader WebhookAuthType = "CUSTOM_HEADER"
)

// Webhook represents a webhook configuration.
type Webhook struct {
	ID                  string             `json:"id"`
	OrgID               string             `json:"orgId"`
	URL                 string             `json:"url"`
	Events              []WebhookEventType `json:"events"`
	Active              bool               `json:"active"`
	Description         *string            `json:"description"`
	ConsecutiveFailures int                `json:"consecutiveFailures"`
	DisabledReason      *string            `json:"disabledReason"`
	DisabledAt          *string            `json:"disabledAt"`
	CreatedAt           string             `json:"createdAt"`
	UpdatedAt           string             `json:"updatedAt"`
	// AuthType is the auth scheme the platform presents to this endpoint. The
	// credential itself is write-only and is never returned.
	AuthType WebhookAuthType `json:"authType"`
	// AuthHeaderName is the header name for CUSTOM_HEADER auth; nil otherwise.
	AuthHeaderName *string `json:"authHeaderName"`
	// RateLimitPerMinute caps deliveries/minute to this endpoint; nil = unlimited.
	RateLimitPerMinute *int `json:"rateLimitPerMinute"`
	// MaxAttempts is the max delivery attempts before dead-lettering; nil = default (3).
	MaxAttempts *int `json:"maxAttempts"`
}

// WebhookAuthConfig is authentication the platform presents to your endpoint on
// each delivery, IN ADDITION to the always-on X-Anima-Signature HMAC — a
// lower-friction option for receivers that prefer a shared secret. Build one
// with a NewNoAuth / NewBearerAuth / NewBasicAuth / NewCustomHeaderAuth helper.
// The credential is write-only: it is never returned on reads.
type WebhookAuthConfig struct {
	// Type is the auth scheme: "none", "bearer", "basic", or "custom_header".
	Type string `json:"type"`
	// Token is the bearer token, for Type "bearer".
	Token string `json:"token,omitempty"`
	// Username is the basic-auth username, for Type "basic".
	Username string `json:"username,omitempty"`
	// Password is the basic-auth password, for Type "basic".
	Password string `json:"password,omitempty"`
	// HeaderName is the custom header name, for Type "custom_header".
	HeaderName string `json:"headerName,omitempty"`
	// Value is the custom header value, for Type "custom_header".
	Value string `json:"value,omitempty"`
}

// NewNoAuth returns an auth config that sends no customer auth header (the
// X-Anima-Signature HMAC is still sent). Pass it on Update to remove existing auth.
func NewNoAuth() *WebhookAuthConfig {
	return &WebhookAuthConfig{Type: "none"}
}

// NewBearerAuth returns a bearer-token auth config, sent to your endpoint as
// "Authorization: Bearer <token>".
func NewBearerAuth(token string) *WebhookAuthConfig {
	return &WebhookAuthConfig{Type: "bearer", Token: token}
}

// NewBasicAuth returns an HTTP basic auth config, sent to your endpoint as
// "Authorization: Basic <base64(username:password)>".
func NewBasicAuth(username, password string) *WebhookAuthConfig {
	return &WebhookAuthConfig{Type: "basic", Username: username, Password: password}
}

// NewCustomHeaderAuth returns a custom-header auth config, sent to your endpoint
// as "<headerName>: <value>".
func NewCustomHeaderAuth(headerName, value string) *WebhookAuthConfig {
	return &WebhookAuthConfig{Type: "custom_header", HeaderName: headerName, Value: value}
}

// CreateWebhookParams contains the parameters for creating a webhook.
type CreateWebhookParams struct {
	URL         string             `json:"url"`
	Events      []WebhookEventType `json:"events"`
	Description string             `json:"description,omitempty"`
	Active      *bool              `json:"active,omitempty"`
	// AuthConfig is optional auth the platform presents to your endpoint on each
	// delivery, in addition to the HMAC signature.
	AuthConfig *WebhookAuthConfig `json:"authConfig,omitempty"`
	// RateLimitPerMinute caps deliveries/minute to this endpoint; nil = unlimited.
	RateLimitPerMinute *int `json:"rateLimitPerMinute,omitempty"`
	// MaxAttempts is the max delivery attempts before dead-lettering (1-10, default 3).
	MaxAttempts *int `json:"maxAttempts,omitempty"`
}

// UpdateWebhookParams contains the parameters for updating a webhook.
type UpdateWebhookParams struct {
	URL         string             `json:"url,omitempty"`
	Events      []WebhookEventType `json:"events,omitempty"`
	Description string             `json:"description,omitempty"`
	Active      *bool              `json:"active,omitempty"`
	// AuthConfig replaces the endpoint auth; pass NewNoAuth() to remove it.
	AuthConfig *WebhookAuthConfig `json:"authConfig,omitempty"`
	// RateLimitPerMinute caps deliveries/minute; nil leaves it unchanged.
	RateLimitPerMinute *int `json:"rateLimitPerMinute,omitempty"`
	// MaxAttempts is the max delivery attempts (1-10); nil leaves it unchanged.
	MaxAttempts *int `json:"maxAttempts,omitempty"`
}

// WebhookListParams contains parameters for listing webhooks.
type WebhookListParams struct {
	ListParams
}

// WebhookTestResult contains the result of testing a webhook.
type WebhookTestResult struct {
	Success    bool   `json:"success"`
	DeliveryID string `json:"deliveryId"`
}

// WebhookDelivery represents a webhook delivery attempt.
type WebhookDelivery struct {
	ID            string                 `json:"id"`
	WebhookID     string                 `json:"webhookId"`
	MessageID     *string                `json:"messageId"`
	Event         WebhookEventType       `json:"event"`
	Payload       map[string]interface{} `json:"payload"`
	StatusCode    *int                   `json:"statusCode"`
	ResponseBody  *string                `json:"responseBody"`
	Attempts      int                    `json:"attempts"`
	MaxAttempts   int                    `json:"maxAttempts"`
	NextAttemptAt *string                `json:"nextAttemptAt"`
	CompletedAt   *string                `json:"completedAt"`
	CreatedAt     string                 `json:"createdAt"`
}

// WebhookDeliveryListParams contains parameters for listing webhook deliveries.
type WebhookDeliveryListParams struct {
	ListParams
}

// WebhooksService provides methods for managing webhooks.
type WebhooksService struct {
	client *httpClient
}

// newWebhooksService creates a new WebhooksService.
func newWebhooksService(c *httpClient) *WebhooksService {
	return &WebhooksService{client: c}
}

// Create creates a new webhook.
func (s *WebhooksService) Create(ctx context.Context, params CreateWebhookParams) (*Webhook, error) {
	wh, err := Do[Webhook](ctx, s.client, http.MethodPost, "/webhooks", params, nil)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

// Get retrieves a webhook by ID.
func (s *WebhooksService) Get(ctx context.Context, id string) (*Webhook, error) {
	wh, err := Do[Webhook](ctx, s.client, http.MethodGet, fmt.Sprintf("/webhooks/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

// List returns a paginated list of webhooks.
func (s *WebhooksService) List(ctx context.Context, params *WebhookListParams) (*Page[Webhook], error) {
	var q url.Values
	if params != nil {
		q = params.ListParams.ToQuery()
	}
	page, err := Do[Page[Webhook]](ctx, s.client, http.MethodGet, "/webhooks", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging returns an iterator that automatically paginates through all webhooks.
func (s *WebhooksService) ListAutoPaging(params *WebhookListParams) *ListIterator[Webhook] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[Webhook], error) {
		p := &WebhookListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// Update updates a webhook.
func (s *WebhooksService) Update(ctx context.Context, id string, params UpdateWebhookParams) (*Webhook, error) {
	wh, err := Do[Webhook](ctx, s.client, http.MethodPut, fmt.Sprintf("/webhooks/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

// Delete deletes a webhook.
func (s *WebhooksService) Delete(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/webhooks/%s", id), nil, nil)
	return err
}

// Test sends a test event to a webhook endpoint.
func (s *WebhooksService) Test(ctx context.Context, id string, event WebhookEventType) (*WebhookTestResult, error) {
	body := struct {
		ID    string           `json:"id"`
		Event WebhookEventType `json:"event,omitempty"`
	}{ID: id, Event: event}
	result, err := Do[WebhookTestResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/webhooks/%s/test", id), body, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListDeliveries returns a paginated list of delivery attempts for a webhook.
func (s *WebhooksService) ListDeliveries(ctx context.Context, webhookID string, params *WebhookDeliveryListParams) (*Page[WebhookDelivery], error) {
	q := url.Values{}
	q.Set("webhookId", webhookID)
	if params != nil {
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
	}
	page, err := Do[Page[WebhookDelivery]](ctx, s.client, http.MethodGet, fmt.Sprintf("/webhooks/%s/deliveries", webhookID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}
