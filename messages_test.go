package anima

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

// decodeRawBody decodes a request body into a generic map so tests can assert
// the exact wire shape (key names and presence), not just Go-side round-trips.
func decodeRawBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("failed to decode request body: %v", err)
	}
	return body
}

func TestMessagesService_SendEmail(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/email", func(w http.ResponseWriter, r *http.Request) {
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
		to, ok := body["to"].([]interface{})
		if !ok || len(to) != 1 || to[0] != "human@example.com" {
			t.Errorf("expected to ['human@example.com'], got %v", body["to"])
		}
		if body["subject"] != "Hello" {
			t.Errorf("expected subject 'Hello', got %v", body["subject"])
		}
		if body["body"] != "Hi there" {
			t.Errorf("expected body 'Hi there', got %v", body["body"])
		}

		// The messages contract does not accept null for the optional fields:
		// a plain send must omit these keys entirely, not send them as null.
		for _, key := range []string{"attachments", "inReplyTo", "references", "cc", "bcc", "headers", "metadata"} {
			if _, present := body[key]; present {
				t.Errorf("expected %q to be omitted from a plain send, but it was present", key)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{
			ID:        "msg_abc",
			AgentID:   "agent_123",
			Channel:   MessageChannelEmail,
			Direction: MessageDirectionOutbound,
			Status:    MessageStatusQueued,
			ToAddress: "human@example.com",
			Body:      "Hi there",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	msg, err := client.Messages.SendEmail(context.Background(), SendEmailParams{
		AgentID: "agent_123",
		To:      []string{"human@example.com"},
		Subject: "Hello",
		Body:    "Hi there",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "msg_abc" {
		t.Errorf("expected ID 'msg_abc', got %q", msg.ID)
	}
	if msg.Channel != MessageChannelEmail {
		t.Errorf("expected Channel 'EMAIL', got %q", msg.Channel)
	}
	if msg.Status != MessageStatusQueued {
		t.Errorf("expected Status 'QUEUED', got %q", msg.Status)
	}
}

// TestMessagesService_SendEmail_AttachmentsAndThreading asserts that
// attachments, inReplyTo, and references actually reach the wire in the
// contract's shape. This is the regression test for the gap where SDK types
// omitted these fields entirely and the API returned 200 while silently
// dropping the attachment.
func TestMessagesService_SendEmail_AttachmentsAndThreading(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/email", func(w http.ResponseWriter, r *http.Request) {
		body := decodeRawBody(t, r)

		atts, ok := body["attachments"].([]interface{})
		if !ok || len(atts) != 2 {
			t.Fatalf("expected 2 attachments on the wire, got %v", body["attachments"])
		}

		inline, ok := atts[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected attachment object, got %T", atts[0])
		}
		if inline["filename"] != "report.pdf" {
			t.Errorf("expected filename 'report.pdf', got %v", inline["filename"])
		}
		if inline["contentType"] != "application/pdf" {
			t.Errorf("expected contentType 'application/pdf', got %v", inline["contentType"])
		}
		if inline["content"] != "JVBERi0xLjQK" {
			t.Errorf("expected base64 content 'JVBERi0xLjQK', got %v", inline["content"])
		}
		if inline["contentId"] != "report" {
			t.Errorf("expected contentId 'report', got %v", inline["contentId"])
		}
		if _, present := inline["url"]; present {
			t.Errorf("expected url to be omitted on an inline attachment, but it was present")
		}

		remote, ok := atts[1].(map[string]interface{})
		if !ok {
			t.Fatalf("expected attachment object, got %T", atts[1])
		}
		if remote["url"] != "https://example.com/logo.png" {
			t.Errorf("expected url 'https://example.com/logo.png', got %v", remote["url"])
		}
		if _, present := remote["content"]; present {
			t.Errorf("expected content to be omitted on a URL attachment, but it was present")
		}

		if body["inReplyTo"] != "<msg-1@agents.useanima.sh>" {
			t.Errorf("expected inReplyTo '<msg-1@agents.useanima.sh>', got %v", body["inReplyTo"])
		}
		refs, ok := body["references"].([]interface{})
		if !ok || len(refs) != 2 || refs[0] != "<msg-0@agents.useanima.sh>" || refs[1] != "<msg-1@agents.useanima.sh>" {
			t.Errorf("expected references chain of 2 Message-IDs, got %v", body["references"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{
			ID:      "msg_reply",
			Channel: MessageChannelEmail,
			Status:  MessageStatusQueued,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	msg, err := client.Messages.SendEmail(context.Background(), SendEmailParams{
		AgentID: "agent_123",
		To:      []string{"human@example.com"},
		Subject: "Re: Hello",
		Body:    "Reply with attachments",
		Attachments: []EmailAttachment{
			{
				Filename:    "report.pdf",
				ContentType: "application/pdf",
				Content:     "JVBERi0xLjQK",
				ContentID:   "report",
			},
			{
				URL: "https://example.com/logo.png",
			},
		},
		InReplyTo:  "<msg-1@agents.useanima.sh>",
		References: []string{"<msg-0@agents.useanima.sh>", "<msg-1@agents.useanima.sh>"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "msg_reply" {
		t.Errorf("expected ID 'msg_reply', got %q", msg.ID)
	}
}

func TestMessagesService_SendEmail_ValidationError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/email", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Attachment must specify exactly one of `content` (base64) or `url`.",
				"code":    "BAD_REQUEST",
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	_, err := client.Messages.SendEmail(context.Background(), SendEmailParams{
		AgentID:     "agent_123",
		To:          []string{"human@example.com"},
		Subject:     "Hello",
		Body:        "Hi",
		Attachments: []EmailAttachment{{Filename: "empty.bin"}},
	})
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

func TestMessagesService_SendSMS(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/phone/send-sms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body := decodeRawBody(t, r)
		if body["agentId"] != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %v", body["agentId"])
		}
		if body["to"] != "+15551234567" {
			t.Errorf("expected to '+15551234567', got %v", body["to"])
		}
		if body["body"] != "ping" {
			t.Errorf("expected body 'ping', got %v", body["body"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{
			ID:        "msg_sms",
			Channel:   MessageChannelSMS,
			Direction: MessageDirectionOutbound,
			Status:    MessageStatusQueued,
			ToAddress: "+15551234567",
			Body:      "ping",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	msg, err := client.Messages.SendSMS(context.Background(), SendSMSParams{
		AgentID: "agent_123",
		To:      "+15551234567",
		Body:    "ping",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "msg_sms" {
		t.Errorf("expected ID 'msg_sms', got %q", msg.ID)
	}
	if msg.Channel != MessageChannelSMS {
		t.Errorf("expected Channel 'SMS', got %q", msg.Channel)
	}
}

func TestMessagesService_Get(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/msg_abc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Message{
			ID:      "msg_abc",
			Channel: MessageChannelEmail,
			Status:  MessageStatusDelivered,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	msg, err := client.Messages.Get(context.Background(), "msg_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "msg_abc" {
		t.Errorf("expected ID 'msg_abc', got %q", msg.ID)
	}
	if msg.Status != MessageStatusDelivered {
		t.Errorf("expected Status 'DELIVERED', got %q", msg.Status)
	}
}

func TestMessagesService_List(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		q := r.URL.Query()
		if got := q.Get("agentId"); got != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %q", got)
		}
		if got := q.Get("channel"); got != "EMAIL" {
			t.Errorf("expected channel 'EMAIL', got %q", got)
		}
		if got := q.Get("direction"); got != "INBOUND" {
			t.Errorf("expected direction 'INBOUND', got %q", got)
		}
		if got := q.Get("dateRange.from"); got != "2026-07-01T00:00:00Z" {
			t.Errorf("expected dateRange.from '2026-07-01T00:00:00Z', got %q", got)
		}
		if got := q.Get("limit"); got != "5" {
			t.Errorf("expected limit '5', got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page[Message]{
			Items: []Message{
				{ID: "msg_1", Channel: MessageChannelEmail},
				{ID: "msg_2", Channel: MessageChannelEmail},
			},
			Pagination: CursorPagination{HasMore: false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Messages.List(context.Background(), &MessageListParams{
		ListParams: ListParams{Limit: 5},
		AgentID:    "agent_123",
		Channel:    MessageChannelEmail,
		Direction:  MessageDirectionInbound,
		DateRange:  &DateRange{From: "2026-07-01T00:00:00Z"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(page.Items))
	}
}

func TestMessagesService_Search(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body := decodeRawBody(t, r)
		if body["query"] != "invoice" {
			t.Errorf("expected query 'invoice', got %v", body["query"])
		}
		filters, ok := body["filters"].(map[string]interface{})
		if !ok || filters["channel"] != "EMAIL" {
			t.Errorf("expected filters.channel 'EMAIL', got %v", body["filters"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Page[Message]{
			Items:      []Message{{ID: "msg_hit"}},
			Pagination: CursorPagination{HasMore: false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Messages.Search(context.Background(), MessageSearchParams{
		Query:   "invoice",
		Filters: &MessageSearchFilters{Channel: MessageChannelEmail},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "msg_hit" {
		t.Errorf("expected single hit 'msg_hit', got %v", page.Items)
	}
}

func TestMessagesService_SemanticSearch(t *testing.T) {
	threshold := 0.5

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/search/semantic", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body := decodeRawBody(t, r)
		if body["query"] != "emails about the delayed shipment" {
			t.Errorf("expected natural-language query, got %v", body["query"])
		}
		if body["agentId"] != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %v", body["agentId"])
		}
		if body["limit"] != float64(5) {
			t.Errorf("expected limit 5, got %v", body["limit"])
		}
		if body["threshold"] != 0.5 {
			t.Errorf("expected threshold 0.5, got %v", body["threshold"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"id":         "msg_1",
					"content":    "Your shipment is delayed by two weeks",
					"similarity": 0.91,
					"channel":    "EMAIL",
					"direction":  "INBOUND",
					"createdAt":  "2026-07-17T10:00:00Z",
					"agentId":    "agent_123",
				},
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	res, err := client.Messages.SemanticSearch(context.Background(), SemanticSearchParams{
		Query:     "emails about the delayed shipment",
		AgentID:   "agent_123",
		Limit:     5,
		Threshold: &threshold,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res.Results))
	}
	hit := res.Results[0]
	if hit.ID != "msg_1" {
		t.Errorf("expected result ID 'msg_1', got %q", hit.ID)
	}
	// Similarity is the ranking signal — it must survive the round trip, not
	// be dropped by a lossy struct.
	if hit.Similarity != 0.91 {
		t.Errorf("expected similarity 0.91, got %v", hit.Similarity)
	}
	if hit.Channel != MessageChannelEmail {
		t.Errorf("expected channel EMAIL, got %q", hit.Channel)
	}
}

// TestMessagesService_SemanticSearch_DefaultsOmitted asserts that unset
// optional params are omitted from the wire so the server's documented
// defaults (limit 10, threshold 0.7) apply — sending zero values instead
// would fail validation (limit min 1) or return everything (threshold 0).
func TestMessagesService_SemanticSearch_DefaultsOmitted(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/search/semantic", func(w http.ResponseWriter, r *http.Request) {
		body := decodeRawBody(t, r)
		if body["query"] != "anything" {
			t.Errorf("expected query 'anything', got %v", body["query"])
		}
		for _, key := range []string{"agentId", "limit", "threshold"} {
			if _, present := body[key]; present {
				t.Errorf("expected %q to be omitted when unset, but it was present", key)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	// No results is a valid outcome, not an error — callers must be able to
	// distinguish "nothing matched" from "search unavailable" (below).
	res, err := client.Messages.SemanticSearch(context.Background(), SemanticSearchParams{
		Query: "anything",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(res.Results))
	}
}

// TestMessagesService_SemanticSearch_ProviderOutage asserts that an
// embedding-provider outage surfaces as a loud 5xx error, NOT as an empty
// result set — silently returning [] on outage is exactly the failure mode
// this endpoint used to have server-side.
func TestMessagesService_SemanticSearch_ProviderOutage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/search/semantic", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Embedding provider unavailable",
				"code":    "SERVICE_UNAVAILABLE",
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	_, err := client.Messages.SemanticSearch(context.Background(), SemanticSearchParams{
		Query: "anything",
	})
	if err == nil {
		t.Fatal("expected an error on provider outage, got nil")
	}
	if !errors.Is(err, ErrInternalServer) {
		t.Errorf("expected ErrInternalServer, got %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", apiErr.Status)
	}
}
