package anima

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
