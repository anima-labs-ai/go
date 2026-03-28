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

// SecurityScanParams contains the parameters for a content security scan.
type SecurityScanParams struct {
	OrgID    string                 `json:"orgId"`
	AgentID  string                 `json:"agentId,omitempty"`
	Channel  string                 `json:"channel"` // "EMAIL" or "SMS"
	Subject  string                 `json:"subject,omitempty"`
	Body     string                 `json:"body"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityScanWarning represents a warning from a content scan.
type SecurityScanWarning struct {
	RuleID      string           `json:"ruleId"`
	Severity    SecuritySeverity `json:"severity"`
	Description string           `json:"description"`
	Match       string           `json:"match,omitempty"`
}

// SecurityScanResult contains the result of a content security scan.
type SecurityScanResult struct {
	Blocked  bool                  `json:"blocked"`
	Warnings []SecurityScanWarning `json:"warnings"`
	Summary  string                `json:"summary"`
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

// ScanContent scans message content for security threats such as PII leakage
// or prompt injection.
func (s *SecurityService) ScanContent(ctx context.Context, params SecurityScanParams) (*SecurityScanResult, error) {
	result, err := Do[SecurityScanResult](ctx, s.client, http.MethodPost, "/security/scan", params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListEvents returns a paginated list of security events.
func (s *SecurityService) ListEvents(ctx context.Context, params SecurityEventsListParams) (*Page[SecurityEvent], error) {
	q := params.ToQuery()
	page, err := Do[Page[SecurityEvent]](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/security/events", params.OrgID), nil, q)
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
