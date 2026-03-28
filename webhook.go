package anima

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultWebhookTolerance is the default maximum age of a webhook event (5 minutes).
	DefaultWebhookTolerance = 5 * time.Minute
)

// WebhookEvent represents a parsed and verified webhook event from Anima.
type WebhookEvent struct {
	ID        string                 `json:"id,omitempty"`
	Type      string                 `json:"type"`
	Data      map[string]any         `json:"data"`
	CreatedAt string                 `json:"createdAt,omitempty"`
}

// WebhookVerifyOptions configures webhook signature verification.
type WebhookVerifyOptions struct {
	// Tolerance is the maximum allowed age of the webhook timestamp.
	// Defaults to DefaultWebhookTolerance (5 minutes).
	Tolerance time.Duration
	// Now overrides the current time (useful for testing).
	Now time.Time
}

// VerifyWebhookSignature verifies that a webhook payload was signed by Anima.
// The signatureHeader should be in the format "t=<unix_timestamp>,v1=<hex_signature>".
// Returns true if the signature is valid and the timestamp is within tolerance.
func VerifyWebhookSignature(payload []byte, signatureHeader, secret string, opts *WebhookVerifyOptions) (bool, error) {
	timestamp, signature, err := parseSignatureHeader(signatureHeader)
	if err != nil {
		return false, err
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

	// Check timestamp age.
	age := math.Abs(float64(now.Unix() - timestamp))
	if age > tolerance.Seconds() {
		return false, nil
	}

	// Build signed payload: "<timestamp>.<payload>".
	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))

	// Compute expected HMAC-SHA256 signature.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expected := mac.Sum(nil)

	actual, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("%w: invalid hex in signature", ErrWebhookInvalid)
	}

	if len(expected) != len(actual) {
		return false, nil
	}

	return hmac.Equal(expected, actual), nil
}

// ConstructWebhookEvent verifies the signature and parses the payload into a
// WebhookEvent. Returns an error if the signature is invalid or the payload
// cannot be parsed.
func ConstructWebhookEvent(payload []byte, signatureHeader, secret string, opts *WebhookVerifyOptions) (*WebhookEvent, error) {
	valid, err := VerifyWebhookSignature(payload, signatureHeader, secret, opts)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("%w", ErrWebhookInvalid)
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, newValidationError("invalid webhook payload format", nil)
	}

	eventType, ok := raw["type"].(string)
	if !ok || eventType == "" {
		return nil, newValidationError("webhook payload missing event type", nil)
	}

	data, ok := raw["data"].(map[string]any)
	if !ok {
		return nil, newValidationError("webhook payload missing data object", nil)
	}

	event := &WebhookEvent{
		Type: eventType,
		Data: data,
	}

	if id, ok := raw["id"].(string); ok {
		event.ID = id
	}
	if createdAt, ok := raw["createdAt"].(string); ok {
		event.CreatedAt = createdAt
	}

	return event, nil
}

// parseSignatureHeader extracts timestamp and v1 signature from the header.
func parseSignatureHeader(header string) (int64, string, error) {
	var timestamp int64
	var signature string
	foundT := false
	foundV := false

	parts := strings.Split(header, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			v, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return 0, "", fmt.Errorf("%w: invalid timestamp in signature header", ErrWebhookInvalid)
			}
			timestamp = v
			foundT = true
		case "v1":
			signature = kv[1]
			foundV = true
		}
	}

	if !foundT || !foundV {
		return 0, "", fmt.Errorf("%w: invalid webhook signature header format", ErrWebhookInvalid)
	}

	return timestamp, signature, nil
}
