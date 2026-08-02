package anima_test

// Conformance probe against a DEPLOYED Anima API.
//
// Every other test in this package asserts against an httptest mux — a fixture
// written by whoever wrote the code. That is why 30 calls to routes the API has
// never served stayed green for months, and why VerifyCredential decoded to an
// all-zero struct: the mock invented the same wrong shape the struct expected.
//
// This is the half a mock cannot cover. Each probe calls a real read-only
// endpoint over HTTP and classifies the outcome:
//
//	outcome         verdict  why
//	--------------  -------  ------------------------------------------------
//	200 + shape ok  pass     route exists and the response looks as declared
//	401 / 403       pass     route exists; this key just lacks the scope
//	404             FAIL     the route is gone — the phantom-route class
//	400 / 422       FAIL     the API rejected OUR request: bad enum, bad param
//	5xx             FAIL     reported separately; usually not the SDK's fault
//
// The 400 row earns its keep: it catches a request the SDK built wrongly,
// which no mock will ever reject.
//
// Go's encoding/json ignores unknown fields and zero-fills missing ones, so
// decoding alone proves little. Where a probe asserts on the payload, that
// assertion IS the shape check — keep them.
//
// STRICTLY READ-ONLY. No probe creates, mutates or deletes anything: this runs
// against a real organization. Do not add a POST here.
//
//	ANIMA_LIVE_API_KEY=mk_...  # master key sees the most surface
//	ANIMA_LIVE_ORG_ID=org_...  # required for org-scoped probes
//	ANIMA_LIVE_BASE_URL=...    # optional, defaults to production
//	go test -run TestLive ./...
//
// Without ANIMA_LIVE_API_KEY every probe skips, so normal CI is unaffected.
//
// Package anima_test (not anima): this exercises the SDK exactly as a consumer
// does, through the exported surface only.

import (
	"context"
	"errors"
	"os"
	"testing"

	anima "github.com/anima-labs-ai/go"
)

func liveClient(t *testing.T) *anima.Client {
	t.Helper()
	key := os.Getenv("ANIMA_LIVE_API_KEY")
	if key == "" {
		t.Skip("live conformance probe: set ANIMA_LIVE_API_KEY to run")
	}
	var opts []anima.Option
	if base := os.Getenv("ANIMA_LIVE_BASE_URL"); base != "" {
		opts = append(opts, anima.WithBaseURL(base))
	}
	return anima.NewClient(key, opts...)
}

func liveOrgID(t *testing.T) string {
	t.Helper()
	orgID := os.Getenv("ANIMA_LIVE_ORG_ID")
	if orgID == "" {
		t.Skip("set ANIMA_LIVE_ORG_ID for org-scoped probes")
	}
	return orgID
}

// probe applies the verdict table above to one read-only call.
func probe(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	var apiErr *anima.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("transport-level failure (not an API response): %v", err)
	}
	switch {
	case apiErr.Status == 404:
		t.Errorf("404 — the API does not serve this route. Either the SDK path is "+
			"wrong or the endpoint was removed: %v", apiErr)
	case apiErr.Status == 400 || apiErr.Status == 422:
		t.Errorf("400/422 — the API rejected the request the SDK built. This is the "+
			"class of bug a mock cannot catch (wrong enum casing, wrong query param "+
			"name, wrong path param): %v", apiErr)
	case apiErr.Status == 401 || apiErr.Status == 403:
		// The route exists and is reachable; this key is simply not scoped for
		// it. A pass for conformance purposes.
	case apiErr.Status == 429:
		t.Skipf("rate limited by the live API: %v", apiErr)
	case apiErr.Status >= 500:
		t.Errorf("5xx from the live API (likely not the SDK's fault): %v", apiErr)
	default:
		t.Errorf("unexpected status %d: %v", apiErr.Status, apiErr)
	}
}

func TestLiveCoreSurface(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()

	t.Run("Agents.List", func(t *testing.T) {
		page, err := client.Agents.List(ctx, &anima.AgentListParams{ListParams: anima.ListParams{Limit: 1}})
		probe(t, err)
		if err == nil && page.Items == nil {
			t.Error("items was null — the envelope is not what the SDK declares")
		}
	})

	t.Run("Domains.List", func(t *testing.T) {
		_, err := client.Domains.List(ctx)
		probe(t, err)
	})

	t.Run("Inboxes.List", func(t *testing.T) {
		_, err := client.Inboxes.List(ctx, &anima.InboxListParams{ListParams: anima.ListParams{Limit: 1}})
		probe(t, err)
	})

	t.Run("Webhooks.List", func(t *testing.T) {
		_, err := client.Webhooks.List(ctx, nil)
		probe(t, err)
	})
}

func TestLiveVoiceSurface(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()

	t.Run("Voices.List", func(t *testing.T) {
		_, err := client.Voices.List(ctx, anima.ListVoicesParams{})
		probe(t, err)
	})

	t.Run("Calls.List", func(t *testing.T) {
		_, err := client.Calls.List(ctx, anima.ListCallsParams{Limit: 1})
		probe(t, err)
	})
}

func TestLiveVaultSurface(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()

	t.Run("Vault.ListIdentities", func(t *testing.T) {
		_, err := client.Vault.ListIdentities(ctx, &anima.ListVaultIdentitiesParams{
			ListParams: anima.ListParams{Limit: 1},
		})
		probe(t, err)
	})

	t.Run("Vault.Audit", func(t *testing.T) {
		_, err := client.Vault.Audit(ctx, &anima.VaultAuditParams{ListParams: anima.ListParams{Limit: 1}})
		probe(t, err)
	})

	// Added 2026-08; nothing had exercised this route from any SDK.
	t.Run("Vault.ListCredentialRequests", func(t *testing.T) {
		_, err := client.Vault.ListCredentialRequests(ctx, &anima.ListCredentialRequestsParams{
			ListParams: anima.ListParams{Limit: 1},
		})
		probe(t, err)
	})
}

// The surface that was most wrong: every anomaly and audit path was nested one
// level too deep, and every compliance enum was lowercase.
func TestLiveOrgScopedSurface(t *testing.T) {
	client := liveClient(t)
	orgID := liveOrgID(t)
	ctx := context.Background()

	t.Run("Audit.List", func(t *testing.T) {
		_, err := client.Audit.List(ctx, orgID, &anima.AuditLogListParams{ListParams: anima.ListParams{Limit: 1}})
		probe(t, err)
	})

	t.Run("Anomaly.ListAlerts", func(t *testing.T) {
		_, err := client.Anomaly.ListAlerts(ctx, orgID, &anima.AnomalyAlertListParams{
			ListParams: anima.ListParams{Limit: 1},
		})
		probe(t, err)
	})

	t.Run("Anomaly.ListRules", func(t *testing.T) {
		_, err := client.Anomaly.ListRules(ctx, orgID, nil)
		probe(t, err)
	})

	t.Run("Security.GetScannerStatus", func(t *testing.T) {
		_, err := client.Security.GetScannerStatus(ctx, orgID)
		probe(t, err)
	})

	t.Run("Compliance.ListControls", func(t *testing.T) {
		_, err := client.Compliance.ListControls(ctx, orgID, nil)
		probe(t, err)
	})

	t.Run("Compliance.ListTemplates", func(t *testing.T) {
		_, err := client.Compliance.ListTemplates(ctx, orgID)
		probe(t, err)
	})
}
