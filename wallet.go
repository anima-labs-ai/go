package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Wallet represents an agent's crypto wallet.
type Wallet struct {
	ID        string  `json:"id"`
	AgentID   string  `json:"agentId"`
	Address   string  `json:"address"`
	Network   string  `json:"network"`
	Balance   string  `json:"balance,omitempty"`
	Currency  string  `json:"currency,omitempty"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// CreateWalletParams contains the parameters for creating an agent wallet.
type CreateWalletParams struct {
	Network  string `json:"network,omitempty"`
	Currency string `json:"currency,omitempty"`
}

// UpdateWalletParams contains the parameters for updating an agent wallet.
type UpdateWalletParams struct {
	Status string `json:"status,omitempty"`
}

// WalletPayParams contains the parameters for making a payment from a wallet.
type WalletPayParams struct {
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Currency string `json:"currency,omitempty"`
	Memo     string `json:"memo,omitempty"`
}

// WalletPayResult contains the result of a wallet payment.
type WalletPayResult struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	To            string `json:"to"`
}

// X402FetchParams contains the parameters for an X-402 payment-gated fetch.
type X402FetchParams struct {
	URL      string `json:"url"`
	MaxPrice string `json:"maxPrice,omitempty"`
}

// X402FetchResult contains the result of an X-402 fetch.
type X402FetchResult struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body"`
	Paid    bool              `json:"paid"`
	Amount  string            `json:"amount,omitempty"`
}

// WalletTransaction represents a transaction on an agent wallet.
type WalletTransaction struct {
	ID            string `json:"id"`
	WalletID      string `json:"walletId"`
	Type          string `json:"type"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	To            string `json:"to,omitempty"`
	From          string `json:"from,omitempty"`
	Memo          string `json:"memo,omitempty"`
	TransactionHash string `json:"transactionHash,omitempty"`
	CreatedAt     string `json:"createdAt"`
}

// WalletTransactionList wraps a list of wallet transactions.
type WalletTransactionList struct {
	Items []WalletTransaction `json:"items"`
}

// WalletTransactionListParams contains parameters for listing wallet transactions.
type WalletTransactionListParams struct {
	Cursor string
	Limit  int
	Status string
}

// WalletService provides methods for managing agent crypto wallets.
type WalletService struct {
	client *httpClient
}

// newWalletService creates a new WalletService.
func newWalletService(c *httpClient) *WalletService {
	return &WalletService{client: c}
}

// Create creates a wallet for an agent.
func (s *WalletService) Create(ctx context.Context, agentID string, params *CreateWalletParams) (*Wallet, error) {
	wallet, err := Do[Wallet](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/wallet", agentID), params, nil)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Get retrieves an agent's wallet.
func (s *WalletService) Get(ctx context.Context, agentID string) (*Wallet, error) {
	wallet, err := Do[Wallet](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s/wallet", agentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Update updates an agent's wallet.
func (s *WalletService) Update(ctx context.Context, agentID string, params UpdateWalletParams) (*Wallet, error) {
	wallet, err := Do[Wallet](ctx, s.client, http.MethodPut, fmt.Sprintf("/agents/%s/wallet", agentID), params, nil)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// Pay makes a payment from an agent's wallet.
func (s *WalletService) Pay(ctx context.Context, agentID string, params WalletPayParams) (*WalletPayResult, error) {
	result, err := Do[WalletPayResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/wallet/pay", agentID), params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// X402Fetch performs an X-402 payment-gated HTTP fetch through the agent's wallet.
func (s *WalletService) X402Fetch(ctx context.Context, agentID string, params X402FetchParams) (*X402FetchResult, error) {
	result, err := Do[X402FetchResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/wallet/x402-fetch", agentID), params, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Transactions lists transactions for an agent's wallet.
func (s *WalletService) Transactions(ctx context.Context, agentID string, params *WalletTransactionListParams) (*WalletTransactionList, error) {
	var q url.Values
	if params != nil {
		q = url.Values{}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Status != "" {
			q.Set("status", params.Status)
		}
	}
	list, err := Do[WalletTransactionList](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s/wallet/transactions", agentID), nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Freeze freezes an agent's wallet, preventing all outgoing transactions.
func (s *WalletService) Freeze(ctx context.Context, agentID string) error {
	_, err := Do[SuccessResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/wallet/freeze", agentID), nil, nil)
	return err
}

// Unfreeze unfreezes a previously frozen agent wallet.
func (s *WalletService) Unfreeze(ctx context.Context, agentID string) error {
	_, err := Do[SuccessResult](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/wallet/unfreeze", agentID), nil, nil)
	return err
}
