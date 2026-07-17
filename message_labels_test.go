package anima

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
)

// Spec B3 — labels + read state.
//
// Labels are the agent's workflow state machine: without them every list
// returns the same undifferentiated stream forever. anima#307 shipped them
// server-side; these tests pin the client half — that a label filter reaches
// the API in the ONE shape it reads, and that UpdateLabels cannot report
// success for a call that changes nothing.
//
// The URL shape carries the whole feature. url.Values.Set would keep only the
// last label, quietly turning "urgent AND unread" into "unread" and returning
// MORE mail than the caller asked for. Nothing in the type system can see that
// — only a test that reads the raw query can.

func TestMessageListParams_LabelsBecomeRepeatedKeys(t *testing.T) {
	params := MessageListParams{Labels: []string{"urgent", "unread"}}

	got := params.ToQuery()["labels"]
	want := []string{"urgent", "unread"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected one query key per label %v, got %v — Set instead of Add would drop all but the last", want, got)
	}
	// The encoded form is what actually goes on the wire.
	if enc := params.ToQuery().Encode(); enc != "labels=urgent&labels=unread" {
		t.Errorf("expected repeated-key encoding, got %q", enc)
	}
}

func TestMessageListParams_SingleLabelIsALoneValue(t *testing.T) {
	// `?labels=unread` — the most common label query there is — 400'd until
	// anima#309 taught the contract to accept a lone value.
	if enc := (MessageListParams{Labels: []string{"unread"}}).ToQuery().Encode(); enc != "labels=unread" {
		t.Errorf("expected labels=unread, got %q", enc)
	}
}

func TestMessageListParams_IncludeSpamFalseIsTransmitted(t *testing.T) {
	no := false
	q := MessageListParams{IncludeSpam: &no}.ToQuery()

	// An explicit false is the caller overriding. A plain bool field would be
	// indistinguishable from unset here and the override would vanish.
	if got := q.Get("includeSpam"); got != "false" {
		t.Errorf("expected includeSpam=false to reach the wire, got %q", got)
	}

	yes := true
	if got := (MessageListParams{IncludeSpam: &yes}).ToQuery().Get("includeSpam"); got != "true" {
		t.Errorf("expected includeSpam=true, got %q", got)
	}
}

func TestMessageListParams_NoLabelParamsMeansNoLabelKeys(t *testing.T) {
	q := MessageListParams{}.ToQuery()
	if _, ok := q["labels"]; ok {
		t.Error("empty Labels must not emit a labels key")
	}
	if _, ok := q["includeSpam"]; ok {
		t.Error("nil IncludeSpam must not emit an includeSpam key")
	}
}

func TestMessagesService_UpdateLabels(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/msg_123/labels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		// id rides in the body as well as the path — the contract requires it.
		if body["id"] != "msg_123" {
			t.Errorf("expected id in body, got %v", body["id"])
		}
		add, ok := body["addLabels"].([]interface{})
		if !ok || len(add) != 1 || add[0] != "read" {
			t.Errorf("expected addLabels [read], got %v", body["addLabels"])
		}
		remove, ok := body["removeLabels"].([]interface{})
		if !ok || len(remove) != 1 || remove[0] != "unread" {
			t.Errorf("expected removeLabels [unread], got %v", body["removeLabels"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_123","agentId":"agent_1","channel":"EMAIL","direction":"INBOUND","status":"DELIVERED","fromAddress":"a@b.c","toAddress":"d@e.f","body":"hi","labels":["read"],"attachments":[],"createdAt":"2026-07-17T00:00:00Z","updatedAt":"2026-07-17T00:00:00Z"}`))
	})
	client, ts := newTestClient(mux)
	defer ts.Close()

	msg, err := client.Messages.UpdateLabels(context.Background(), "msg_123", UpdateLabelsParams{
		AddLabels:    []string{"read"},
		RemoveLabels: []string{"unread"},
	})
	if err != nil {
		t.Fatalf("UpdateLabels returned error: %v", err)
	}
	// The returned message must carry the new labels, so callers never guess.
	if !reflect.DeepEqual(msg.Labels, []string{"read"}) {
		t.Errorf("expected labels [read], got %v", msg.Labels)
	}
}

func TestMessagesService_UpdateLabels_OmitsTheOperationNotAskedFor(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/msg_123/labels", func(w http.ResponseWriter, r *http.Request) {
		body := decodeRawBody(t, r)
		// omitempty must drop the absent operation: "leave the rest alone" is
		// the absent key, not an empty array the server still has to process.
		if _, present := body["removeLabels"]; present {
			t.Errorf("removeLabels must be absent when not supplied, got %v", body["removeLabels"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_123","agentId":"a","channel":"EMAIL","direction":"INBOUND","status":"DELIVERED","fromAddress":"a@b.c","toAddress":"d@e.f","body":"hi","labels":["archived"],"attachments":[],"createdAt":"2026-07-17T00:00:00Z","updatedAt":"2026-07-17T00:00:00Z"}`))
	})
	client, ts := newTestClient(mux)
	defer ts.Close()

	if _, err := client.Messages.UpdateLabels(context.Background(), "msg_123", UpdateLabelsParams{
		AddLabels: []string{"archived"},
	}); err != nil {
		t.Fatalf("UpdateLabels returned error: %v", err)
	}
}

func TestMessagesService_UpdateLabels_EmptyCallIsRefusedBeforeTheWire(t *testing.T) {
	mux := http.NewServeMux()
	called := false
	mux.HandleFunc("/v1/messages/msg_123/labels", func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	client, ts := newTestClient(mux)
	defer ts.Close()

	// The failure this prevents: an agent "marks the message read", gets a
	// success back, and the labels never changed.
	_, err := client.Messages.UpdateLabels(context.Background(), "msg_123", UpdateLabelsParams{})
	if !errors.Is(err, ErrNoLabelChanges) {
		t.Errorf("expected ErrNoLabelChanges, got %v", err)
	}
	if called {
		t.Error("no request must be sent for a no-op label update")
	}
}

func TestMessageSearchFilters_CarryLabels(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/search", func(w http.ResponseWriter, r *http.Request) {
		body := decodeRawBody(t, r)
		filters, ok := body["filters"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected filters object, got %v", body["filters"])
		}
		labels, ok := filters["labels"].([]interface{})
		if !ok || len(labels) != 1 || labels[0] != "unread" {
			t.Errorf("expected filters.labels [unread], got %v", filters["labels"])
		}
		// A plain bool with omitempty would silently drop false here.
		if filters["includeSpam"] != false {
			t.Errorf("expected filters.includeSpam false to survive omitempty, got %v", filters["includeSpam"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"pagination":{"nextCursor":null,"hasMore":false}}`))
	})
	client, ts := newTestClient(mux)
	defer ts.Close()

	no := false
	if _, err := client.Messages.Search(context.Background(), MessageSearchParams{
		Query:   "invoice",
		Filters: &MessageSearchFilters{Labels: []string{"unread"}, IncludeSpam: &no},
	}); err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
}

func TestMessage_LabelsParseOffTheWire(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages/msg_1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_1","agentId":"a","channel":"EMAIL","direction":"INBOUND","status":"DELIVERED","fromAddress":"a@b.c","toAddress":"d@e.f","body":"hi","labels":["unread","urgent"],"attachments":[],"createdAt":"2026-07-17T00:00:00Z","updatedAt":"2026-07-17T00:00:00Z"}`))
	})
	client, ts := newTestClient(mux)
	defer ts.Close()

	msg, err := client.Messages.Get(context.Background(), "msg_1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !reflect.DeepEqual(msg.Labels, []string{"unread", "urgent"}) {
		t.Errorf("expected labels [unread urgent], got %v", msg.Labels)
	}
}
