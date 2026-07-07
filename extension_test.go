package anima

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestExtensionService_Connect_MasterKey(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/extension/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Decode into a raw map so we can assert on exactly which keys the
		// SDK put on the wire (omitempty behavior).
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if raw["agentId"] != "agent_001" {
			t.Errorf("expected agentId 'agent_001', got %v", raw["agentId"])
		}
		if raw["ttl"] != "1h" {
			t.Errorf("expected ttl '1h', got %v", raw["ttl"])
		}

		w.Header().Set("Content-Type", "application/json")
		expires := "2026-07-07T01:00:00Z"
		json.NewEncoder(w).Encode(ConnectExtensionResult{
			AgentID:           "agent_001",
			ConnectURL:        "https://useanima.sh/extension/connect?exchange=xchg_abc",
			ExpiresAt:         &expires,
			ExchangeExpiresAt: "2026-07-07T00:05:00Z",
			Policy:            "pre_approved",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Extension.Connect(context.Background(), ConnectExtensionParams{
		AgentID: "agent_001",
		TTL:     "1h",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "agent_001" {
		t.Errorf("expected AgentID 'agent_001', got %q", result.AgentID)
	}
	if result.ConnectURL != "https://useanima.sh/extension/connect?exchange=xchg_abc" {
		t.Errorf("unexpected ConnectURL: %q", result.ConnectURL)
	}
	if result.ExpiresAt == nil || *result.ExpiresAt != "2026-07-07T01:00:00Z" {
		t.Errorf("expected ExpiresAt '2026-07-07T01:00:00Z', got %v", result.ExpiresAt)
	}
	if result.ExchangeExpiresAt != "2026-07-07T00:05:00Z" {
		t.Errorf("expected ExchangeExpiresAt '2026-07-07T00:05:00Z', got %q", result.ExchangeExpiresAt)
	}
	if result.Policy != "pre_approved" {
		t.Errorf("expected Policy 'pre_approved', got %q", result.Policy)
	}
}

func TestExtensionService_Connect_AgentKey(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/extension/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// With an agent key the caller omits AgentID; omitempty must keep it
		// off the wire entirely. Session TTL is also omitted here.
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if _, present := raw["agentId"]; present {
			t.Errorf("expected agentId to be absent, got %v", raw["agentId"])
		}
		if _, present := raw["ttl"]; present {
			t.Errorf("expected ttl to be absent, got %v", raw["ttl"])
		}

		w.Header().Set("Content-Type", "application/json")
		// A "session" policy yields a null expiresAt — the pointer must decode to nil.
		json.NewEncoder(w).Encode(map[string]any{
			"agentId":           "agent_002",
			"connectUrl":        "https://useanima.sh/extension/connect?exchange=xchg_def",
			"expiresAt":         nil,
			"exchangeExpiresAt": "2026-07-07T00:05:00Z",
			"policy":            "session",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Extension.Connect(context.Background(), ConnectExtensionParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "agent_002" {
		t.Errorf("expected AgentID 'agent_002', got %q", result.AgentID)
	}
	if result.ExpiresAt != nil {
		t.Errorf("expected ExpiresAt to be nil for a session policy, got %q", *result.ExpiresAt)
	}
	if result.Policy != "session" {
		t.Errorf("expected Policy 'session', got %q", result.Policy)
	}
}
