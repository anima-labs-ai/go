package anima

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

func computeSignature(payload []byte, timestamp int64, secret string) string {
	signed := fmt.Sprintf("%d.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"type":"agent.created","data":{"id":"ag_123"}}`)
	now := time.Now()
	timestamp := now.Unix()
	sig := computeSignature(payload, timestamp, secret)
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, sig)

	valid, err := VerifyWebhookSignature(payload, header, secret, &WebhookVerifyOptions{Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected signature to be valid")
	}
}

func TestVerifyWebhookSignature_InvalidSignature(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"type":"agent.created","data":{"id":"ag_123"}}`)
	now := time.Now()
	timestamp := now.Unix()
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, "deadbeef00112233445566778899aabbccddeeff00112233445566778899aabbcc")

	valid, err := VerifyWebhookSignature(payload, header, secret, &WebhookVerifyOptions{Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected signature to be invalid")
	}
}

func TestVerifyWebhookSignature_ExpiredTimestamp(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"type":"agent.created","data":{"id":"ag_123"}}`)
	now := time.Now()
	oldTimestamp := now.Add(-10 * time.Minute).Unix()
	sig := computeSignature(payload, oldTimestamp, secret)
	header := fmt.Sprintf("t=%d,v1=%s", oldTimestamp, sig)

	valid, err := VerifyWebhookSignature(payload, header, secret, &WebhookVerifyOptions{
		Now:       now,
		Tolerance: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected expired timestamp to be invalid")
	}
}

func TestVerifyWebhookSignature_CustomTolerance(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"type":"test","data":{}}`)
	now := time.Now()
	// 8 minutes ago — would fail default 5 min tolerance but pass 10 min.
	timestamp := now.Add(-8 * time.Minute).Unix()
	sig := computeSignature(payload, timestamp, secret)
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, sig)

	valid, err := VerifyWebhookSignature(payload, header, secret, &WebhookVerifyOptions{
		Now:       now,
		Tolerance: 10 * time.Minute,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid with 10 minute tolerance")
	}
}

func TestVerifyWebhookSignature_MalformedHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"missing v1", "t=123456"},
		{"missing t", "v1=abcdef"},
		{"empty", ""},
		{"garbage", "foobar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VerifyWebhookSignature([]byte("{}"), tt.header, "secret", nil)
			if err == nil {
				t.Error("expected error for malformed header")
			}
			if !errors.Is(err, ErrWebhookInvalid) {
				t.Errorf("expected ErrWebhookInvalid, got %v", err)
			}
		})
	}
}

func TestConstructWebhookEvent_Valid(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"id":"evt_1","type":"agent.created","data":{"id":"ag_123","name":"TestBot"},"createdAt":"2024-01-01T00:00:00Z"}`)
	now := time.Now()
	timestamp := now.Unix()
	sig := computeSignature(payload, timestamp, secret)
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, sig)

	event, err := ConstructWebhookEvent(payload, header, secret, &WebhookVerifyOptions{Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Type != "agent.created" {
		t.Errorf("Type = %q, want agent.created", event.Type)
	}
	if event.ID != "evt_1" {
		t.Errorf("ID = %q, want evt_1", event.ID)
	}
	if event.Data["id"] != "ag_123" {
		t.Errorf("Data.id = %v, want ag_123", event.Data["id"])
	}
	if event.Data["name"] != "TestBot" {
		t.Errorf("Data.name = %v, want TestBot", event.Data["name"])
	}
	if event.CreatedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want 2024-01-01T00:00:00Z", event.CreatedAt)
	}
}

func TestConstructWebhookEvent_InvalidSignature(t *testing.T) {
	payload := []byte(`{"type":"test","data":{}}`)
	now := time.Now()
	header := fmt.Sprintf("t=%d,v1=%s", now.Unix(), "0000000000000000000000000000000000000000000000000000000000000000")

	_, err := ConstructWebhookEvent(payload, header, "secret", &WebhookVerifyOptions{Now: now})
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
	if !errors.Is(err, ErrWebhookInvalid) {
		t.Errorf("expected ErrWebhookInvalid, got %v", err)
	}
}

func TestConstructWebhookEvent_MissingType(t *testing.T) {
	secret := "test"
	payload := []byte(`{"data":{"id":"123"}}`)
	now := time.Now()
	timestamp := now.Unix()
	sig := computeSignature(payload, timestamp, secret)
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, sig)

	_, err := ConstructWebhookEvent(payload, header, secret, &WebhookVerifyOptions{Now: now})
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestConstructWebhookEvent_MissingData(t *testing.T) {
	secret := "test"
	payload := []byte(`{"type":"test"}`)
	now := time.Now()
	timestamp := now.Unix()
	sig := computeSignature(payload, timestamp, secret)
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, sig)

	_, err := ConstructWebhookEvent(payload, header, secret, &WebhookVerifyOptions{Now: now})
	if err == nil {
		t.Fatal("expected error for missing data")
	}
}

func TestVerifyWebhookSignature_WrongSecret(t *testing.T) {
	payload := []byte(`{"type":"test","data":{}}`)
	now := time.Now()
	timestamp := now.Unix()
	sig := computeSignature(payload, timestamp, "correct_secret")
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, sig)

	valid, err := VerifyWebhookSignature(payload, header, "wrong_secret", &WebhookVerifyOptions{Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected invalid with wrong secret")
	}
}

func intPtr(v int) *int { return &v }

// marshalToMap round-trips v through JSON into a generic map for assertions.
func marshalToMap(t *testing.T, v any) map[string]any {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestCreateWebhookParams_MarshalAuthAndThrottle(t *testing.T) {
	params := CreateWebhookParams{
		URL:                "https://example.com/hook",
		Events:             []WebhookEventType{WebhookEventMessageReceived},
		AuthConfig:         NewBearerAuth("tok_secret"),
		RateLimitPerMinute: intPtr(120),
		MaxAttempts:        intPtr(5),
	}

	got := marshalToMap(t, params)

	auth, ok := got["authConfig"].(map[string]any)
	if !ok {
		t.Fatalf("authConfig missing or wrong type: %v", got["authConfig"])
	}
	if auth["type"] != "bearer" {
		t.Errorf("authConfig.type = %v, want bearer", auth["type"])
	}
	if auth["token"] != "tok_secret" {
		t.Errorf("authConfig.token = %v, want tok_secret", auth["token"])
	}
	// Non-bearer fields must be omitted from the marshaled auth config.
	for _, k := range []string{"username", "password", "headerName", "value"} {
		if _, present := auth[k]; present {
			t.Errorf("authConfig should omit %q for bearer, got %v", k, auth[k])
		}
	}
	if got["rateLimitPerMinute"] != float64(120) {
		t.Errorf("rateLimitPerMinute = %v, want 120", got["rateLimitPerMinute"])
	}
	if got["maxAttempts"] != float64(5) {
		t.Errorf("maxAttempts = %v, want 5", got["maxAttempts"])
	}
}

func TestWebhookAuthConfig_Variants(t *testing.T) {
	tests := []struct {
		name string
		cfg  *WebhookAuthConfig
		want map[string]any
	}{
		{"none", NewNoAuth(), map[string]any{"type": "none"}},
		{"bearer", NewBearerAuth("t"), map[string]any{"type": "bearer", "token": "t"}},
		{"basic", NewBasicAuth("u", "p"), map[string]any{"type": "basic", "username": "u", "password": "p"}},
		{"custom_header", NewCustomHeaderAuth("X-Key", "v"), map[string]any{"type": "custom_header", "headerName": "X-Key", "value": "v"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := marshalToMap(t, tt.cfg)
			if len(got) != len(tt.want) {
				t.Errorf("key count = %d, want %d (got %v)", len(got), len(tt.want), got)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("%s = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestCreateWebhookParams_OmitsAdvancedWhenUnset(t *testing.T) {
	params := CreateWebhookParams{
		URL:    "https://example.com/hook",
		Events: []WebhookEventType{WebhookEventMessageReceived},
	}
	got := marshalToMap(t, params)
	for _, k := range []string{"authConfig", "rateLimitPerMinute", "maxAttempts"} {
		if _, present := got[k]; present {
			t.Errorf("expected %q to be omitted when unset, got %v", k, got[k])
		}
	}
}

func TestUpdateWebhookParams_MarshalAuthAndThrottle(t *testing.T) {
	params := UpdateWebhookParams{
		AuthConfig:         NewNoAuth(),
		RateLimitPerMinute: intPtr(30),
		MaxAttempts:        intPtr(1),
	}
	got := marshalToMap(t, params)
	auth, ok := got["authConfig"].(map[string]any)
	if !ok || auth["type"] != "none" {
		t.Fatalf("authConfig = %v, want {type:none}", got["authConfig"])
	}
	if got["rateLimitPerMinute"] != float64(30) {
		t.Errorf("rateLimitPerMinute = %v, want 30", got["rateLimitPerMinute"])
	}
	if got["maxAttempts"] != float64(1) {
		t.Errorf("maxAttempts = %v, want 1", got["maxAttempts"])
	}
}

func TestWebhook_UnmarshalReadModel(t *testing.T) {
	raw := `{
		"id":"wh_1","orgId":"org_1","url":"https://example.com/hook",
		"events":["message.received"],"active":true,"description":null,
		"consecutiveFailures":0,"disabledReason":null,"disabledAt":null,
		"createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-01-01T00:00:00Z",
		"authType":"CUSTOM_HEADER","authHeaderName":"X-Key",
		"rateLimitPerMinute":60,"maxAttempts":5
	}`
	var wh Webhook
	if err := json.Unmarshal([]byte(raw), &wh); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if wh.AuthType != WebhookAuthTypeCustomHeader {
		t.Errorf("AuthType = %q, want CUSTOM_HEADER", wh.AuthType)
	}
	if wh.AuthHeaderName == nil || *wh.AuthHeaderName != "X-Key" {
		t.Errorf("AuthHeaderName = %v, want X-Key", wh.AuthHeaderName)
	}
	if wh.RateLimitPerMinute == nil || *wh.RateLimitPerMinute != 60 {
		t.Errorf("RateLimitPerMinute = %v, want 60", wh.RateLimitPerMinute)
	}
	if wh.MaxAttempts == nil || *wh.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %v, want 5", wh.MaxAttempts)
	}
}
