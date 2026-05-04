package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CardType represents the type of a card.
type CardType string

const (
	CardTypeVirtual  CardType = "VIRTUAL"
	CardTypePhysical CardType = "PHYSICAL"
)

// CardStatus represents the status of a card.
type CardStatus string

const (
	CardStatusActive   CardStatus = "ACTIVE"
	CardStatusFrozen   CardStatus = "FROZEN"
	CardStatusCanceled CardStatus = "CANCELED"
)

// Card represents a virtual or physical card in the Anima platform.
type Card struct {
	ID                 string     `json:"id"`
	AgentID            string     `json:"agentId"`
	OrgID              string     `json:"orgId"`
	ProviderCardID     string     `json:"providerCardId"`
	CardType           CardType   `json:"cardType"`
	Status             CardStatus `json:"status"`
	Last4              string     `json:"last4"`
	Brand              string     `json:"brand"`
	ExpMonth           int        `json:"expMonth"`
	ExpYear            int        `json:"expYear"`
	Currency           string     `json:"currency"`
	Label              *string    `json:"label"`
	SpendLimitDaily    *int       `json:"spendLimitDaily"`
	SpendLimitMonthly  *int       `json:"spendLimitMonthly"`
	SpendLimitPerAuth  *int       `json:"spendLimitPerAuth"`
	SpentToday         int        `json:"spentToday"`
	SpentThisMonth     int        `json:"spentThisMonth"`
	KillSwitchActive   bool       `json:"killSwitchActive"`
	CreatedAt          string     `json:"createdAt"`
	UpdatedAt          string     `json:"updatedAt"`
}

// CreateCardParams contains the parameters for creating a card.
type CreateCardParams struct {
	AgentID           string            `json:"agentId"`
	Label             string            `json:"label,omitempty"`
	Currency          string            `json:"currency,omitempty"`
	SpendLimitDaily   *int              `json:"spendLimitDaily,omitempty"`
	SpendLimitMonthly *int              `json:"spendLimitMonthly,omitempty"`
	SpendLimitPerAuth *int              `json:"spendLimitPerAuth,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// UpdateCardParams contains the parameters for updating a card.
type UpdateCardParams struct {
	Label             string     `json:"label,omitempty"`
	Status            CardStatus `json:"status,omitempty"`
	SpendLimitDaily   *int       `json:"spendLimitDaily,omitempty"`
	SpendLimitMonthly *int       `json:"spendLimitMonthly,omitempty"`
	SpendLimitPerAuth *int       `json:"spendLimitPerAuth,omitempty"`
}

// CardListParams contains parameters for listing cards.
type CardListParams struct {
	ListParams
	AgentID string
	Status  CardStatus
}

// ToQuery converts CardListParams to URL query values.
func (p CardListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	return q
}

// PolicyAction is the action a spending policy takes.
type PolicyAction string

const (
	PolicyActionAutoApprove      PolicyAction = "AUTO_APPROVE"
	PolicyActionRequireApproval  PolicyAction = "REQUIRE_APPROVAL"
	PolicyActionAlwaysDecline    PolicyAction = "ALWAYS_DECLINE"
)

// SpendingPolicy represents a spending policy for a card.
type SpendingPolicy struct {
	ID                string       `json:"id"`
	CardID            string       `json:"cardId"`
	OrgID             string       `json:"orgId"`
	Name              string       `json:"name"`
	Priority          int          `json:"priority"`
	Action            PolicyAction `json:"action"`
	MaxAmountCents    *int         `json:"maxAmountCents"`
	MinAmountCents    *int         `json:"minAmountCents"`
	AllowedCategories []string     `json:"allowedCategories"`
	BlockedCategories []string     `json:"blockedCategories"`
	AllowedMerchants  []string     `json:"allowedMerchants"`
	BlockedMerchants  []string     `json:"blockedMerchants"`
	AllowedCountries  []string     `json:"allowedCountries"`
	BlockedCountries  []string     `json:"blockedCountries"`
	CreatedAt         string       `json:"createdAt"`
}

// CreatePolicyParams contains the parameters for creating a spending policy.
type CreatePolicyParams struct {
	CardID            string       `json:"cardId"`
	Name              string       `json:"name"`
	Priority          int          `json:"priority,omitempty"`
	Action            PolicyAction `json:"action"`
	MaxAmountCents    *int         `json:"maxAmountCents,omitempty"`
	MinAmountCents    *int         `json:"minAmountCents,omitempty"`
	AllowedCategories []string     `json:"allowedCategories,omitempty"`
	BlockedCategories []string     `json:"blockedCategories,omitempty"`
	AllowedMerchants  []string     `json:"allowedMerchants,omitempty"`
	BlockedMerchants  []string     `json:"blockedMerchants,omitempty"`
	AllowedCountries  []string     `json:"allowedCountries,omitempty"`
	BlockedCountries  []string     `json:"blockedCountries,omitempty"`
}

// UpdatePolicyParams contains the parameters for updating a spending policy.
type UpdatePolicyParams struct {
	Name              string       `json:"name,omitempty"`
	Priority          int          `json:"priority,omitempty"`
	Action            PolicyAction `json:"action,omitempty"`
	MaxAmountCents    *int         `json:"maxAmountCents,omitempty"`
	MinAmountCents    *int         `json:"minAmountCents,omitempty"`
	AllowedCategories []string     `json:"allowedCategories,omitempty"`
	BlockedCategories []string     `json:"blockedCategories,omitempty"`
	AllowedMerchants  []string     `json:"allowedMerchants,omitempty"`
	BlockedMerchants  []string     `json:"blockedMerchants,omitempty"`
	AllowedCountries  []string     `json:"allowedCountries,omitempty"`
	BlockedCountries  []string     `json:"blockedCountries,omitempty"`
}

// TransactionStatus represents the status of a card transaction.
type TransactionStatus string

const (
	TransactionStatusPending  TransactionStatus = "PENDING"
	TransactionStatusApproved TransactionStatus = "APPROVED"
	TransactionStatusDeclined TransactionStatus = "DECLINED"
	TransactionStatusReversed TransactionStatus = "REVERSED"
	TransactionStatusExpired  TransactionStatus = "EXPIRED"
)

// CardTransaction represents a card transaction.
type CardTransaction struct {
	ID                   string            `json:"id"`
	CardID               string            `json:"cardId"`
	Status               TransactionStatus `json:"status"`
	Decision             *string           `json:"decision"`
	AmountCents          int               `json:"amountCents"`
	Currency             string            `json:"currency"`
	MerchantName         *string           `json:"merchantName"`
	MerchantCategory     *string           `json:"merchantCategory"`
	MerchantCategoryCode *string           `json:"merchantCategoryCode"`
	CreatedAt            string            `json:"createdAt"`
}

// TransactionListParams contains parameters for listing transactions.
type TransactionListParams struct {
	ListParams
	CardID  string
	AgentID string
	Status  string
}

// ToQuery converts TransactionListParams to URL query values.
func (p TransactionListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.CardID != "" {
		q.Set("cardId", p.CardID)
	}
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	return q
}

// KillSwitchParams contains the parameters for the kill switch.
type KillSwitchParams struct {
	AgentID string `json:"agentId,omitempty"`
	CardID  string `json:"cardId,omitempty"`
	Active  bool   `json:"active"`
}

// KillSwitchResult contains the result of a kill switch operation.
type KillSwitchResult struct {
	Affected int  `json:"affected"`
	Active   bool `json:"active"`
}

// ApprovalStatus represents the status of a card approval.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusDeclined ApprovalStatus = "DECLINED"
	ApprovalStatusExpired  ApprovalStatus = "EXPIRED"
)

// CardApproval represents a pending card approval.
type CardApproval struct {
	ID          string         `json:"id"`
	OrgID       string         `json:"orgId"`
	CardID      string         `json:"cardId"`
	AmountCents int            `json:"amountCents"`
	Currency    string         `json:"currency"`
	MerchantName *string       `json:"merchantName"`
	Status      ApprovalStatus `json:"status"`
	DecidedBy   *string        `json:"decidedBy"`
	ExpiresAt   string         `json:"expiresAt"`
	CreatedAt   string         `json:"createdAt"`
}

// ApprovalListParams contains parameters for listing approvals.
type ApprovalListParams struct {
	ListParams
	Status ApprovalStatus
}

// ToQuery converts ApprovalListParams to URL query values.
func (p ApprovalListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	return q
}

// ApprovalDecision is the decision to approve or decline a card transaction.
type ApprovalDecision string

const (
	ApprovalDecisionApproved ApprovalDecision = "APPROVED"
	ApprovalDecisionDeclined ApprovalDecision = "DECLINED"
)

// policyListEnvelope wraps a list or items response from the policies endpoint.
type policyListEnvelope struct {
	Items []SpendingPolicy `json:"items"`
}

// CardList wraps a list of cards with a cursor.
type CardList struct {
	Items  []Card  `json:"items"`
	Cursor *string `json:"cursor,omitempty"`
}

// TransactionList wraps a list of transactions with a cursor.
type TransactionList struct {
	Items  []CardTransaction `json:"items"`
	Cursor *string           `json:"cursor,omitempty"`
}

// ApprovalList wraps a list of approvals with a cursor.
type ApprovalList struct {
	Items  []CardApproval `json:"items"`
	Cursor *string        `json:"cursor,omitempty"`
}

// CardsService provides methods for managing cards, spending policies,
// transactions, and approvals.
type CardsService struct {
	client *httpClient
}

// newCardsService creates a new CardsService.
func newCardsService(c *httpClient) *CardsService {
	return &CardsService{client: c}
}

// Create creates a new virtual card.
func (s *CardsService) Create(ctx context.Context, params CreateCardParams) (*Card, error) {
	card, err := Do[Card](ctx, s.client, http.MethodPost, "/cards", params, nil)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// Get retrieves a card by ID.
func (s *CardsService) Get(ctx context.Context, id string) (*Card, error) {
	card, err := Do[Card](ctx, s.client, http.MethodGet, fmt.Sprintf("/cards/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// List returns a list of cards.
func (s *CardsService) List(ctx context.Context, params *CardListParams) (*CardList, error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	list, err := Do[CardList](ctx, s.client, http.MethodGet, "/cards", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Update updates a card.
func (s *CardsService) Update(ctx context.Context, id string, params UpdateCardParams) (*Card, error) {
	card, err := Do[Card](ctx, s.client, http.MethodPatch, fmt.Sprintf("/cards/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// Delete cancels and removes a card.
func (s *CardsService) Delete(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/cards/%s", id), nil, nil)
	return err
}

// Freeze freezes a card, preventing any transactions.
func (s *CardsService) Freeze(ctx context.Context, id string) (*Card, error) {
	card, err := Do[Card](ctx, s.client, http.MethodPost, fmt.Sprintf("/cards/%s/freeze", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// Unfreeze unfreezes a frozen card.
func (s *CardsService) Unfreeze(ctx context.Context, id string) (*Card, error) {
	card, err := Do[Card](ctx, s.client, http.MethodPost, fmt.Sprintf("/cards/%s/unfreeze", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// CreatePolicy creates a spending policy for a card.
func (s *CardsService) CreatePolicy(ctx context.Context, params CreatePolicyParams) (*SpendingPolicy, error) {
	policy, err := Do[SpendingPolicy](ctx, s.client, http.MethodPost, "/cards/policies", params, nil)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListPolicies lists all spending policies for a card.
func (s *CardsService) ListPolicies(ctx context.Context, cardID string) ([]SpendingPolicy, error) {
	q := url.Values{}
	q.Set("cardId", cardID)
	envelope, err := Do[policyListEnvelope](ctx, s.client, http.MethodGet, "/cards/policies", nil, q)
	if err != nil {
		return nil, err
	}
	return envelope.Items, nil
}

// UpdatePolicy updates a spending policy.
func (s *CardsService) UpdatePolicy(ctx context.Context, policyID string, params UpdatePolicyParams) (*SpendingPolicy, error) {
	policy, err := Do[SpendingPolicy](ctx, s.client, http.MethodPatch, fmt.Sprintf("/cards/policies/%s", policyID), params, nil)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// DeletePolicy deletes a spending policy.
func (s *CardsService) DeletePolicy(ctx context.Context, policyID string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/cards/policies/%s", policyID), nil, nil)
	return err
}

// ListTransactions returns a list of card transactions.
func (s *CardsService) ListTransactions(ctx context.Context, params *TransactionListParams) (*TransactionList, error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	list, err := Do[TransactionList](ctx, s.client, http.MethodGet, "/cards/transactions", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// GetTransaction retrieves a transaction by ID.
func (s *CardsService) GetTransaction(ctx context.Context, transactionID string) (*CardTransaction, error) {
	txn, err := Do[CardTransaction](ctx, s.client, http.MethodGet, fmt.Sprintf("/cards/transactions/%s", transactionID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

// KillSwitch activates or deactivates the kill switch for a card or agent.
func (s *CardsService) KillSwitch(ctx context.Context, params KillSwitchParams) (*KillSwitchResult, error) {
	result, err := Do[KillSwitchResult](ctx, s.client, http.MethodPost, "/cards/kill-switch", params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListApprovals returns a list of pending card approvals.
func (s *CardsService) ListApprovals(ctx context.Context, params *ApprovalListParams) (*ApprovalList, error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	list, err := Do[ApprovalList](ctx, s.client, http.MethodGet, "/cards/approvals", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// DecideApproval approves or declines a card approval.
func (s *CardsService) DecideApproval(ctx context.Context, approvalID string, decision ApprovalDecision) (*CardApproval, error) {
	body := struct {
		Decision ApprovalDecision `json:"decision"`
	}{Decision: decision}
	approval, err := Do[CardApproval](ctx, s.client, http.MethodPost, fmt.Sprintf("/cards/approvals/%s/decision", approvalID), body, nil)
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// IntPtr returns a pointer to the given int. Useful for optional fields.
func IntPtr(v int) *int {
	return &v
}

// StringPtr returns a pointer to the given string. Useful for optional fields.
func StringPtr(v string) *string {
	return &v
}

// BoolPtr returns a pointer to the given bool. Useful for optional fields.
func BoolPtr(v bool) *bool {
	return &v
}
