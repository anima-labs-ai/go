package anima

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestInboxesService_Create(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/inboxes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ak_test_key" {
			t.Errorf("expected Authorization: Bearer ak_test_key, got %s", got)
		}

		body := decodeRawBody(t, r)
		if body["username"] != "support" {
			t.Errorf("expected username 'support', got %v", body["username"])
		}
		if body["domain"] != "agents.useanima.sh" {
			t.Errorf("expected domain 'agents.useanima.sh', got %v", body["domain"])
		}
		if body["displayName"] != "Support Inbox" {
			t.Errorf("expected displayName 'Support Inbox', got %v", body["displayName"])
		}
		if body["agentId"] != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %v", body["agentId"])
		}

		displayName := "Support Inbox"
		agentID := "agent_123"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Inbox{
			ID:          "inbox_1",
			Email:       "support@agents.useanima.sh",
			Domain:      "agents.useanima.sh",
			LocalPart:   "support",
			DisplayName: &displayName,
			AgentID:     &agentID,
			CreatedAt:   "2026-07-16T00:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	inbox, err := client.Inboxes.Create(context.Background(), CreateInboxParams{
		Username:    "support",
		Domain:      "agents.useanima.sh",
		DisplayName: "Support Inbox",
		AgentID:     "agent_123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inbox.ID != "inbox_1" {
		t.Errorf("expected ID 'inbox_1', got %q", inbox.ID)
	}
	if inbox.Email != "support@agents.useanima.sh" {
		t.Errorf("expected Email 'support@agents.useanima.sh', got %q", inbox.Email)
	}
	if inbox.DisplayName == nil || *inbox.DisplayName != "Support Inbox" {
		t.Errorf("expected DisplayName 'Support Inbox', got %v", inbox.DisplayName)
	}
	if inbox.AgentID == nil || *inbox.AgentID != "agent_123" {
		t.Errorf("expected AgentID 'agent_123', got %v", inbox.AgentID)
	}
}

// TestInboxesService_Create_Minimal asserts that a zero-value create sends an
// empty JSON object. The contract validates username with min(1) when the key
// is present, so serializing `"username": ""` (instead of omitting it) would
// turn "let the server generate an address" into a 400.
func TestInboxesService_Create_Minimal(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/inboxes", func(w http.ResponseWriter, r *http.Request) {
		body := decodeRawBody(t, r)
		if len(body) != 0 {
			t.Errorf("expected empty body for minimal create, got keys %v", body)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Inbox{
			ID:        "inbox_gen",
			Email:     "swift-falcon-9d2@agents.useanima.sh",
			Domain:    "agents.useanima.sh",
			LocalPart: "swift-falcon-9d2",
			CreatedAt: "2026-07-16T00:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	inbox, err := client.Inboxes.Create(context.Background(), CreateInboxParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inbox.ID != "inbox_gen" {
		t.Errorf("expected ID 'inbox_gen', got %q", inbox.ID)
	}
	if inbox.DisplayName != nil {
		t.Errorf("expected nil DisplayName, got %v", *inbox.DisplayName)
	}
	if inbox.AgentID != nil {
		t.Errorf("expected nil AgentID, got %v", *inbox.AgentID)
	}
}

func TestInboxesService_Get(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/inboxes/inbox_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Inbox{
			ID:        "inbox_1",
			Email:     "support@agents.useanima.sh",
			Domain:    "agents.useanima.sh",
			LocalPart: "support",
			CreatedAt: "2026-07-16T00:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	inbox, err := client.Inboxes.Get(context.Background(), "inbox_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inbox.ID != "inbox_1" {
		t.Errorf("expected ID 'inbox_1', got %q", inbox.ID)
	}
	if inbox.LocalPart != "support" {
		t.Errorf("expected LocalPart 'support', got %q", inbox.LocalPart)
	}
}

func TestInboxesService_Get_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/inboxes/inbox_missing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Inbox not found", "code": "NOT_FOUND"},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	_, err := client.Inboxes.Get(context.Background(), "inbox_missing")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInboxesService_List(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/inboxes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("query"); got != "support" {
			t.Errorf("expected query 'support', got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("expected limit '10', got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page[Inbox]{
			Items: []Inbox{
				{ID: "inbox_1", Email: "support@agents.useanima.sh"},
				{ID: "inbox_2", Email: "support-eu@agents.useanima.sh"},
			},
			Pagination: CursorPagination{HasMore: false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Inboxes.List(context.Background(), &InboxListParams{
		ListParams: ListParams{Limit: 10},
		Query:      "support",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(page.Items))
	}
	if page.Items[0].ID != "inbox_1" {
		t.Errorf("expected first item ID 'inbox_1', got %q", page.Items[0].ID)
	}
}

func TestInboxesService_ListAutoPaging(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/inboxes", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			cursor := "cursor_page2"
			json.NewEncoder(w).Encode(Page[Inbox]{
				Items:      []Inbox{{ID: "inbox_1"}},
				Pagination: CursorPagination{HasMore: true, NextCursor: &cursor},
			})
			return
		}
		if got := r.URL.Query().Get("cursor"); got != "cursor_page2" {
			t.Errorf("expected cursor 'cursor_page2' on page 2, got %q", got)
		}
		json.NewEncoder(w).Encode(Page[Inbox]{
			Items:      []Inbox{{ID: "inbox_2"}},
			Pagination: CursorPagination{HasMore: false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	iter := client.Inboxes.ListAutoPaging(nil)
	var ids []string
	ctx := context.Background()
	for iter.Next(ctx) {
		ids = append(ids, iter.Current().ID)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != "inbox_1" || ids[1] != "inbox_2" {
		t.Errorf("expected [inbox_1 inbox_2], got %v", ids)
	}
}

// TestInboxesService_Update asserts that only the fields explicitly set are
// serialized. An absent agentId key means "leave the association unchanged" —
// if the SDK serialized a null agentId here, the server would unlink the
// inbox's agent as a side effect of renaming it.
func TestInboxesService_Update(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/inboxes/inbox_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		body := decodeRawBody(t, r)
		if body["displayName"] != "Renamed Inbox" {
			t.Errorf("expected displayName 'Renamed Inbox', got %v", body["displayName"])
		}
		if _, present := body["agentId"]; present {
			t.Errorf("expected agentId to be omitted when not set, but it was present")
		}

		displayName := "Renamed Inbox"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Inbox{
			ID:          "inbox_1",
			Email:       "support@agents.useanima.sh",
			Domain:      "agents.useanima.sh",
			LocalPart:   "support",
			DisplayName: &displayName,
			CreatedAt:   "2026-07-16T00:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	displayName := "Renamed Inbox"
	inbox, err := client.Inboxes.Update(context.Background(), "inbox_1", UpdateInboxParams{
		DisplayName: &displayName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inbox.DisplayName == nil || *inbox.DisplayName != "Renamed Inbox" {
		t.Errorf("expected DisplayName 'Renamed Inbox', got %v", inbox.DisplayName)
	}
}

func TestInboxesService_Delete(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/inboxes/inbox_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	if err := client.Inboxes.Delete(context.Background(), "inbox_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
