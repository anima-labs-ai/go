package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Every enum in this file is SCREAMING_SNAKE because that is what the API
// validates against — see packages/contracts/src/schemas/compliance.ts and
// compliance-controls.ts in the anima monorepo, at the commit pinned in
// .anima-ref. They were lowercase, which meant every compliance request this
// SDK could build was rejected. compliance_test.go pins them.

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
	ComplianceControlStatusNotStarted  ComplianceControlStatus = "NOT_STARTED"
	ComplianceControlStatusInProgress  ComplianceControlStatus = "IN_PROGRESS"
	ComplianceControlStatusImplemented ComplianceControlStatus = "IMPLEMENTED"
	ComplianceControlStatusVerified    ComplianceControlStatus = "VERIFIED"
	ComplianceControlStatusFailed      ComplianceControlStatus = "FAILED"
)

// ComplianceControlCategory is a Trust Service Criteria category.
type ComplianceControlCategory string

const (
	ComplianceControlCategoryCC1 ComplianceControlCategory = "CC1"
	ComplianceControlCategoryCC2 ComplianceControlCategory = "CC2"
	ComplianceControlCategoryCC3 ComplianceControlCategory = "CC3"
	ComplianceControlCategoryCC4 ComplianceControlCategory = "CC4"
	ComplianceControlCategoryCC5 ComplianceControlCategory = "CC5"
	ComplianceControlCategoryCC6 ComplianceControlCategory = "CC6"
	ComplianceControlCategoryCC7 ComplianceControlCategory = "CC7"
	ComplianceControlCategoryCC8 ComplianceControlCategory = "CC8"
	ComplianceControlCategoryCC9 ComplianceControlCategory = "CC9"
	ComplianceControlCategoryA1  ComplianceControlCategory = "A1"
	ComplianceControlCategoryPI1 ComplianceControlCategory = "PI1"
	ComplianceControlCategoryC1  ComplianceControlCategory = "C1"
	ComplianceControlCategoryP1  ComplianceControlCategory = "P1"
)

// ComplianceReportType represents the type of a compliance report.
type ComplianceReportType string

const (
	ComplianceReportTypeSOC2Summary    ComplianceReportType = "SOC2_SUMMARY"
	ComplianceReportTypeActivityReport ComplianceReportType = "ACTIVITY_REPORT"
	ComplianceReportTypeAccessReview   ComplianceReportType = "ACCESS_REVIEW"
	ComplianceReportTypeAuditExport    ComplianceReportType = "AUDIT_EXPORT"
	ComplianceReportTypeGDPRDSAR       ComplianceReportType = "GDPR_DSAR"
)

// ComplianceReportStatus represents the generation status of a report.
type ComplianceReportStatus string

const (
	ComplianceReportStatusPending    ComplianceReportStatus = "PENDING"
	ComplianceReportStatusGenerating ComplianceReportStatus = "GENERATING"
	ComplianceReportStatusCompleted  ComplianceReportStatus = "COMPLETED"
	ComplianceReportStatusFailed     ComplianceReportStatus = "FAILED"
)

// ComplianceReportFormat represents the export format of a report.
type ComplianceReportFormat string

const (
	ComplianceReportFormatJSON ComplianceReportFormat = "JSON"
	ComplianceReportFormatCSV  ComplianceReportFormat = "CSV"
	ComplianceReportFormatPDF  ComplianceReportFormat = "PDF"
)

// DSARStatus represents the status of a Data Subject Access Request.
type DSARStatus string

const (
	DSARStatusReceived   DSARStatus = "RECEIVED"
	DSARStatusVerified   DSARStatus = "VERIFIED"
	DSARStatusInProgress DSARStatus = "IN_PROGRESS"
	DSARStatusCompleted  DSARStatus = "COMPLETED"
	DSARStatusDenied     DSARStatus = "DENIED"
	DSARStatusOverdue    DSARStatus = "OVERDUE"
)

// DSARType represents the kind of data-subject request, keyed `type` on the
// wire. This was DSARRequestType, sent as `requestType`, which the API rejects.
type DSARType string

const (
	DSARTypeAccess      DSARType = "ACCESS"
	DSARTypeDelete      DSARType = "DELETE"
	DSARTypeRectify     DSARType = "RECTIFY"
	DSARTypePortability DSARType = "PORTABILITY"
	DSARTypeRestrict    DSARType = "RESTRICT"
)

// ComplianceControl represents a compliance control within a framework.
type ComplianceControl struct {
	ID           string                    `json:"id"`
	OrgID        string                    `json:"orgId"`
	Framework    ComplianceFramework       `json:"framework"`
	ControlID    string                    `json:"controlId"`
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	Category     ComplianceControlCategory `json:"category"`
	Status       ComplianceControlStatus   `json:"status"`
	Owner        *string                   `json:"owner"`
	LastTestedAt *string                   `json:"lastTestedAt"`
	NextReviewAt *string                   `json:"nextReviewAt"`
	CreatedAt    string                    `json:"createdAt"`
	UpdatedAt    string                    `json:"updatedAt"`
}

// ComplianceControlListParams contains parameters for listing compliance controls.
type ComplianceControlListParams struct {
	ListParams
	Framework ComplianceFramework
	Category  ComplianceControlCategory
	Status    ComplianceControlStatus
}

// ToQuery converts ComplianceControlListParams to URL query values.
func (p ComplianceControlListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Framework != "" {
		q.Set("framework", string(p.Framework))
	}
	if p.Category != "" {
		q.Set("category", string(p.Category))
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
	Type        ComplianceReportType   `json:"type"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Format      ComplianceReportFormat `json:"format,omitempty"`
	GeneratedBy string                 `json:"generatedBy,omitempty"`
	PeriodStart string                 `json:"periodStart,omitempty"`
	PeriodEnd   string                 `json:"periodEnd,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ComplianceReport represents a generated compliance report.
type ComplianceReport struct {
	ID           string                 `json:"id"`
	OrgID        string                 `json:"orgId"`
	Type         ComplianceReportType   `json:"type"`
	Title        string                 `json:"title"`
	Description  *string                `json:"description"`
	Status       ComplianceReportStatus `json:"status"`
	Format       ComplianceReportFormat `json:"format"`
	Parameters   map[string]interface{} `json:"parameters"`
	Content      map[string]interface{} `json:"content"`
	ErrorMessage *string                `json:"errorMessage"`
	GeneratedBy  *string                `json:"generatedBy"`
	PeriodStart  *string                `json:"periodStart"`
	PeriodEnd    *string                `json:"periodEnd"`
	CompletedAt  *string                `json:"completedAt"`
	CreatedAt    string                 `json:"createdAt"`
	UpdatedAt    string                 `json:"updatedAt"`
}

// ComplianceReportListParams contains parameters for listing compliance reports.
type ComplianceReportListParams struct {
	ListParams
	Type   ComplianceReportType
	Status ComplianceReportStatus
}

// ToQuery converts ComplianceReportListParams to URL query values.
func (p ComplianceReportListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Type != "" {
		q.Set("type", string(p.Type))
	}
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	return q
}

// ExportReportInput optionally overrides the report's stored format.
type ExportReportInput struct {
	Format ComplianceReportFormat `json:"format,omitempty"`
}

// ExportReportOutput carries an exported report. The bytes are returned inline;
// there is no signed download URL.
type ExportReportOutput struct {
	Data        string `json:"data"`
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
}

// ComplianceTemplate describes one available report template.
type ComplianceTemplate struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ListTemplatesOutput is the report-template catalogue.
type ListTemplatesOutput struct {
	Items []ComplianceTemplate `json:"items"`
}

// DashboardReportSummary is one report row on the dashboard.
type DashboardReportSummary struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"createdAt"`
	CompletedAt *string `json:"completedAt"`
}

// DashboardDSARSummary is one DSAR row on the dashboard.
type DashboardDSARSummary struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	SubjectEmail string `json:"subjectEmail"`
	DueAt        string `json:"dueAt"`
	CreatedAt    string `json:"createdAt"`
}

// DashboardReportsSection summarises generated reports.
type DashboardReportsSection struct {
	Total         int                      `json:"total"`
	ByType        map[string]int           `json:"byType"`
	ByStatus      map[string]int           `json:"byStatus"`
	RecentReports []DashboardReportSummary `json:"recentReports"`
}

// DashboardDSARsSection summarises data-subject requests.
type DashboardDSARsSection struct {
	Total                 int                    `json:"total"`
	ByStatus              map[string]int         `json:"byStatus"`
	ByType                map[string]int         `json:"byType"`
	Overdue               int                    `json:"overdue"`
	AverageResolutionDays *float64               `json:"averageResolutionDays"`
	RecentRequests        []DashboardDSARSummary `json:"recentRequests"`
}

// ComplianceFrameworkSummary contains control progress for a single framework.
type ComplianceFrameworkSummary struct {
	Framework        string `json:"framework"`
	TotalControls    int    `json:"totalControls"`
	ImplementedCount int    `json:"implementedCount"`
	Progress         int    `json:"progress"`
}

// DashboardComplianceSection summarises control implementation progress.
type DashboardComplianceSection struct {
	OverallProgress    int                          `json:"overallProgress"`
	FrameworkSummaries []ComplianceFrameworkSummary `json:"frameworkSummaries"`
}

// ComplianceDashboard is the compliance overview for an organization.
type ComplianceDashboard struct {
	Reports    DashboardReportsSection    `json:"reports"`
	DSARs      DashboardDSARsSection      `json:"dsars"`
	Compliance DashboardComplianceSection `json:"compliance"`
}

// CreateDSARInput contains parameters for creating a Data Subject Access Request.
type CreateDSARInput struct {
	// Type is keyed `type` on the wire. It used to be sent as `requestType`,
	// which the API rejects.
	Type         DSARType `json:"type"`
	SubjectEmail string   `json:"subjectEmail"`
	SubjectName  string   `json:"subjectName,omitempty"`
	SubjectID    string   `json:"subjectId,omitempty"`
	Description  string   `json:"description,omitempty"`
	// DueInDays accepts 1-90 and defaults to 30, the GDPR response window.
	DueInDays int                    `json:"dueInDays,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DataSubjectRequest represents a GDPR Data Subject Access Request.
type DataSubjectRequest struct {
	ID           string                 `json:"id"`
	OrgID        string                 `json:"orgId"`
	Type         DSARType               `json:"type"`
	Status       DSARStatus             `json:"status"`
	SubjectEmail string                 `json:"subjectEmail"`
	SubjectName  *string                `json:"subjectName"`
	SubjectID    *string                `json:"subjectId"`
	Description  *string                `json:"description"`
	RequestedAt  string                 `json:"requestedAt"`
	VerifiedAt   *string                `json:"verifiedAt"`
	DueAt        string                 `json:"dueAt"`
	CompletedAt  *string                `json:"completedAt"`
	ProcessedBy  *string                `json:"processedBy"`
	Response     map[string]interface{} `json:"response"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    string                 `json:"createdAt"`
	UpdatedAt    string                 `json:"updatedAt"`
}

// DSARListParams contains parameters for listing DSARs.
type DSARListParams struct {
	ListParams
	Status DSARStatus
	Type   DSARType
}

// ToQuery converts DSARListParams to URL query values.
func (p DSARListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	if p.Type != "" {
		q.Set("type", string(p.Type))
	}
	return q
}

// UpdateDSARStatusInput moves a DSAR along its lifecycle.
type UpdateDSARStatusInput struct {
	Status      DSARStatus             `json:"status"`
	ProcessedBy string                 `json:"processedBy,omitempty"`
	Response    map[string]interface{} `json:"response,omitempty"`
}

// ComplianceService provides methods for managing compliance controls, reports,
// dashboards, and Data Subject Access Requests (DSARs).
//
// Every route is org-scoped and requires a master key (mk_*).
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
	page, err := Do[Page[ComplianceControl]](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/compliance/controls", orgID), nil, q)
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
	ctrl, err := Do[ComplianceControl](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/compliance/controls/%s", orgID, controlID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &ctrl, nil
}

// UpdateControlStatus updates the status and optionally the owner of a compliance control.
func (s *ComplianceService) UpdateControlStatus(ctx context.Context, orgID, controlID string, input ComplianceControlStatusInput) (*ComplianceControl, error) {
	ctrl, err := Do[ComplianceControl](ctx, s.client, http.MethodPatch, fmt.Sprintf("/orgs/%s/compliance/controls/%s", orgID, controlID), input, nil)
	if err != nil {
		return nil, err
	}
	return &ctrl, nil
}

// SeedFramework seeds all predefined controls for a compliance framework (e.g. SOC2).
func (s *ComplianceService) SeedFramework(ctx context.Context, orgID string, input SeedFrameworkInput) (*SeedFrameworkOutput, error) {
	result, err := Do[SeedFrameworkOutput](ctx, s.client, http.MethodPost, fmt.Sprintf("/orgs/%s/compliance/seed", orgID), input, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListTemplates returns the available report templates.
func (s *ComplianceService) ListTemplates(ctx context.Context, orgID string) (*ListTemplatesOutput, error) {
	result, err := Do[ListTemplatesOutput](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/compliance/templates", orgID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateReport generates a new compliance report.
func (s *ComplianceService) GenerateReport(ctx context.Context, orgID string, input GenerateReportInput) (*ComplianceReport, error) {
	report, err := Do[ComplianceReport](ctx, s.client, http.MethodPost, fmt.Sprintf("/orgs/%s/compliance/reports", orgID), input, nil)
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
	page, err := Do[Page[ComplianceReport]](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/compliance/reports", orgID), nil, q)
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
	report, err := Do[ComplianceReport](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/compliance/reports/%s", orgID, reportID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// ExportReport exports a generated report. The bytes come back inline as
// Data/ContentType/Filename.
//
// Replaces DownloadReport, which issued a GET to a /download sub-path that the
// API does not serve.
func (s *ComplianceService) ExportReport(ctx context.Context, orgID, reportID string, input *ExportReportInput) (*ExportReportOutput, error) {
	body := ExportReportInput{}
	if input != nil {
		body = *input
	}
	result, err := Do[ExportReportOutput](ctx, s.client, http.MethodPost, fmt.Sprintf("/orgs/%s/compliance/reports/%s/export", orgID, reportID), body, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteReport deletes a generated compliance report.
func (s *ComplianceService) DeleteReport(ctx context.Context, orgID, reportID string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/orgs/%s/compliance/reports/%s", orgID, reportID), nil, nil)
	return err
}

// GetDashboard retrieves the compliance dashboard.
func (s *ComplianceService) GetDashboard(ctx context.Context, orgID string) (*ComplianceDashboard, error) {
	dashboard, err := Do[ComplianceDashboard](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/compliance/dashboard", orgID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &dashboard, nil
}

// CreateDSAR creates a new Data Subject Access Request.
func (s *ComplianceService) CreateDSAR(ctx context.Context, orgID string, input CreateDSARInput) (*DataSubjectRequest, error) {
	dsar, err := Do[DataSubjectRequest](ctx, s.client, http.MethodPost, fmt.Sprintf("/orgs/%s/compliance/dsars", orgID), input, nil)
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
	page, err := Do[Page[DataSubjectRequest]](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/compliance/dsars", orgID), nil, q)
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

// GetDSAR retrieves a single Data Subject Access Request by ID.
func (s *ComplianceService) GetDSAR(ctx context.Context, orgID, dsarID string) (*DataSubjectRequest, error) {
	dsar, err := Do[DataSubjectRequest](ctx, s.client, http.MethodGet, fmt.Sprintf("/orgs/%s/compliance/dsars/%s", orgID, dsarID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &dsar, nil
}

// UpdateDSARStatus moves a Data Subject Access Request along its lifecycle.
//
// Replaces CompleteDSAR, which POSTed to a /complete sub-path that does not
// exist. The API models this as a PATCH carrying the new status.
func (s *ComplianceService) UpdateDSARStatus(ctx context.Context, orgID, dsarID string, input UpdateDSARStatusInput) (*DataSubjectRequest, error) {
	dsar, err := Do[DataSubjectRequest](ctx, s.client, http.MethodPatch, fmt.Sprintf("/orgs/%s/compliance/dsars/%s", orgID, dsarID), input, nil)
	if err != nil {
		return nil, err
	}
	return &dsar, nil
}
