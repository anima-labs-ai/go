package anima

// Webhook verification, matching what the platform actually sends.
//
// Scheme "v1": HMAC-SHA256, hex-encoded, over "{timestamp}.{rawBody}", where
// timestamp is the ISO-8601 string from X-Anima-Timestamp. Two headers travel
// with every delivery:
//
//	X-Anima-Signature:  v1=<hex>
//	X-Anima-Timestamp:  <ISO-8601>
//
// The timestamp is inside the signed content rather than merely alongside it,
// which is what makes replay rejection possible: a captured delivery fails the
// freshness window, and editing the timestamp to get past it invalidates the
// MAC.
//
// This file previously implemented a Stripe-style single header
// ("t=<unix>,v1=<hex>") and MAC'd over a unix-seconds timestamp, then required
// a {"type": ..., "data": {...}} envelope. Nothing the platform sends has ever
// matched either, so both exported functions failed on every real delivery. The
// tests did not catch it because they built their own fixture from the same
// wrong assumptions. See webhook_test.go, which now derives its fixture from the
// platform's published scheme, and webhook_conformance_test.go, which checks
// this SDK against the monorepo's own signer.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	// DefaultWebhookTolerance is the default maximum age of a webhook event (5 minutes).
	DefaultWebhookTolerance = 5 * time.Minute

	signatureVersion = "v1"
)

// WebhookHeaders carries the two signature headers that arrive with every
// delivery. Both are required: the timestamp is part of the signed content, so
// a receiver that reads only the signature cannot recompute the MAC.
type WebhookHeaders struct {
	// Signature is the X-Anima-Signature header, "v1=<hex>". A bare hex digest
	// is also accepted.
	Signature string
	// Timestamp is the X-Anima-Timestamp header, ISO-8601.
	Timestamp string
}

// WebhookEvent is a delivered webhook, as it arrives on the wire.
//
// The payload is flat. There is no "data" envelope: Event and OccurredAt sit
// alongside the event's own fields, so a message.received delivery also carries
// messageId, agentId, channel, direction, fromAddress, toAddress, threadId and
// — for email — subject and spam.
//
// Fields holds the whole decoded payload, including "event" and "occurredAt",
// keyed by their wire names. The message body is not included; fetch
// GET /v1/messages/{id} when you need content.
type WebhookEvent struct {
	Event      WebhookEventType
	OccurredAt string
	Fields     map[string]any
}

// UnmarshalJSON decodes a flat delivery payload.
func (e *WebhookEvent) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	name, _ := raw["event"].(string)
	occurredAt, _ := raw["occurredAt"].(string)
	e.Event = WebhookEventType(name)
	e.OccurredAt = occurredAt
	e.Fields = raw
	return nil
}

// MarshalJSON re-encodes the payload as received.
func (e WebhookEvent) MarshalJSON() ([]byte, error) {
	if e.Fields != nil {
		return json.Marshal(e.Fields)
	}
	return json.Marshal(map[string]any{
		"event":      string(e.Event),
		"occurredAt": e.OccurredAt,
	})
}

// WebhookVerifyOptions configures webhook signature verification.
type WebhookVerifyOptions struct {
	// Tolerance is the maximum allowed age of the webhook timestamp.
	// Defaults to DefaultWebhookTolerance (5 minutes).
	Tolerance time.Duration
	// Now overrides the current time (useful for testing).
	Now time.Time
}

// VerifyWebhookSignature verifies that a webhook payload was signed by Anima
// and that the delivery is fresh.
//
// Pass the raw request body. Verifying a re-serialised body fails even for a
// genuine delivery, because re-encoding can reorder keys or change whitespace
// and the MAC covers bytes.
func VerifyWebhookSignature(payload []byte, headers WebhookHeaders, secret string, opts *WebhookVerifyOptions) (bool, error) {
	signedAt, err := time.Parse(time.RFC3339, headers.Timestamp)
	if err != nil {
		return false, fmt.Errorf("%w: X-Anima-Timestamp is not RFC3339", ErrWebhookInvalid)
	}

	tolerance := DefaultWebhookTolerance
	now := time.Now()
	if opts != nil {
		if opts.Tolerance > 0 {
			tolerance = opts.Tolerance
		}
		if !opts.Now.IsZero() {
			now = opts.Now
		}
	}

	if age := now.Sub(signedAt); age > tolerance || age < -tolerance {
		return false, nil
	}

	// The signed content pairs the timestamp exactly as it travelled, so an
	// attacker who edits the header to look fresh invalidates the MAC.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(headers.Timestamp + "." + string(payload)))
	expected := mac.Sum(nil)

	actual, err := hex.DecodeString(digestFromHeader(headers.Signature))
	if err != nil {
		return false, fmt.Errorf("%w: invalid hex in signature", ErrWebhookInvalid)
	}

	if len(expected) != len(actual) {
		return false, nil
	}

	return hmac.Equal(expected, actual), nil
}

// ConstructWebhookEvent verifies the signature and parses the payload.
//
// The payload is flat, so the returned event is the delivery: Event and
// OccurredAt, with every field available in Fields.
func ConstructWebhookEvent(payload []byte, headers WebhookHeaders, secret string, opts *WebhookVerifyOptions) (*WebhookEvent, error) {
	valid, err := VerifyWebhookSignature(payload, headers, secret, opts)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("%w", ErrWebhookInvalid)
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, newValidationError(400, "invalid webhook payload format", nil)
	}

	name, ok := raw["event"].(string)
	if !ok || name == "" {
		return nil, newValidationError(400, "webhook payload missing event name", nil)
	}

	occurredAt, ok := raw["occurredAt"].(string)
	if !ok || occurredAt == "" {
		return nil, newValidationError(400, "webhook payload missing occurredAt", nil)
	}

	return &WebhookEvent{
		Event:      WebhookEventType(name),
		OccurredAt: occurredAt,
		Fields:     raw,
	}, nil
}

// digestFromHeader strips the "v1=" prefix. A bare hex digest passes through.
func digestFromHeader(signature string) string {
	return strings.TrimPrefix(signature, signatureVersion+"=")
}
