package anima

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestDraftsService_Create_MinimalOmitsUnsetFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email/drafts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ak_test_key" {
			t.Errorf("expected Authorization: Bearer ak_test_key, got %s", got)
		}

		body := decodeRawBody(t, r)
		if body["agentId"] != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %v", body["agentId"])
		}

		// Drafts may be created incomplete, but the contract does not accept
		// null for unset optionals — a minimal create must omit those keys
		// entirely, not send them as null (the server would 400).
		for _, key := range []string{
			"fromIdentityId", "to", "cc", "bcc", "subject", "body",
			"bodyHtml", "inReplyTo", "references", "metadata",
		} {
			if _, present := body[key]; present {
				t.Errorf("expected %q to be omitted from a minimal draft create, but it was present", key)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":             "draft_1",
			"agentId":        "agent_123",
			"orgId":          "org_1",
			"fromIdentityId": nil,
			"to":             []string{},
			"cc":             []string{},
			"bcc":            []string{},
			"subject":        nil,
			"body":           nil,
			"bodyHtml":       nil,
			"inReplyTo":      nil,
			"references":     []string{},
			"metadata":       nil,
			"createdAt":      "2026-07-17T10:00:00Z",
			"updatedAt":      "2026-07-17T10:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	draft, err := client.Drafts.Create(context.Background(), CreateDraftParams{
		AgentID: "agent_123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.ID != "draft_1" {
		t.Errorf("expected ID 'draft_1', got %q", draft.ID)
	}
	// An incomplete draft round-trips its nulls as nil pointers, not
	// zero-value strings — the caller must be able to tell "no subject yet"
	// from "empty subject".
	if draft.Subject != nil {
		t.Errorf("expected nil Subject on an incomplete draft, got %q", *draft.Subject)
	}
	if draft.Body != nil {
		t.Errorf("expected nil Body on an incomplete draft, got %q", *draft.Body)
	}
	if draft.FromIdentityID != nil {
		t.Errorf("expected nil FromIdentityID, got %q", *draft.FromIdentityID)
	}
	if len(draft.To) != 0 {
		t.Errorf("expected empty To, got %v", draft.To)
	}
}

func TestDraftsService_Create_FullWireShape(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email/drafts", func(w http.ResponseWriter, r *http.Request) {
		body := decodeRawBody(t, r)

		if body["agentId"] != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %v", body["agentId"])
		}
		if body["fromIdentityId"] != "ident_1" {
			t.Errorf("expected fromIdentityId 'ident_1', got %v", body["fromIdentityId"])
		}
		to, ok := body["to"].([]interface{})
		if !ok || len(to) != 1 || to[0] != "human@example.com" {
			t.Errorf("expected to ['human@example.com'], got %v", body["to"])
		}
		if body["subject"] != "Quarterly report" {
			t.Errorf("expected subject 'Quarterly report', got %v", body["subject"])
		}
		if body["body"] != "Draft body" {
			t.Errorf("expected body 'Draft body', got %v", body["body"])
		}
		if body["bodyHtml"] != "<p>Draft body</p>" {
			t.Errorf("expected bodyHtml '<p>Draft body</p>', got %v", body["bodyHtml"])
		}
		if body["inReplyTo"] != "<msg-1@agents.useanima.sh>" {
			t.Errorf("expected inReplyTo '<msg-1@agents.useanima.sh>', got %v", body["inReplyTo"])
		}
		refs, ok := body["references"].([]interface{})
		if !ok || len(refs) != 1 || refs[0] != "<msg-0@agents.useanima.sh>" {
			t.Errorf("expected references ['<msg-0@agents.useanima.sh>'], got %v", body["references"])
		}
		meta, ok := body["metadata"].(map[string]interface{})
		if !ok || meta["campaign"] != "q3" {
			t.Errorf("expected metadata.campaign 'q3', got %v", body["metadata"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         "draft_2",
			"agentId":    "agent_123",
			"orgId":      "org_1",
			"to":         []string{"human@example.com"},
			"subject":    "Quarterly report",
			"body":       "Draft body",
			"references": []string{"<msg-0@agents.useanima.sh>"},
			"createdAt":  "2026-07-17T10:00:00Z",
			"updatedAt":  "2026-07-17T10:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	draft, err := client.Drafts.Create(context.Background(), CreateDraftParams{
		AgentID:        "agent_123",
		FromIdentityID: "ident_1",
		To:             []string{"human@example.com"},
		Subject:        "Quarterly report",
		Body:           "Draft body",
		BodyHTML:       "<p>Draft body</p>",
		InReplyTo:      "<msg-1@agents.useanima.sh>",
		References:     []string{"<msg-0@agents.useanima.sh>"},
		Metadata:       map[string]interface{}{"campaign": "q3"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.ID != "draft_2" {
		t.Errorf("expected ID 'draft_2', got %q", draft.ID)
	}
	if draft.Subject == nil || *draft.Subject != "Quarterly report" {
		t.Errorf("expected Subject 'Quarterly report', got %v", draft.Subject)
	}
}

func TestDraftsService_Get(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email/drafts/draft_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "draft_1",
			"agentId": "agent_123",
			"orgId":   "org_1",
			"subject": "Hello",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	draft, err := client.Drafts.Get(context.Background(), "draft_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.ID != "draft_1" {
		t.Errorf("expected ID 'draft_1', got %q", draft.ID)
	}
}

func TestDraftsService_List(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email/drafts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		q := r.URL.Query()
		if got := q.Get("agentId"); got != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %q", got)
		}
		if got := q.Get("limit"); got != "10" {
			t.Errorf("expected limit '10', got %q", got)
		}
		if got := q.Get("cursor"); got != "draft_0" {
			t.Errorf("expected cursor 'draft_0', got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "draft_1", "agentId": "agent_123", "subject": "First"},
				{"id": "draft_2", "agentId": "agent_123", "subject": nil},
			},
			"pagination": map[string]interface{}{"nextCursor": nil, "hasMore": false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Drafts.List(context.Background(), &DraftListParams{
		ListParams: ListParams{Limit: 10, Cursor: "draft_0"},
		AgentID:    "agent_123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(page.Items))
	}
	if page.Items[0].Subject == nil || *page.Items[0].Subject != "First" {
		t.Errorf("expected first draft subject 'First', got %v", page.Items[0].Subject)
	}
	if page.Items[1].Subject != nil {
		t.Errorf("expected second draft subject nil, got %v", *page.Items[1].Subject)
	}
	if page.Pagination.HasMore {
		t.Error("expected HasMore false")
	}
}

// TestDraftsService_Send asserts the intent of the send operation: it
// converts the draft into a real Message — the response is a Message with
// delivery state, not the draft.
func TestDraftsService_Send(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email/drafts/draft_1/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["id"] != "draft_1" {
			t.Errorf("expected id 'draft_1' in body, got %v", body["id"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{
			ID:        "msg_from_draft",
			AgentID:   "agent_123",
			Channel:   MessageChannelEmail,
			Direction: MessageDirectionOutbound,
			Status:    MessageStatusQueued,
			ToAddress: "human@example.com",
			Body:      "Draft body",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	msg, err := client.Drafts.Send(context.Background(), "draft_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "msg_from_draft" {
		t.Errorf("expected Message ID 'msg_from_draft', got %q", msg.ID)
	}
	if msg.Status != MessageStatusQueued {
		t.Errorf("expected a Message with delivery status QUEUED, got %q", msg.Status)
	}
}

// TestDraftsService_Send_IncompleteDraft asserts that sending a draft
// without recipients/subject/body surfaces the server's validation error
// loudly (the draft survives server-side for fix-and-retry).
func TestDraftsService_Send_IncompleteDraft(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email/drafts/draft_empty/send", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Draft must have at least one recipient before sending",
				"code":    "BAD_REQUEST",
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	_, err := client.Drafts.Send(context.Background(), "draft_empty")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.Status)
	}
}

func TestDraftsService_Delete(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email/drafts/draft_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "draft_1",
			"agentId": "agent_123",
			"orgId":   "org_1",
			"subject": "Hello",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	// Delete returns the deleted draft (not a bare success flag) so callers
	// can log or restore what was discarded.
	draft, err := client.Drafts.Delete(context.Background(), "draft_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.ID != "draft_1" {
		t.Errorf("expected deleted draft ID 'draft_1', got %q", draft.ID)
	}
}
