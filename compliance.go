package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ComplianceFramework represents a compliance framework (e.g. SOC2, GDPR, PCI).
type ComplianceFramework string

const (
	ComplianceFrameworkSOC2 ComplianceFramework = "SOC2"
	ComplianceFrameworkGDPR ComplianceFramework = "GDPR"
	ComplianceFrameworkPCI  ComplianceFramework = "PCI"
)

// ComplianceControlStatus represents the status of a compliance control.
type ComplianceControlStatus string

const (
	ComplianceControlStatusNotStarted  ComplianceControlStatus = "not_started"
	ComplianceControlStatusInProgress  ComplianceControlStatus = "in_progress"
	ComplianceControlStatusImplemented ComplianceControlStatus = "implemented"
	ComplianceControlStatusVerified    ComplianceControlStatus = "verified"
	ComplianceControlStatusFailed      ComplianceControlStatus = "failed"
)

// ComplianceReportType represents the type of a compliance report.
type ComplianceReportType string

const (
	ComplianceReportTypeSOC2Summary    ComplianceReportType = "soc2_summary"
	ComplianceReportTypeActivityReport ComplianceReportType = "activity_report"
	ComplianceReportTypeAccessReview   ComplianceReportType = "access_review"
	ComplianceReportTypeAuditExport    ComplianceReportType = "audit_export"
	ComplianceReportTypeGDPRDSAR       ComplianceReportType = "gdpr_dsar"
)

// DSARStatus represents the status of a Data Subject Access Request.
type DSARStatus string

const (
	DSARStatusPending    DSARStatus = "pending"
	DSARStatusInProgress DSARStatus = "in_progress"
	DSARStatusCompleted  DSARStatus = "completed"
	DSARStatusRejected   DSARStatus = "rejected"
)

// DSARRequestType represents the type of a DSAR.
type DSARRequestType string

const (
	DSARRequestTypeAccess        DSARRequestType = "access"
	DSARRequestTypeDeletion      DSARRequestType = "deletion"
	DSARRequestTypeRectification DSARRequestType = "rectification"
	DSARRequestTypePortability   DSARRequestType = "portability"
)

// ComplianceControl represents a compliance control within a framework.
type ComplianceControl struct {
	ID           string                  `json:"id"`
	OrgID        string                  `json:"orgId"`
	Framework    ComplianceFramework     `json:"framework"`
	ControlID    string                  `json:"controlId"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	Category     string                  `json:"category"`
	Status       ComplianceControlStatus `json:"status"`
	Owner        *string                 `json:"owner"`
	LastTestedAt *string                 `json:"lastTestedAt"`
	NextReviewAt *string                 `json:"nextReviewAt"`
	CreatedAt    string                  `json:"createdAt"`
	UpdatedAt    string                  `json:"updatedAt"`
}

// ComplianceControlListParams contains parameters for listing compliance controls.
type ComplianceControlListParams struct {
	ListParams
	Framework ComplianceFramework
	Category  string
	Status    ComplianceControlStatus
}

// ToQuery converts ComplianceControlListParams to URL query values.
func (p ComplianceControlListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Framework != "" {
		q.Set("framework", string(p.Framework))
	}
	if p.Category != "" {
		q.Set("category", p.Category)
	}
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	return q
}

// ComplianceControlStatusInput contains parameters for updating a control's status.
type ComplianceControlStatusInput struct {
	Status ComplianceControlStatus `json:"status"`
	Owner  string                  `json:"owner,omitempty"`
}

// SeedFrameworkInput contains parameters for seeding a compliance framework.
type SeedFrameworkInput struct {
	Framework ComplianceFramework `json:"framework"`
}

// SeedFrameworkOutput contains the result of seeding a compliance framework.
type SeedFrameworkOutput struct {
	ControlsCreated int                 `json:"controlsCreated"`
	Framework       ComplianceFramework `json:"framework"`
}

// GenerateReportInput contains parameters for generating a compliance report.
type GenerateReportInput struct {
	Type     ComplianceReportType   `json:"type"`
	From     string                 `json:"from,omitempty"`
	To       string                 `json:"to,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceReport represents a generated compliance report.
type ComplianceReport struct {
	ID          string                 `json:"id"`
	OrgID       string                 `json:"orgId"`
	Type        ComplianceReportType   `json:"type"`
	Status      string                 `json:"status"`
	Title       string                 `json:"title"`
	Summary     *string                `json:"summary"`
	Data        map[string]interface{} `json:"data"`
	GeneratedAt string                 `json:"generatedAt"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
}

// ComplianceReportListParams contains parameters for listing compliance reports.
type ComplianceReportListParams struct {
	ListParams
	Type ComplianceReportType
}

// ToQuery converts ComplianceReportListParams to URL query values.
func (p ComplianceReportListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Type != "" {
		q.Set("type", string(p.Type))
	}
	return q
}

// ComplianceReportDownloadOutput contains a download URL for a compliance report.
type ComplianceReportDownloadOutput struct {
	URL       string `json:"url"`
	Format    string `json:"format"`
	ExpiresAt string `json:"expiresAt"`
}

// ComplianceFrameworkSummary contains summary statistics for a single framework.
type ComplianceFrameworkSummary struct {
	TotalControls int     `json:"totalControls"`
	Implemented   int     `json:"implemented"`
	Verified      int     `json:"verified"`
	Failed        int     `json:"failed"`
	NotStarted    int     `json:"notStarted"`
	Score         float64 `json:"score"`
}

// ComplianceDashboard contains an overview of compliance status across frameworks.
type ComplianceDashboard struct {
	OrgID          string                                `json:"orgId"`
	Frameworks     map[string]ComplianceFrameworkSummary  `json:"frameworks"`
	OverallScore   float64                               `json:"overallScore"`
	RecentActivity []ComplianceReport                    `json:"recentActivity"`
}

// CreateDSARInput contains parameters for creating a Data Subject Access Request.
type CreateDSARInput struct {
	SubjectEmail string                 `json:"subjectEmail"`
	RequestType  DSARRequestType        `json:"requestType"`
	Description  string                 `json:"description,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DataSubjectRequest represents a GDPR Data Subject Access Request.
type DataSubjectRequest struct {
	ID           string                 `json:"id"`
	OrgID        string                 `json:"orgId"`
	SubjectEmail string                 `json:"subjectEmail"`
	RequestType  string                 `json:"requestType"`
	Status       DSARStatus             `json:"status"`
	Description  *string                `json:"description"`
	Metadata     map[string]interface{} `json:"metadata"`
	CompletedAt  *string                `json:"completedAt"`
	CreatedAt    string                 `json:"createdAt"`
	UpdatedAt    string                 `json:"updatedAt"`
}

// DSARListParams contains parameters for listing DSARs.
type DSARListParams struct {
	ListParams
	Status DSARStatus
}

// ToQuery converts DSARListParams to URL query values.
func (p DSARListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	return q
}

// CompleteDSARInput contains parameters for completing a DSAR.
type CompleteDSARInput struct {
	Notes    string                 `json:"notes,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceService provides methods for managing compliance controls, reports,
// dashboards, and Data Subject Access Requests (DSARs).
type ComplianceService struct {
	client *httpClient
}

// newComplianceService creates a new ComplianceService.
func newComplianceService(c *httpClient) *ComplianceService {
	return &ComplianceService{client: c}
}

// ListControls returns a paginated list of compliance controls for an organization.
func (s *ComplianceService) ListControls(ctx context.Context, orgID string, params *ComplianceControlListParams) (*Page[ComplianceControl], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[ComplianceControl]](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/compliance/controls", orgID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListControlsAutoPaging returns an iterator that automatically paginates through all compliance controls.
func (s *ComplianceService) ListControlsAutoPaging(orgID string, params *ComplianceControlListParams) *ListIterator[ComplianceControl] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[ComplianceControl], error) {
		p := &ComplianceControlListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListControls(ctx, orgID, p)
	})
}

// GetControl retrieves a single compliance control by ID.
func (s *ComplianceService) GetControl(ctx context.Context, orgID, controlID string) (*ComplianceControl, error) {
	ctrl, err := Do[ComplianceControl](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/compliance/controls/%s", orgID, controlID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &ctrl, nil
}

// UpdateControlStatus updates the status and optionally the owner of a compliance control.
func (s *ComplianceService) UpdateControlStatus(ctx context.Context, orgID, controlID string, input ComplianceControlStatusInput) (*ComplianceControl, error) {
	ctrl, err := Do[ComplianceControl](ctx, s.client, http.MethodPatch, fmt.Sprintf("/v1/orgs/%s/compliance/controls/%s", orgID, controlID), input, nil)
	if err != nil {
		return nil, err
	}
	return &ctrl, nil
}

// SeedFramework seeds all predefined controls for a compliance framework (e.g. SOC2).
func (s *ComplianceService) SeedFramework(ctx context.Context, orgID string, input SeedFrameworkInput) (*SeedFrameworkOutput, error) {
	result, err := Do[SeedFrameworkOutput](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/compliance/seed", orgID), input, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateReport generates a new compliance report.
func (s *ComplianceService) GenerateReport(ctx context.Context, orgID string, input GenerateReportInput) (*ComplianceReport, error) {
	report, err := Do[ComplianceReport](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/compliance/reports", orgID), input, nil)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReports returns a paginated list of compliance reports for an organization.
func (s *ComplianceService) ListReports(ctx context.Context, orgID string, params *ComplianceReportListParams) (*Page[ComplianceReport], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[ComplianceReport]](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/compliance/reports", orgID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListReportsAutoPaging returns an iterator that automatically paginates through all compliance reports.
func (s *ComplianceService) ListReportsAutoPaging(orgID string, params *ComplianceReportListParams) *ListIterator[ComplianceReport] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[ComplianceReport], error) {
		p := &ComplianceReportListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListReports(ctx, orgID, p)
	})
}

// GetReport retrieves a single compliance report by ID.
func (s *ComplianceService) GetReport(ctx context.Context, orgID, reportID string) (*ComplianceReport, error) {
	report, err := Do[ComplianceReport](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/compliance/reports/%s", orgID, reportID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// DownloadReport returns a temporary download URL for a compliance report.
func (s *ComplianceService) DownloadReport(ctx context.Context, orgID, reportID string) (*ComplianceReportDownloadOutput, error) {
	result, err := Do[ComplianceReportDownloadOutput](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/compliance/reports/%s/download", orgID, reportID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDashboard retrieves the compliance dashboard with summary statistics across all frameworks.
func (s *ComplianceService) GetDashboard(ctx context.Context, orgID string) (*ComplianceDashboard, error) {
	dashboard, err := Do[ComplianceDashboard](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/compliance/dashboard", orgID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &dashboard, nil
}

// CreateDSAR creates a new Data Subject Access Request.
func (s *ComplianceService) CreateDSAR(ctx context.Context, orgID string, input CreateDSARInput) (*DataSubjectRequest, error) {
	dsar, err := Do[DataSubjectRequest](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/compliance/dsars", orgID), input, nil)
	if err != nil {
		return nil, err
	}
	return &dsar, nil
}

// ListDSARs returns a paginated list of Data Subject Access Requests.
func (s *ComplianceService) ListDSARs(ctx context.Context, orgID string, params *DSARListParams) (*Page[DataSubjectRequest], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[DataSubjectRequest]](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/compliance/dsars", orgID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListDSARsAutoPaging returns an iterator that automatically paginates through all DSARs.
func (s *ComplianceService) ListDSARsAutoPaging(orgID string, params *DSARListParams) *ListIterator[DataSubjectRequest] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[DataSubjectRequest], error) {
		p := &DSARListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListDSARs(ctx, orgID, p)
	})
}

// CompleteDSAR marks a Data Subject Access Request as completed.
func (s *ComplianceService) CompleteDSAR(ctx context.Context, orgID, dsarID string, input *CompleteDSARInput) (*DataSubjectRequest, error) {
	dsar, err := Do[DataSubjectRequest](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/compliance/dsars/%s/complete", orgID, dsarID), input, nil)
	if err != nil {
		return nil, err
	}
	return &dsar, nil
}
