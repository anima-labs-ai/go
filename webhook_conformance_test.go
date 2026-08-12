package anima

// Cross-repo conformance: does this SDK verify what the platform actually signs?
//
// webhook_test.go reproduces the platform's signing scheme by hand, which is an
// improvement on what came before but still a copy — and a copy can drift back.
// This runs the monorepo's own signer and feeds its output to the SDK, so the
// two cannot disagree without something here going red.
//
// That is the check that was missing. This SDK shipped a Stripe-style
// "t=<unix>,v1=<hex>" header and a {"type": ..., "data": {...}} envelope; the
// platform has always sent "X-Anima-Signature: v1=<hex>" with a separate
// ISO-8601 "X-Anima-Timestamp", over a flat payload. Both sides had passing
// tests. Only something spanning them could have caught it.
//
// Go cannot import TypeScript, so this executes webhook-signature.ts with a JS
// runtime. That module is pure — no database, no server boot — so nothing else
// is needed.
//
// Skips when the monorepo or the runtime is absent, which is the case in CI.
// Deliberately loud about it: a silent skip here would restore the exact blind
// spot this test exists to close.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// signerPath resolves the monorepo's webhook signer. The SDK ships as its own
// repo but in dev sits next to the monorepo.
func signerPath() string {
	if override := os.Getenv("ANIMA_WEBHOOK_SIGNER_PATH"); override != "" {
		return override
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, "..", "anima", "apps", "api", "src", "services", "webhook-signature.ts")
}

// platformHeaders runs buildWebhookSignatureHeaders from the monorepo and
// returns exactly what it produced.
func platformHeaders(t *testing.T, secret, body, timestamp string) map[string]string {
	t.Helper()

	runtime, err := exec.LookPath("bun")
	if err != nil {
		t.Skipf("SKIPPED — no `bun` on PATH to execute the monorepo's TypeScript signer. "+
			"This is the only cross-repo check that the SDK verifies what the platform "+
			"signs. Original error: %v", err)
	}

	path := signerPath()
	if _, err := os.Stat(path); err != nil {
		t.Skipf("SKIPPED — monorepo signer not found at %s. This is the only cross-repo "+
			"check that the SDK verifies what the platform signs. Set "+
			"ANIMA_WEBHOOK_SIGNER_PATH to run it.", path)
	}

	args, err := json.Marshal([]string{path, secret, body, timestamp})
	if err != nil {
		t.Fatalf("marshal script args: %v", err)
	}

	script := fmt.Sprintf(
		"const [p, s, b, ts] = %s;"+
			"const m = await import(p);"+
			"console.log(JSON.stringify(m.buildWebhookSignatureHeaders(s, b, ts)));",
		string(args),
	)

	out, err := exec.Command(runtime, "-e", script).Output()
	if err != nil {
		t.Fatalf("running the monorepo signer failed: %v", err)
	}

	var headers map[string]string
	if err := json.Unmarshal(out, &headers); err != nil {
		t.Fatalf("decoding signer output %q: %v", string(out), err)
	}
	return headers
}

func TestConformance_SDKVerifiesPlatformSignedDelivery(t *testing.T) {
	headers := platformHeaders(t, testSecret, string(testPayload), testSignedAt)

	valid, err := VerifyWebhookSignature(testPayload,
		WebhookHeaders{
			Signature: headers["X-Anima-Signature"],
			Timestamp: headers["X-Anima-Timestamp"],
		},
		testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("the SDK rejected headers the platform produced")
	}
}

func TestConformance_PlatformEmitsExactlyTheHeadersTheSDKReads(t *testing.T) {
	headers := platformHeaders(t, testSecret, string(testPayload), testSignedAt)

	if len(headers) != 2 {
		t.Errorf("expected exactly 2 headers, got %d: %v", len(headers), headers)
	}
	sig, ok := headers["X-Anima-Signature"]
	if !ok {
		t.Fatal("missing X-Anima-Signature")
	}
	if len(sig) != len("v1=")+64 || sig[:3] != "v1=" {
		t.Errorf("X-Anima-Signature = %q, want v1=<64 hex>", sig)
	}
	if headers["X-Anima-Timestamp"] != testSignedAt {
		t.Errorf("X-Anima-Timestamp = %q, want %q", headers["X-Anima-Timestamp"], testSignedAt)
	}
}

func TestConformance_ConstructParsesPlatformSignedDelivery(t *testing.T) {
	headers := platformHeaders(t, testSecret, string(testPayload), testSignedAt)

	event, err := ConstructWebhookEvent(testPayload,
		WebhookHeaders{
			Signature: headers["X-Anima-Signature"],
			Timestamp: headers["X-Anima-Timestamp"],
		},
		testSecret,
		&WebhookVerifyOptions{Now: signedAt(t)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.Event != WebhookEventMessageReceived {
		t.Errorf("Event = %q", event.Event)
	}
	if event.Fields["messageId"] != "cme9x2k1p0001s601abcdefgh" {
		t.Errorf("Fields[messageId] = %v", event.Fields["messageId"])
	}
	if _, present := event.Fields["data"]; present {
		t.Error("payload should be flat — no data envelope")
	}
}
