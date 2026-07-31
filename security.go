package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// SecurityEventType represents the type of a security event.
type SecurityEventType string

const (
	SecurityEventPIIDetected       SecurityEventType = "PII_DETECTED"
	SecurityEventInjectionDetected SecurityEventType = "INJECTION_DETECTED"
	SecurityEventRateLimited       SecurityEventType = "RATE_LIMITED"
	SecurityEventBlocked           SecurityEventType = "BLOCKED"
	SecurityEventApproved          SecurityEventType = "APPROVED"
	SecurityEventRejected          SecurityEventType = "REJECTED"
)

// SecuritySeverity represents the severity of a security event or warning.
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "LOW"
	SecuritySeverityMedium   SecuritySeverity = "MEDIUM"
	SecuritySeverityHigh     SecuritySeverity = "HIGH"
	SecuritySeverityCritical SecuritySeverity = "CRITICAL"
)

// SecurityScanParams, SecurityScanWarning and SecurityScanResult used to sit
// here. They typed a POST /security/scan endpoint the API has never served.

// AiScannerStatus describes the AI content scanner's health.
type AiScannerStatus struct {
	// Active reports whether the scanner runs on message traffic, not merely
	// whether an LLM provider is configured.
	Active         bool    `json:"active"`
	Provider       *string `json:"provider"`
	FallbackReason *string `json:"fallbackReason"`
}

// ScannerStatus is the current scanner health for an organization.
type ScannerStatus struct {
	AiScanner AiScannerStatus `json:"aiScanner"`
}

// SecurityEvent represents a security event in the Anima platform.
type SecurityEvent struct {
	ID         string                 `json:"id"`
	OrgID      string                 `json:"orgId"`
	AgentID    *string                `json:"agentId"`
	MessageID  *string                `json:"messageId"`
	Type       SecurityEventType      `json:"type"`
	Severity   SecuritySeverity       `json:"severity"`
	Details    map[string]interface{} `json:"details"`
	Resolved   bool                   `json:"resolved"`
	ResolvedBy *string                `json:"resolvedBy"`
	ResolvedAt *string                `json:"resolvedAt"`
	CreatedAt  string                 `json:"createdAt"`
}

// SecurityEventsListParams contains parameters for listing security events.
type SecurityEventsListParams struct {
	ListParams
	OrgID    string
	AgentID  string
	Type     SecurityEventType
	Severity SecuritySeverity
}

// ToQuery converts SecurityEventsListParams to URL query values.
func (p SecurityEventsListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	q.Set("orgId", p.OrgID)
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.Type != "" {
		q.Set("type", string(p.Type))
	}
	if p.Severity != "" {
		q.Set("severity", string(p.Severity))
	}
	return q
}

// SecurityService provides methods for content scanning and security events.
type SecurityService struct {
	client *httpClient
}

// newSecurityService creates a new SecurityService.
func newSecurityService(c *httpClient) *SecurityService {
	return &SecurityService{client: c}
}

// No ScanContent. It POSTed to /security/scan, which the API has never served
// — scanning runs inside the send paths, not as a callable route. The security
// surface the API does expose is the event feed and the scanner status below.

// GetScannerStatus reports whether the AI content scanner is running on message
// traffic, and why it is in fallback if it is not.
func (s *SecurityService) GetScannerStatus(ctx context.Context, orgID string) (*ScannerStatus, error) {
	result, err := Do[ScannerStatus](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/security/scanner-status", orgID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListEvents returns a paginated list of security events.
func (s *SecurityService) ListEvents(ctx context.Context, params SecurityEventsListParams) (*Page[SecurityEvent], error) {
	q := params.ToQuery()
	page, err := Do[Page[SecurityEvent]](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/security/events", params.OrgID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListEventsAutoPaging returns an iterator that automatically paginates through all security events.
func (s *SecurityService) ListEventsAutoPaging(params SecurityEventsListParams) *ListIterator[SecurityEvent] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[SecurityEvent], error) {
		p := params
		p.Cursor = cursor
		return s.ListEvents(ctx, p)
	})
}
