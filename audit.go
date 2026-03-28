package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AuditActorType represents the type of actor that performed an audited action.
type AuditActorType string

const (
	AuditActorAPIKey AuditActorType = "api_key"
	AuditActorUser   AuditActorType = "user"
	AuditActorSystem AuditActorType = "system"
	AuditActorAgent  AuditActorType = "agent"
)

// AuditResult represents the outcome of an audited action.
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
	AuditResultDenied  AuditResult = "denied"
)

// AuditExportFormat represents the export format for audit logs.
type AuditExportFormat string

const (
	AuditExportFormatCSV  AuditExportFormat = "csv"
	AuditExportFormatJSON AuditExportFormat = "json"
)

// AuditLog represents an immutable audit log entry.
type AuditLog struct {
	ID           string                 `json:"id"`
	OrgID        string                 `json:"orgId"`
	ActorType    AuditActorType         `json:"actorType"`
	ActorID      string                 `json:"actorId"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resourceType"`
	ResourceID   string                 `json:"resourceId"`
	Result       AuditResult            `json:"result"`
	IPAddress    *string                `json:"ipAddress"`
	UserAgent    *string                `json:"userAgent"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    string                 `json:"createdAt"`
}

// AuditLogListParams contains parameters for listing audit log entries.
type AuditLogListParams struct {
	ListParams
	ActorID      string
	ActorType    AuditActorType
	Action       string
	ResourceType string
	ResourceID   string
	Result       AuditResult
	From         string
	To           string
}

// ToQuery converts AuditLogListParams to URL query values.
func (p AuditLogListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.ActorID != "" {
		q.Set("actorId", p.ActorID)
	}
	if p.ActorType != "" {
		q.Set("actorType", string(p.ActorType))
	}
	if p.Action != "" {
		q.Set("action", p.Action)
	}
	if p.ResourceType != "" {
		q.Set("resourceType", p.ResourceType)
	}
	if p.ResourceID != "" {
		q.Set("resourceId", p.ResourceID)
	}
	if p.Result != "" {
		q.Set("result", string(p.Result))
	}
	if p.From != "" {
		q.Set("from", p.From)
	}
	if p.To != "" {
		q.Set("to", p.To)
	}
	return q
}

// AuditLogExportParams contains parameters for exporting audit logs.
type AuditLogExportParams struct {
	Format       AuditExportFormat `json:"format,omitempty"`
	From         string            `json:"from,omitempty"`
	To           string            `json:"to,omitempty"`
	ActorID      string            `json:"actorId,omitempty"`
	Action       string            `json:"action,omitempty"`
	ResourceType string            `json:"resourceType,omitempty"`
}

// AuditLogExportOutput contains the result of an audit log export.
type AuditLogExportOutput struct {
	URL         string            `json:"url"`
	Format      AuditExportFormat `json:"format"`
	RecordCount int               `json:"recordCount"`
	ExpiresAt   string            `json:"expiresAt"`
}

// AuditService provides methods for querying and exporting immutable audit logs.
type AuditService struct {
	client *httpClient
}

// newAuditService creates a new AuditService.
func newAuditService(c *httpClient) *AuditService {
	return &AuditService{client: c}
}

// List returns a paginated list of audit log entries for an organization.
func (s *AuditService) List(ctx context.Context, orgID string, params *AuditLogListParams) (*Page[AuditLog], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[AuditLog]](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/audit/logs", orgID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging returns an iterator that automatically paginates through all audit log entries.
func (s *AuditService) ListAutoPaging(orgID string, params *AuditLogListParams) *ListIterator[AuditLog] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[AuditLog], error) {
		p := &AuditLogListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, orgID, p)
	})
}

// Get retrieves a single audit log entry by ID.
func (s *AuditService) Get(ctx context.Context, orgID, logID string) (*AuditLog, error) {
	log, err := Do[AuditLog](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/audit/logs/%s", orgID, logID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// Export initiates an audit log export and returns a download URL.
func (s *AuditService) Export(ctx context.Context, orgID string, params *AuditLogExportParams) (*AuditLogExportOutput, error) {
	result, err := Do[AuditLogExportOutput](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/audit/export", orgID), params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
