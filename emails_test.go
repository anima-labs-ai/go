package anima

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestEmailsService_List(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("agentId"); got != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("expected limit '10', got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page[Message]{
			Items: []Message{
				{ID: "msg_1", Channel: MessageChannelEmail, Direction: MessageDirectionInbound},
			},
			Pagination: CursorPagination{HasMore: false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Emails.List(context.Background(), &EmailListParams{
		ListParams: ListParams{Limit: 10},
		AgentID:    "agent_123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(page.Items))
	}
	if page.Items[0].ID != "msg_1" {
		t.Errorf("expected first item ID 'msg_1', got %q", page.Items[0].ID)
	}
}

func TestEmailsService_ListAutoPaging(t *testing.T) {
	callCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/email", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			cursor := "cursor_page2"
			json.NewEncoder(w).Encode(Page[Message]{
				Items:      []Message{{ID: "msg_1"}},
				Pagination: CursorPagination{HasMore: true, NextCursor: &cursor},
			})
			return
		}
		// The second request must carry the cursor from the first page.
		if got := r.URL.Query().Get("cursor"); got != "cursor_page2" {
			t.Errorf("expected cursor 'cursor_page2' on page 2, got %q", got)
		}
		json.NewEncoder(w).Encode(Page[Message]{
			Items:      []Message{{ID: "msg_2"}},
			Pagination: CursorPagination{HasMore: false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	iter := client.Emails.ListAutoPaging(nil)
	var ids []string
	ctx := context.Background()
	for iter.Next(ctx) {
		ids = append(ids, iter.Current().ID)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != "msg_1" || ids[1] != "msg_2" {
		t.Errorf("expected [msg_1 msg_2], got %v", ids)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestEmailsService_UploadAttachment(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/msg_abc/attachments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body := decodeRawBody(t, r)
		if body["messageId"] != "msg_abc" {
			t.Errorf("expected messageId 'msg_abc' in body, got %v", body["messageId"])
		}
		if body["filename"] != "report.pdf" {
			t.Errorf("expected filename 'report.pdf', got %v", body["filename"])
		}
		if body["mimeType"] != "application/pdf" {
			t.Errorf("expected mimeType 'application/pdf', got %v", body["mimeType"])
		}
		if body["sizeBytes"] != float64(2048) {
			t.Errorf("expected sizeBytes 2048, got %v", body["sizeBytes"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Attachment{
			ID:        "att_1",
			Filename:  "report.pdf",
			MimeType:  "application/pdf",
			SizeBytes: 2048,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	att, err := client.Emails.UploadAttachment(context.Background(), "msg_abc", UploadAttachmentParams{
		Filename:  "report.pdf",
		MimeType:  "application/pdf",
		SizeBytes: 2048,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if att.ID != "att_1" {
		t.Errorf("expected ID 'att_1', got %q", att.ID)
	}
}

func TestEmailsService_GetAttachmentURL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/attachments/att_1/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AttachmentDownload{
			URL:       "https://storage.example.com/att_1?sig=abc",
			ExpiresAt: "2026-07-16T12:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	dl, err := client.Emails.GetAttachmentURL(context.Background(), "att_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dl.URL != "https://storage.example.com/att_1?sig=abc" {
		t.Errorf("unexpected URL %q", dl.URL)
	}
}
