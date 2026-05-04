package anima

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVaultService_ShareCredential(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vault/share", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var params ShareCredentialParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.CredentialID != "cred_001" {
			t.Errorf("expected credentialId 'cred_001', got %q", params.CredentialID)
		}
		if params.TargetAgentID != "agent_002" {
			t.Errorf("expected targetAgentId 'agent_002', got %q", params.TargetAgentID)
		}
		if params.Permission != "READ" {
			t.Errorf("expected permission 'READ', got %q", params.Permission)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VaultShare{
			ID:            "share_001",
			CredentialID:  "cred_001",
			SourceAgentID: "agent_001",
			TargetAgentID: "agent_002",
			Permission:    "READ",
			CreatedAt:     "2025-01-01T00:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	share, err := client.Vault.ShareCredential(context.Background(), ShareCredentialParams{
		AgentID:       "agent_001",
		CredentialID:  "cred_001",
		TargetAgentID: "agent_002",
		Permission:    "READ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if share.ID != "share_001" {
		t.Errorf("expected ID 'share_001', got %q", share.ID)
	}
	if share.Permission != "READ" {
		t.Errorf("expected Permission 'READ', got %q", share.Permission)
	}
}

func TestVaultService_ListShares(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vault/shares", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("direction"); got != "granted" {
			t.Errorf("expected direction 'granted', got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VaultShareList{
			Items: []VaultShare{{
				ID:            "share_001",
				CredentialID:  "cred_001",
				SourceAgentID: "agent_001",
				TargetAgentID: "agent_002",
				Permission:    "READ",
				CreatedAt:     "2025-01-01T00:00:00Z",
			}},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	list, err := client.Vault.ListShares(context.Background(), "granted", "agent_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 share, got %d", len(list.Items))
	}
	if list.Items[0].ID != "share_001" {
		t.Errorf("expected share ID 'share_001', got %q", list.Items[0].ID)
	}
}

func TestVaultService_RevokeShare(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vault/share/revoke", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var params RevokeShareParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.ShareID != "share_001" {
			t.Errorf("expected shareId 'share_001', got %q", params.ShareID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SuccessResult{Success: true})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	err := client.Vault.RevokeShare(context.Background(), RevokeShareParams{
		ShareID: "share_001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVaultService_CreateToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vault/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var params CreateTokenParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.CredentialID != "cred_001" {
			t.Errorf("expected credentialId 'cred_001', got %q", params.CredentialID)
		}
		if params.Scope != "autofill" {
			t.Errorf("expected scope 'autofill', got %q", params.Scope)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VaultToken{
			Token:        "vtk_abc123",
			CredentialID: "cred_001",
			Scope:        "autofill",
			ExpiresAt:    "2025-01-01T00:01:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	token, err := client.Vault.CreateToken(context.Background(), CreateTokenParams{
		CredentialID: "cred_001",
		Scope:        "autofill",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.Token != "vtk_abc123" {
		t.Errorf("expected token 'vtk_abc123', got %q", token.Token)
	}
	if token.Scope != "autofill" {
		t.Errorf("expected scope 'autofill', got %q", token.Scope)
	}
}

func TestVaultService_ExchangeToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vault/token/exchange", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Token != "vtk_abc123" {
			t.Errorf("expected token 'vtk_abc123', got %q", body.Token)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VaultCredential{
			ID:       "cred_001",
			Type:     CredentialTypeLogin,
			Name:     "GitHub",
			Favorite: false,
			Login: &VaultLoginData{
				Username: "octocat",
				Password: "secret123",
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	cred, err := client.Vault.ExchangeToken(context.Background(), "vtk_abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred.ID != "cred_001" {
		t.Errorf("expected ID 'cred_001', got %q", cred.ID)
	}
	if cred.Login == nil || cred.Login.Username != "octocat" {
		t.Error("expected login.username 'octocat'")
	}
}

func TestVaultService_RevokeTokens(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/vault/token/revoke", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var params RevokeTokensParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.CredentialID != "cred_001" {
			t.Errorf("expected credentialId 'cred_001', got %q", params.CredentialID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RevokeTokensResult{
			Success: true,
			Revoked: 3,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Vault.RevokeTokens(context.Background(), RevokeTokensParams{
		CredentialID: "cred_001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.Revoked != 3 {
		t.Errorf("expected 3 revoked, got %d", result.Revoked)
	}
}
