package anima

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// The org-wide number list and the SMS conversation surface, added when the
// phone contracts gained listIdentities and smsThreadStats.

func TestPhonesService_ListIdentities(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/phone/identities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("query"); got != "555-123" {
			t.Errorf("expected query '555-123', got %q", got)
		}
		if got := r.URL.Query().Get("agentId"); got != "agent_1" {
			t.Errorf("expected agentId 'agent_1', got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{
				"id": "pi_1", "phoneNumber": "+15551234567", "providerId": nil,
				"capabilities": map[string]bool{"sms": true, "mms": false, "voice": true},
				"tenDlcStatus": "UNREGISTERED", "isPrimary": true, "voiceId": nil,
				"createdAt": "2026-08-20T00:00:00Z",
				"agentId":   "agent_1", "agentName": "Support", "agentSlug": "support",
			}},
			"pagination": map[string]any{"nextCursor": nil, "hasMore": false},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Phones.ListIdentities(context.Background(),
		&ListPhoneIdentitiesParams{Query: "555-123", AgentID: "agent_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page.Items))
	}
	// The embedded PhoneIdentity and the joined agent fields must both decode:
	// this list exists precisely so a caller need not join agents by hand.
	if page.Items[0].PhoneNumber != "+15551234567" {
		t.Errorf("expected embedded PhoneNumber to decode, got %q", page.Items[0].PhoneNumber)
	}
	if page.Items[0].AgentName != "Support" {
		t.Errorf("expected AgentName 'Support', got %q", page.Items[0].AgentName)
	}
}

// The thread list is offset-paged and answers {items, total, hasMore}. Decoding
// it as a Page[T] would leave HasMore at its zero value and report the first
// page as the whole list, so the envelope is asserted directly.
func TestPhonesService_ListSMSThreads(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/phone/sms/threads", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("agentId"); got != "agent_1" {
			t.Errorf("expected agentId 'agent_1', got %q", got)
		}
		if got := q.Get("limit"); got != "20" {
			t.Errorf("expected limit '20', got %q", got)
		}
		if got := q.Get("offset"); got != "40" {
			t.Errorf("expected offset '40', got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{
				"threadId": "msg_1", "agentId": "agent_1",
				"participantAddress": "+15550001", "agentAddress": "+15551234567",
				"lastMessageAt": "2026-08-20T00:00:00Z", "lastMessageSnippet": "hi",
				"lastMessageDirection": "INBOUND", "messageCount": 3, "unreadCount": 1,
			}},
			"total": 57, "hasMore": true,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	list, err := client.Phones.ListSMSThreads(context.Background(),
		&SMSThreadListParams{AgentID: "agent_1", Limit: 20, Offset: 40})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !list.HasMore {
		t.Error("expected HasMore true — a false here is how an offset list silently truncates")
	}
	if list.Total != 57 {
		t.Errorf("expected Total 57, got %d", list.Total)
	}
	if list.Items[0].LastMessageDirection != SMSMessageDirectionInbound {
		t.Errorf("expected INBOUND, got %q", list.Items[0].LastMessageDirection)
	}
}

// Unread is *bool so an explicit false is representable. The API gates on
// `params.unread ? ...` and treats absent and false alike today, so this pins
// the encoding — what the caller asked for reaches the wire — rather than a
// live bug.
func TestPhonesService_ListSMSThreads_ExplicitUnreadFalse(t *testing.T) {
	for _, tc := range []struct {
		name   string
		unread *bool
		want   string
	}{
		{"unset sends nothing", nil, ""},
		{"explicit false is sent", boolPtr(false), "false"},
		{"explicit true is sent", boolPtr(true), "true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/phone/sms/threads", func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("unread"); got != tc.want {
					t.Errorf("expected unread %q, got %q", tc.want, got)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"items": []any{}, "total": 0, "hasMore": false})
			})
			client, ts := newTestClient(mux)
			defer ts.Close()
			if _, err := client.Phones.ListSMSThreads(context.Background(),
				&SMSThreadListParams{Unread: tc.unread}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }

func TestPhonesService_GetSMSThread(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/phone/sms/threads/msg_1", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "25" {
			t.Errorf("expected limit '25', got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"threadId": "msg_1", "agentId": "agent_1",
			"participantAddress": "+15550001", "agentAddress": "+15551234567",
			"messages": []any{}, "messageCount": 3, "hasMore": true,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	thread, err := client.Phones.GetSMSThread(context.Background(), "msg_1", &SMSThreadGetParams{Limit: 25})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if thread.ThreadID != "msg_1" || !thread.HasMore {
		t.Errorf("unexpected thread: %+v", thread)
	}
}

func TestPhonesService_SMSThreadStats(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/phone/sms/stats", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("agentId"); got != "agent_1" {
			t.Errorf("expected agentId 'agent_1', got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"agentId": "agent_1", "conversations": 4, "unread": 2, "lastMessageAt": "2026-08-20T00:00:00Z"},
				// lastMessageAt is nullable: an agent with no messages reports null.
				{"agentId": "agent_2", "conversations": 0, "unread": 0, "lastMessageAt": nil},
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	stats, err := client.Phones.SMSThreadStats(context.Background(), &SMSThreadStatsParams{AgentID: "agent_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Items[0].Unread != 2 {
		t.Errorf("expected Unread 2, got %d", stats.Items[0].Unread)
	}
	if stats.Items[1].LastMessageAt != nil {
		t.Errorf("expected nil LastMessageAt, got %v", *stats.Items[1].LastMessageAt)
	}
}

func TestPhonesService_SMSSuppressions(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/phone/sms-suppressions", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("phoneNumber"); got != "+15550001" {
			t.Errorf("expected phoneNumber '+15550001', got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{
				"id": "sup_1", "phoneNumber": "+15550001", "agentId": nil,
				"reason": "STOP_KEYWORD", "source": "inbound-stop-keyword",
				"createdAt": "2026-08-20T00:00:00Z",
			}},
			"pagination": map[string]any{"nextCursor": nil, "hasMore": false},
		})
	})
	mux.HandleFunc("/v1/phone/sms-unsuppress", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["phoneNumber"] != "+15550001" {
			t.Errorf("expected phoneNumber in body, got %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"phoneNumber": "+15550001", "removed": 2})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	page, err := client.Phones.ListSMSSuppressions(context.Background(),
		&SMSSuppressionListParams{PhoneNumber: "+15550001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// agentId is null for a workspace-wide suppression, which is the common
	// case — a STOP applies to the org, not to whichever agent was texting.
	if page.Items[0].AgentID != nil {
		t.Errorf("expected nil AgentID for a workspace-wide suppression, got %v", *page.Items[0].AgentID)
	}
	if page.Items[0].Reason != SMSSuppressionReasonStopKeyword {
		t.Errorf("expected STOP_KEYWORD, got %q", page.Items[0].Reason)
	}

	result, err := client.Phones.UnsuppressSMS(context.Background(),
		UnsuppressSMSParams{PhoneNumber: "+15550001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Removed != 2 {
		t.Errorf("expected Removed 2, got %d", result.Removed)
	}
}
