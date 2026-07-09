package anima

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestA2AService_Dispatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agents/agent_from/a2a/dispatch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/agents/agent_from/a2a/dispatch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ak_test_key" {
			t.Errorf("expected Authorization: Bearer ak_test_key, got %s", got)
		}

		// Decode into a raw map so we can assert on exactly which keys the
		// SDK put on the wire — fromAgentId is merged in alongside the
		// embedded DispatchA2ATaskParams fields.
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if raw["fromAgentId"] != "agent_from" {
			t.Errorf("expected fromAgentId 'agent_from', got %v", raw["fromAgentId"])
		}
		if raw["toDid"] != "did:web:example.com:agents:agent_to" {
			t.Errorf("expected toDid 'did:web:example.com:agents:agent_to', got %v", raw["toDid"])
		}
		if raw["type"] != "greeting" {
			t.Errorf("expected type 'greeting', got %v", raw["type"])
		}
		inputMap, ok := raw["input"].(map[string]any)
		if !ok || inputMap["message"] != "hello" {
			t.Errorf("expected input.message 'hello', got %v", raw["input"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(A2ATask{
			ID:        "task_001",
			AgentID:   "agent_from",
			Status:    A2ATaskStatusPending,
			CreatedAt: "2026-07-09T00:00:00Z",
			UpdatedAt: "2026-07-09T00:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	task, err := client.A2A.Dispatch(context.Background(), "agent_from", DispatchA2ATaskParams{
		ToDID: "did:web:example.com:agents:agent_to",
		Type:  "greeting",
		Input: map[string]any{"message": "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "task_001" {
		t.Errorf("expected ID 'task_001', got %q", task.ID)
	}
	if task.Status != A2ATaskStatusPending {
		t.Errorf("expected Status 'pending', got %q", task.Status)
	}
}

func TestA2AService_Discover(t *testing.T) {
	// The foreign agent's own host — Discover must hit this directly, not
	// the SDK client's configured base URL.
	agentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/.well-known/agent.json" {
			t.Errorf("expected path /.well-known/agent.json, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("expected no Authorization header on a foreign-host fetch, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"did":  "did:web:example.com:agents:agent_to",
			"name": "Example Agent",
		})
	}))
	defer agentServer.Close()

	// The SDK client's own mux is never hit by Discover; it exists only to
	// give us a *Client to call the method on.
	client, ts := newTestClient(http.NewServeMux())
	defer ts.Close()

	card, err := client.A2A.Discover(context.Background(), agentServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card["did"] != "did:web:example.com:agents:agent_to" {
		t.Errorf("expected did 'did:web:example.com:agents:agent_to', got %v", card["did"])
	}
	if card["name"] != "Example Agent" {
		t.Errorf("expected name 'Example Agent', got %v", card["name"])
	}
}

func TestA2AService_Discover_TrimsTrailingSlash(t *testing.T) {
	var gotPath string
	agentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"did": "did:web:example.com"})
	}))
	defer agentServer.Close()

	client, ts := newTestClient(http.NewServeMux())
	defer ts.Close()

	if _, err := client.A2A.Discover(context.Background(), agentServer.URL+"/"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/.well-known/agent.json" {
		t.Errorf("expected /.well-known/agent.json, got %q", gotPath)
	}
}

func TestA2AService_Discover_ErrorStatus(t *testing.T) {
	agentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer agentServer.Close()

	client, ts := newTestClient(http.NewServeMux())
	defer ts.Close()

	if _, err := client.A2A.Discover(context.Background(), agentServer.URL); err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}
