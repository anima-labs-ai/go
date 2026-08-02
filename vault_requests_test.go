package anima

// Tests for the vault surface the go SDK was missing entirely: listing vault
// identities, querying the audit trail, brokering an outbound call through a
// credential, and the human-in-the-loop credential request lifecycle. Six
// routes the API has served all along, reachable from node and python but not
// from here.
//
// Source of truth, at the commit pinned in .anima-ref:
//
//	packages/contracts/src/schemas/vault.ts
//	packages/contracts/src/contracts/vault.ts

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVaultService_ListIdentities(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/vault/identities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("status"); got != "ACTIVE" {
			t.Errorf("expected status=ACTIVE, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Errorf("expected limit=5, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"id":              "vault_1",
					"agentId":         "agent_1",
					"orgId":           "org_1",
					"status":          "ACTIVE",
					"credentialCount": 3,
					"lastSyncAt":      nil,
					"createdAt":       "2026-07-31T00:00:00Z",
					"agentName":       "Support Bot",
					"agentSlug":       "support-bot",
				},
			},
			"pagination": map[string]interface{}{"nextCursor": nil, "hasMore": false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Vault.ListIdentities(context.Background(), &ListVaultIdentitiesParams{
		ListParams: ListParams{Limit: 5},
		Status:     "ACTIVE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 identity, got %d", len(page.Items))
	}
	// The embedded VaultIdentity and the list-only fields must both decode.
	if page.Items[0].AgentID != "agent_1" {
		t.Errorf("expected agentId 'agent_1', got %q", page.Items[0].AgentID)
	}
	if page.Items[0].AgentName != "Support Bot" {
		t.Errorf("expected agentName 'Support Bot', got %q", page.Items[0].AgentName)
	}
	if page.Items[0].CredentialCount != 3 {
		t.Errorf("expected credentialCount 3, got %d", page.Items[0].CredentialCount)
	}
}

func TestVaultService_Audit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/vault/audit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		q := r.URL.Query()
		for param, want := range map[string]string{
			"credentialId": "cred_1",
			"action":       "access_reveal",
			"since":        "2026-07-01T00:00:00Z",
		} {
			if got := q.Get(param); got != want {
				t.Errorf("expected %s=%q, got %q", param, want, got)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"id":           "audit_1",
					"credentialId": "cred_1",
					"agentId":      "agent_1",
					"orgId":        "org_1",
					"action":       "access_reveal",
					"actor":        "user_1",
					"ipAddress":    "203.0.113.4",
					"metadata":     map[string]interface{}{"reason": "support"},
					"createdAt":    "2026-07-31T00:00:00Z",
				},
			},
			"pagination": map[string]interface{}{"nextCursor": nil, "hasMore": false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Vault.Audit(context.Background(), &VaultAuditParams{
		CredentialID: "cred_1",
		Action:       "access_reveal",
		Since:        "2026-07-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(page.Items))
	}
	if page.Items[0].Actor != "user_1" {
		t.Errorf("expected actor 'user_1', got %q", page.Items[0].Actor)
	}
	if page.Items[0].IPAddress == nil || *page.Items[0].IPAddress != "203.0.113.4" {
		t.Errorf("expected ipAddress '203.0.113.4', got %v", page.Items[0].IPAddress)
	}
}

// UseCredential is the broker: the agent gets the upstream response and never
// the secret. The test asserts the request is forwarded verbatim and that the
// response body comes back untouched.
func TestVaultService_UseCredential(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/vault/credentials/cred_1/use", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var params UseCredentialParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Method != "GET" {
			t.Errorf("expected method GET, got %q", params.Method)
		}
		if params.URL != "https://api.example.com/me" {
			t.Errorf("unexpected url %q", params.URL)
		}
		if params.Headers["Accept"] != "application/json" {
			t.Errorf("headers not forwarded: %v", params.Headers)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    200,
			"headers":   map[string]string{"content-type": "application/json"},
			"body":      `{"ok":true}`,
			"truncated": false,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Vault.UseCredential(context.Background(), "cred_1", UseCredentialParams{
		Method:  "GET",
		URL:     "https://api.example.com/me",
		Headers: map[string]string{"Accept": "application/json"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != 200 {
		t.Errorf("expected upstream status 200, got %d", result.Status)
	}
	if result.Body != `{"ok":true}` {
		t.Errorf("unexpected upstream body %q", result.Body)
	}
	if result.Truncated {
		t.Error("expected truncated=false")
	}
}

// A created request is PENDING with a fill URL and no credential yet; the
// agent polls for the reference. CredentialID stays nil until FULFILLED, which
// is the whole point of the pointer.
func TestVaultService_CreateCredentialRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/vault/credential-requests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var params CreateCredentialRequestParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Type != CredentialTypeAPIKey {
			t.Errorf("expected type 'api_key', got %q", params.Type)
		}
		if params.Reason != "needed to call the billing API" {
			t.Errorf("unexpected reason %q", params.Reason)
		}
		if params.TTLSeconds != 600 {
			t.Errorf("expected ttlSeconds 600, got %d", params.TTLSeconds)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"requestId":    "req_1",
			"fillUrl":      "https://vault.useanima.sh/fill/tok_1",
			"status":       "PENDING",
			"expiresAt":    "2026-07-31T00:10:00Z",
			"emailSent":    true,
			"credentialId": nil,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	req, err := client.Vault.CreateCredentialRequest(context.Background(), CreateCredentialRequestParams{
		Type:       CredentialTypeAPIKey,
		Name:       "Billing API key",
		Reason:     "needed to call the billing API",
		TTLSeconds: 600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Status != CredentialRequestStatusPending {
		t.Errorf("expected PENDING, got %q", req.Status)
	}
	if req.CredentialID != nil {
		t.Errorf("expected no credential on the pending path, got %v", *req.CredentialID)
	}
	if !req.EmailSent {
		t.Error("expected emailSent=true")
	}
}

func TestVaultService_GetCredentialRequestStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/vault/credential-requests/req_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":        "FULFILLED",
			"credentialId":  "cred_9",
			"maskedPreview": "****1234",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Vault.GetCredentialRequestStatus(context.Background(), "req_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != CredentialRequestStatusFulfilled {
		t.Errorf("expected FULFILLED, got %q", result.Status)
	}
	if result.CredentialID == nil || *result.CredentialID != "cred_9" {
		t.Errorf("expected credentialId 'cred_9', got %v", result.CredentialID)
	}
	// The preview confirms which secret arrived without revealing it.
	if result.MaskedPreview == nil || *result.MaskedPreview != "****1234" {
		t.Errorf("expected maskedPreview '****1234', got %v", result.MaskedPreview)
	}
}

func TestVaultService_CancelCredentialRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/vault/credential-requests/req_1/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "CANCELLED"})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Vault.CancelCredentialRequest(context.Background(), "req_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != CredentialRequestStatusCancelled {
		t.Errorf("expected CANCELLED, got %q", result.Status)
	}
}

// CredentialType must carry exactly the seven values CredentialTypeSchema
// enumerates (packages/contracts/src/schemas/vault.ts). Only four were
// declared: oauth_token, api_key and certificate had no constant, so a caller
// had to hand-write the string.
//
// Asserted as a value set, not a loop of non-empty checks — that version could
// not fail, and would not have noticed the three missing members either.
func TestCredentialTypeCoversTheContract(t *testing.T) {
	got := map[CredentialType]bool{
		CredentialTypeLogin:       true,
		CredentialTypeSecureNote:  true,
		CredentialTypeCard:        true,
		CredentialTypeIdentity:    true,
		CredentialTypeOAuthToken:  true,
		CredentialTypeAPIKey:      true,
		CredentialTypeCertificate: true,
	}
	want := []CredentialType{
		"login", "secure_note", "card", "identity",
		"oauth_token", "api_key", "certificate",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d distinct credential types, got %d — two constants share a value", len(want), len(got))
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("no constant carries the contract value %q", w)
		}
	}
}
