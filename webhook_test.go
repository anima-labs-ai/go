package anima

// These tests exist because the previous ones passed while the code could not
// process a single real delivery.
//
// The old fixture built a Stripe-style "t=<unix>,v1=<hex>" header and a
// {"type": ..., "data": {...}} payload — both invented to match the
// implementation. Nothing compared either against what the platform sends, so
// the suite agreed with the bug and stayed green.
//
// So sign() below is written from the platform's published scheme rather than
// from this SDK's parser, and the payload is the real message.received shape.
// Sources, all of which agree with each other:
//
//   - apps/api/src/services/webhook-signature.ts — buildWebhookSignatureHeaders
//   - apps/api/src/workers/inbound-email.ts      — the emitted payload
//   - docs.useanima.sh/webhooks                  — the customer-facing contract
//
// The TestPreFixScheme_* cases at the bottom are the regression guards.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	testSecret   = "whsec_test_secret"
	testSignedAt = "2026-07-28T12:00:00.000Z"
)

// testPayload is a real message.received delivery: flat, no data envelope.
var testPayload = []byte(`{"event":"message.received","occurredAt":"2026-07-28T12:00:00.000Z",` +
	`"messageId":"cme9x2k1p0001s601abcdefgh","agentId":"cme9x2k1p0000s601ijklmnop",` +
	`"channel":"email","direction":"INBOUND","fromAddress":"user@example.com",` +
	`"toAddress":"support-agent@agents.useanima.sh","threadId":"cme9x2k1p0002s601qrstuvwx",` +
	`"subject":"Hello","spam":false}`)

func signedAt(t *testing.T) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, testSignedAt)
	if err != nil {
		t.Fatalf("bad test timestamp: %v", err)
	}
	return parsed
}

// sign reproduces the platform's signer: HMAC-SHA256 over "{iso}.{body}", hex,
// presented as "v1=<hex>" in its own header.
func sign(payload []byte, timestamp, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + string(payload)))
	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}

func testHeaders(payload []byte) WebhookHeaders {
	return WebhookHeaders{
		Signature: sign(payload, testSignedAt, testSecret),
		Timestamp: testSignedAt,
	}
}

func TestVerifyWebhookSignature_AcceptsGenuineDelivery(t *testing.T) {
	valid, err := VerifyWebhookSignature(testPayload, testHeaders(testPayload), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected a genuine delivery to verify")
	}
}

func TestVerifyWebhookSignature_AcceptsBareHexDigest(t *testing.T) {
	headers := testHeaders(testPayload)
	headers.Signature = digestFromHeader(headers.Signature)

	valid, err := VerifyWebhookSignature(testPayload, headers, testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected a bare hex digest to verify")
	}
}

func TestVerifyWebhookSignature_RejectsTamperedBody(t *testing.T) {
	tampered := []byte(strings.Replace(string(testPayload),
		"user@example.com", "attacker@evil.com", 1))

	valid, err := VerifyWebhookSignature(tampered, testHeaders(testPayload), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected a tampered body to be rejected")
	}
}

func TestVerifyWebhookSignature_RejectsWrongSecret(t *testing.T) {
	valid, err := VerifyWebhookSignature(testPayload, testHeaders(testPayload), "whsec_other",
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected the wrong secret to be rejected")
	}
}

func TestVerifyWebhookSignature_RejectsStaleDelivery(t *testing.T) {
	valid, err := VerifyWebhookSignature(testPayload, testHeaders(testPayload), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t).Add(6 * time.Minute), Tolerance: 5 * time.Minute})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected a stale delivery to be rejected")
	}
}

func TestVerifyWebhookSignature_RejectsEditedTimestampReplay(t *testing.T) {
	// The captured signature is valid only for the timestamp it signed, so
	// moving the clock forward breaks the MAC, not merely the freshness check.
	replayedAt := signedAt(t).Add(time.Hour)
	headers := testHeaders(testPayload)
	headers.Timestamp = replayedAt.Format(time.RFC3339)

	valid, err := VerifyWebhookSignature(testPayload, headers, testSecret,
		&WebhookVerifyOptions{Now: replayedAt})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected an edited timestamp to invalidate the signature")
	}
}

func TestVerifyWebhookSignature_RejectsUnparseableTimestamp(t *testing.T) {
	headers := WebhookHeaders{Signature: sign(testPayload, "not-a-date", testSecret), Timestamp: "not-a-date"}

	_, err := VerifyWebhookSignature(testPayload, headers, testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if !errors.Is(err, ErrWebhookInvalid) {
		t.Errorf("expected ErrWebhookInvalid, got %v", err)
	}
}

func TestConstructWebhookEvent_ReturnsDeliveryFlat(t *testing.T) {
	event, err := ConstructWebhookEvent(testPayload, testHeaders(testPayload), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.Event != WebhookEventMessageReceived {
		t.Errorf("Event = %q, want %q", event.Event, WebhookEventMessageReceived)
	}
	if event.OccurredAt != testSignedAt {
		t.Errorf("OccurredAt = %q, want %q", event.OccurredAt, testSignedAt)
	}
	if got := event.Fields["messageId"]; got != "cme9x2k1p0001s601abcdefgh" {
		t.Errorf("Fields[messageId] = %v", got)
	}
	if got := event.Fields["channel"]; got != "email" {
		t.Errorf("Fields[channel] = %v", got)
	}
	if _, present := event.Fields["data"]; present {
		t.Error("payload should be flat — no data envelope")
	}
}

func TestConstructWebhookEvent_UnknownEventNameIsNotAnError(t *testing.T) {
	// A new event on the platform must not break a deployed consumer.
	body := []byte(`{"event":"widget.exploded","occurredAt":"2026-07-28T12:00:00.000Z"}`)

	event, err := ConstructWebhookEvent(body, testHeaders(body), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Event != WebhookEventType("widget.exploded") {
		t.Errorf("Event = %q", event.Event)
	}
}

func TestConstructWebhookEvent_RejectsInvalidSignature(t *testing.T) {
	headers := WebhookHeaders{Signature: "v1=deadbeef", Timestamp: testSignedAt}

	_, err := ConstructWebhookEvent(testPayload, headers, testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if !errors.Is(err, ErrWebhookInvalid) {
		t.Errorf("expected ErrWebhookInvalid, got %v", err)
	}
}

func TestConstructWebhookEvent_RejectsMissingEventName(t *testing.T) {
	body := []byte(`{"occurredAt":"2026-07-28T12:00:00.000Z"}`)

	_, err := ConstructWebhookEvent(body, testHeaders(body), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err == nil {
		t.Fatal("expected an error for a payload with no event name")
	}
}

func TestConstructWebhookEvent_RejectsMissingOccurredAt(t *testing.T) {
	body := []byte(`{"event":"message.sent"}`)

	_, err := ConstructWebhookEvent(body, testHeaders(body), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err == nil {
		t.Fatal("expected an error for a payload with no occurredAt")
	}
}

func TestWebhookEvent_RoundTripsThroughJSON(t *testing.T) {
	event, err := ConstructWebhookEvent(testPayload, testHeaders(testPayload), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	encoded, err := event.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded WebhookEvent
	if err := decoded.UnmarshalJSON(encoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Event != event.Event || decoded.OccurredAt != event.OccurredAt {
		t.Errorf("round trip lost fields: %+v", decoded)
	}
}

func TestPreFixScheme_CombinedHeaderRejected(t *testing.T) {
	// What this SDK used to expect. The platform has never sent it: the
	// timestamp travels in X-Anima-Timestamp and the signature header holds
	// only "v1=<hex>".
	unix := signedAt(t).Unix()
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(fmt.Sprintf("%d.%s", unix, string(testPayload))))
	old := fmt.Sprintf("t=%d,v1=%s", unix, hex.EncodeToString(mac.Sum(nil)))

	headers := WebhookHeaders{Signature: old, Timestamp: testSignedAt}

	valid, err := VerifyWebhookSignature(testPayload, headers, testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err == nil && valid {
		t.Error("expected the pre-fix combined header to be rejected")
	}
}

func TestPreFixScheme_TypeDataEnvelopeRejected(t *testing.T) {
	// A correctly signed body in the old envelope shape still has no "event",
	// so it cannot be mistaken for a delivery.
	body := []byte(`{"id":"evt_1","type":"message.sent","createdAt":"2026-07-28T12:00:00.000Z","data":{"messageId":"m1"}}`)

	_, err := ConstructWebhookEvent(body, testHeaders(body), testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err == nil {
		t.Error("expected the pre-fix type/data envelope to be rejected")
	}
}
