package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AnomalyMetric represents a behavioral metric tracked for anomaly detection.
type AnomalyMetric string

const (
	AnomalyMetricEmailSendRate    AnomalyMetric = "email_send_rate"
	AnomalyMetricSMSSendRate      AnomalyMetric = "sms_send_rate"
	AnomalyMetricCardTxnCount     AnomalyMetric = "card_txn_count"
	AnomalyMetricVaultAccessRate  AnomalyMetric = "vault_access_rate"
	AnomalyMetricAPICallRate      AnomalyMetric = "api_call_rate"
	AnomalyMetricUniqueRecipients AnomalyMetric = "unique_recipients"
)

// AnomalySeverity represents the severity of an anomaly alert.
type AnomalySeverity string

const (
	AnomalySeverityInfo     AnomalySeverity = "INFO"
	AnomalySeverityWarning  AnomalySeverity = "WARNING"
	AnomalySeverityCritical AnomalySeverity = "CRITICAL"
)

// AnomalyAlertStatus represents the status of an anomaly alert.
type AnomalyAlertStatus string

const (
	AnomalyAlertStatusTriggered     AnomalyAlertStatus = "TRIGGERED"
	AnomalyAlertStatusAcknowledged  AnomalyAlertStatus = "ACKNOWLEDGED"
	AnomalyAlertStatusResolved      AnomalyAlertStatus = "RESOLVED"
	AnomalyAlertStatusFalsePositive AnomalyAlertStatus = "FALSE_POSITIVE"
)

// AnomalyCondition represents the condition type for an anomaly detection rule.
type AnomalyCondition string

const (
	AnomalyConditionZScoreGT         AnomalyCondition = "zscore_gt"
	AnomalyConditionRateMultiplierGT AnomalyCondition = "rate_multiplier_gt"
	AnomalyConditionAbsoluteGT       AnomalyCondition = "absolute_gt"
	AnomalyConditionTimeViolation    AnomalyCondition = "time_violation"
)

// QuarantineAction represents the quarantine action taken when a rule triggers.
type QuarantineAction string

const (
	QuarantineActionNone QuarantineAction = "NONE"
	QuarantineActionSoft QuarantineAction = "SOFT"
	QuarantineActionHard QuarantineAction = "HARD"
)

// QuarantineLevel represents the current quarantine level of an agent.
type QuarantineLevel string

const (
	QuarantineLevelNone QuarantineLevel = "NONE"
	QuarantineLevelSoft QuarantineLevel = "SOFT"
	QuarantineLevelHard QuarantineLevel = "HARD"
)

// BaselinePeriod represents the time period for a behavioral baseline.
type BaselinePeriod string

const (
	BaselinePeriodHourly BaselinePeriod = "hourly"
	BaselinePeriodDaily  BaselinePeriod = "daily"
)

// AnomalyAlert represents an anomaly detection alert.
type AnomalyAlert struct {
	ID             string                 `json:"id"`
	OrgID          string                 `json:"orgId"`
	AgentID        string                 `json:"agentId"`
	Metric         AnomalyMetric          `json:"metric"`
	Severity       AnomalySeverity        `json:"severity"`
	Status         AnomalyAlertStatus     `json:"status"`
	BaselineValue  float64                `json:"baselineValue"`
	ActualValue    float64                `json:"actualValue"`
	ZScore         float64                `json:"zScore"`
	RuleID         *string                `json:"ruleId"`
	Details        map[string]interface{} `json:"details"`
	AcknowledgedBy *string                `json:"acknowledgedBy"`
	AcknowledgedAt *string                `json:"acknowledgedAt"`
	ResolvedBy     *string                `json:"resolvedBy"`
	ResolvedAt     *string                `json:"resolvedAt"`
	CreatedAt      string                 `json:"createdAt"`
}

// AnomalyAlertListParams contains parameters for listing anomaly alerts.
type AnomalyAlertListParams struct {
	ListParams
	AgentID  string
	Metric   AnomalyMetric
	Severity AnomalySeverity
	Status   AnomalyAlertStatus
}

// ToQuery converts AnomalyAlertListParams to URL query values.
func (p AnomalyAlertListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.Metric != "" {
		q.Set("metric", string(p.Metric))
	}
	if p.Severity != "" {
		q.Set("severity", string(p.Severity))
	}
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	return q
}

// AnomalyRule represents a rule for anomaly detection.
type AnomalyRule struct {
	ID               string           `json:"id"`
	OrgID            string           `json:"orgId"`
	Name             string           `json:"name"`
	Metric           AnomalyMetric    `json:"metric"`
	Condition        AnomalyCondition `json:"condition"`
	Threshold        float64          `json:"threshold"`
	Severity         AnomalySeverity  `json:"severity"`
	QuarantineAction QuarantineAction `json:"quarantineAction"`
	CooldownMinutes  int              `json:"cooldownMinutes"`
	Enabled          bool             `json:"enabled"`
	CreatedAt        string           `json:"createdAt"`
	UpdatedAt        string           `json:"updatedAt"`
}

// AnomalyRuleListParams contains parameters for listing anomaly rules.
type AnomalyRuleListParams struct {
	ListParams
	Metric  AnomalyMetric
	Enabled *bool
}

// ToQuery converts AnomalyRuleListParams to URL query values.
func (p AnomalyRuleListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Metric != "" {
		q.Set("metric", string(p.Metric))
	}
	if p.Enabled != nil {
		q.Set("enabled", strconv.FormatBool(*p.Enabled))
	}
	return q
}

// CreateAnomalyRuleInput contains parameters for creating an anomaly detection rule.
type CreateAnomalyRuleInput struct {
	Name             string           `json:"name"`
	Metric           AnomalyMetric    `json:"metric"`
	Condition        AnomalyCondition `json:"condition"`
	Threshold        float64          `json:"threshold"`
	Severity         AnomalySeverity  `json:"severity"`
	QuarantineAction QuarantineAction `json:"quarantineAction,omitempty"`
	CooldownMinutes  *int             `json:"cooldownMinutes,omitempty"`
	Enabled          *bool            `json:"enabled,omitempty"`
}

// UpdateAnomalyRuleInput contains parameters for updating an anomaly detection rule.
type UpdateAnomalyRuleInput struct {
	Name             string           `json:"name,omitempty"`
	Threshold        *float64         `json:"threshold,omitempty"`
	Severity         AnomalySeverity  `json:"severity,omitempty"`
	QuarantineAction QuarantineAction `json:"quarantineAction,omitempty"`
	CooldownMinutes  *int             `json:"cooldownMinutes,omitempty"`
	Enabled          *bool            `json:"enabled,omitempty"`
}

// BaselineMetric represents a single behavioral metric baseline.
type BaselineMetric struct {
	Metric        AnomalyMetric      `json:"metric"`
	Period        BaselinePeriod     `json:"period"`
	Mean          float64            `json:"mean"`
	Stddev        float64            `json:"stddev"`
	SampleCount   int                `json:"sampleCount"`
	HourlyPattern map[string]float64 `json:"hourlyPattern"`
	WindowStart   string             `json:"windowStart"`
	WindowEnd     string             `json:"windowEnd"`
}

// AgentBaseline represents the behavioral baselines for an agent.
type AgentBaseline struct {
	AgentID string           `json:"agentId"`
	OrgID   string           `json:"orgId"`
	Metrics []BaselineMetric `json:"metrics"`
}

// QuarantineInput contains parameters for quarantining an agent.
type QuarantineInput struct {
	Level  QuarantineLevel `json:"level"`
	Reason string          `json:"reason,omitempty"`
}

// QuarantineOutput contains the result of a quarantine or release operation.
type QuarantineOutput struct {
	AgentID         string          `json:"agentId"`
	QuarantineLevel QuarantineLevel `json:"quarantineLevel"`
	QuarantinedAt   *string         `json:"quarantinedAt"`
	Reason          *string         `json:"reason"`
}

// AnomalyService provides methods for managing anomaly detection alerts, rules,
// baselines, and agent quarantine.
type AnomalyService struct {
	client *httpClient
}

// newAnomalyService creates a new AnomalyService.
func newAnomalyService(c *httpClient) *AnomalyService {
	return &AnomalyService{client: c}
}

// ListAlerts returns a paginated list of anomaly alerts for an organization.
func (s *AnomalyService) ListAlerts(ctx context.Context, orgID string, params *AnomalyAlertListParams) (*Page[AnomalyAlert], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[AnomalyAlert]](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/anomaly/alerts", orgID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAlertsAutoPaging returns an iterator that automatically paginates through all anomaly alerts.
func (s *AnomalyService) ListAlertsAutoPaging(orgID string, params *AnomalyAlertListParams) *ListIterator[AnomalyAlert] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[AnomalyAlert], error) {
		p := &AnomalyAlertListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListAlerts(ctx, orgID, p)
	})
}

// AcknowledgeAlert acknowledges an anomaly alert.
func (s *AnomalyService) AcknowledgeAlert(ctx context.Context, orgID, alertID string) (*AnomalyAlert, error) {
	alert, err := Do[AnomalyAlert](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/anomaly/alerts/%s/acknowledge", orgID, alertID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// ResolveAlert resolves an anomaly alert.
func (s *AnomalyService) ResolveAlert(ctx context.Context, orgID, alertID string) (*AnomalyAlert, error) {
	alert, err := Do[AnomalyAlert](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/anomaly/alerts/%s/resolve", orgID, alertID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// MarkFalsePositive marks an anomaly alert as a false positive.
func (s *AnomalyService) MarkFalsePositive(ctx context.Context, orgID, alertID string) (*AnomalyAlert, error) {
	alert, err := Do[AnomalyAlert](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/anomaly/alerts/%s/false-positive", orgID, alertID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// ListRules returns a paginated list of anomaly detection rules for an organization.
func (s *AnomalyService) ListRules(ctx context.Context, orgID string, params *AnomalyRuleListParams) (*Page[AnomalyRule], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[AnomalyRule]](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/anomaly/rules", orgID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListRulesAutoPaging returns an iterator that automatically paginates through all anomaly rules.
func (s *AnomalyService) ListRulesAutoPaging(orgID string, params *AnomalyRuleListParams) *ListIterator[AnomalyRule] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[AnomalyRule], error) {
		p := &AnomalyRuleListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListRules(ctx, orgID, p)
	})
}

// CreateRule creates a new anomaly detection rule.
func (s *AnomalyService) CreateRule(ctx context.Context, orgID string, input CreateAnomalyRuleInput) (*AnomalyRule, error) {
	rule, err := Do[AnomalyRule](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/anomaly/rules", orgID), input, nil)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateRule updates an existing anomaly detection rule.
func (s *AnomalyService) UpdateRule(ctx context.Context, orgID, ruleID string, input UpdateAnomalyRuleInput) (*AnomalyRule, error) {
	rule, err := Do[AnomalyRule](ctx, s.client, http.MethodPatch, fmt.Sprintf("/v1/orgs/%s/anomaly/rules/%s", orgID, ruleID), input, nil)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// DeleteRule deletes an anomaly detection rule.
func (s *AnomalyService) DeleteRule(ctx context.Context, orgID, ruleID string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/v1/orgs/%s/anomaly/rules/%s", orgID, ruleID), nil, nil)
	return err
}

// GetBaseline retrieves the behavioral baselines for an agent.
func (s *AnomalyService) GetBaseline(ctx context.Context, orgID, agentID string) (*AgentBaseline, error) {
	baseline, err := Do[AgentBaseline](ctx, s.client, http.MethodGet, fmt.Sprintf("/v1/orgs/%s/anomaly/baselines/%s", orgID, agentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &baseline, nil
}

// Quarantine sets the quarantine level for an agent.
func (s *AnomalyService) Quarantine(ctx context.Context, orgID, agentID string, input QuarantineInput) (*QuarantineOutput, error) {
	result, err := Do[QuarantineOutput](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/anomaly/quarantine/%s", orgID, agentID), input, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ReleaseQuarantine releases an agent from quarantine.
func (s *AnomalyService) ReleaseQuarantine(ctx context.Context, orgID, agentID string) (*QuarantineOutput, error) {
	result, err := Do[QuarantineOutput](ctx, s.client, http.MethodPost, fmt.Sprintf("/v1/orgs/%s/anomaly/quarantine/%s/release", orgID, agentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
