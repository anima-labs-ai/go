package anima

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a client pointed at a test server.
func newTestClient(handler http.Handler) (*Client, *httptest.Server) {
	ts := httptest.NewServer(handler)
	client := NewClient("ak_test_key",
		WithBaseURL(ts.URL),
		WithMaxRetries(0),
	)
	return client, ts
}

func TestAgentsService_Create(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify auth header.
		if got := r.Header.Get("Authorization"); got != "Bearer ak_test_key" {
			t.Errorf("expected Authorization: Bearer ak_test_key, got %s", got)
		}

		// Verify content type.
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %s", got)
		}

		// Decode request body.
		var params CreateAgentParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if params.Name != "Test Agent" {
			t.Errorf("expected name 'Test Agent', got %q", params.Name)
		}
		if params.OrgID != "org_123" {
			t.Errorf("expected orgId 'org_123', got %q", params.OrgID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Agent{
			ID:     "agent_abc",
			OrgID:  "org_123",
			Name:   "Test Agent",
			Slug:   "test-agent",
			Status: AgentStatusActive,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	agent, err := client.Agents.Create(context.Background(), CreateAgentParams{
		OrgID: "org_123",
		Name:  "Test Agent",
		Slug:  "test-agent",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.ID != "agent_abc" {
		t.Errorf("expected ID 'agent_abc', got %q", agent.ID)
	}
	if agent.Name != "Test Agent" {
		t.Errorf("expected Name 'Test Agent', got %q", agent.Name)
	}
	if agent.Status != AgentStatusActive {
		t.Errorf("expected Status 'ACTIVE', got %q", agent.Status)
	}
}

func TestAgentsService_Get(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents/agent_abc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Agent{
			ID:     "agent_abc",
			OrgID:  "org_123",
			Name:   "My Agent",
			Slug:   "my-agent",
			Status: AgentStatusActive,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	agent, err := client.Agents.Get(context.Background(), "agent_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.ID != "agent_abc" {
		t.Errorf("expected ID 'agent_abc', got %q", agent.ID)
	}
}

func TestAgentsService_List(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// Check query params.
		if got := r.URL.Query().Get("orgId"); got != "org_123" {
			t.Errorf("expected orgId 'org_123', got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("expected limit '10', got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page[Agent]{
			Items: []Agent{
				{ID: "agent_1", Name: "Agent One", Status: AgentStatusActive},
				{ID: "agent_2", Name: "Agent Two", Status: AgentStatusActive},
			},
			Pagination: CursorPagination{HasMore: false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Agents.List(context.Background(), &AgentListParams{
		ListParams: ListParams{Limit: 10},
		OrgID:      "org_123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(page.Items))
	}
	if page.Items[0].ID != "agent_1" {
		t.Errorf("expected first item ID 'agent_1', got %q", page.Items[0].ID)
	}
}

func TestAgentsService_Update(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents/agent_abc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Agent{
			ID:     "agent_abc",
			Name:   "Updated Agent",
			Status: AgentStatusActive,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	agent, err := client.Agents.Update(context.Background(), "agent_abc", UpdateAgentParams{
		Name: "Updated Agent",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.Name != "Updated Agent" {
		t.Errorf("expected Name 'Updated Agent', got %q", agent.Name)
	}
}

func TestAgentsService_Delete(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents/agent_abc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	err := client.Agents.Delete(context.Background(), "agent_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsService_RotateKey(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents/agent_abc/rotate-key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RotateKeyResult{
			APIKey:       "ak_live_new_key_123",
			APIKeyPrefix: "ak_live_new",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Agents.RotateKey(context.Background(), "agent_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.APIKey != "ak_live_new_key_123" {
		t.Errorf("expected APIKey 'ak_live_new_key_123', got %q", result.APIKey)
	}
}

func TestAgentsService_ListAutoPaging(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		cursor := "cursor_page2"
		if callCount == 1 {
			json.NewEncoder(w).Encode(Page[Agent]{
				Items: []Agent{
					{ID: "agent_1", Name: "Agent One"},
				},
				Pagination: CursorPagination{
					HasMore:    true,
					NextCursor: &cursor,
				},
			})
		} else {
			json.NewEncoder(w).Encode(Page[Agent]{
				Items: []Agent{
					{ID: "agent_2", Name: "Agent Two"},
				},
				Pagination: CursorPagination{HasMore: false},
			})
		}
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	iter := client.Agents.ListAutoPaging(nil)
	var agents []Agent
	ctx := context.Background()
	for iter.Next(ctx) {
		agents = append(agents, iter.Current())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestNewClient_ServicesWired(t *testing.T) {
	client := NewClient("ak_test_key")

	if client.Agents == nil {
		t.Error("Agents service is nil")
	}
	if client.Cards == nil {
		t.Error("Cards service is nil")
	}
	if client.Domains == nil {
		t.Error("Domains service is nil")
	}
	if client.Emails == nil {
		t.Error("Emails service is nil")
	}
	if client.Messages == nil {
		t.Error("Messages service is nil")
	}
	if client.Organizations == nil {
		t.Error("Organizations service is nil")
	}
	if client.Phones == nil {
		t.Error("Phones service is nil")
	}
	if client.Security == nil {
		t.Error("Security service is nil")
	}
	if client.Vault == nil {
		t.Error("Vault service is nil")
	}
	if client.Webhooks == nil {
		t.Error("Webhooks service is nil")
	}
}
