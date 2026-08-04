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
//	ANIMA_LIVE_AGENT_ID=...    # required for agent-scoped vault probes
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
	"strings"
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

// Several vault routes are agent-scoped and REJECT a master key that does not
// name an agent ("agentId is required when using a master key"). Without one
// they skip rather than fail for a reason unrelated to conformance.
func liveAgentID(t *testing.T) string {
	t.Helper()
	agentID := os.Getenv("ANIMA_LIVE_AGENT_ID")
	if agentID == "" {
		t.Skip("set ANIMA_LIVE_AGENT_ID for agent-scoped vault probes")
	}
	return agentID
}

func liveOrgID(t *testing.T) string {
	t.Helper()
	orgID := os.Getenv("ANIMA_LIVE_ORG_ID")
	if orgID == "" {
		t.Skip("set ANIMA_LIVE_ORG_ID for org-scoped probes")
	}
	return orgID
}

// reached counts probes that got a real 2xx back.
//
// A 401 is a pass for each probe on its own — it proves the route exists — but
// that does not compose: a key scoped for nothing is rejected everywhere and
// turns this whole file green having checked no path, no query param and no
// response shape. TestLiveRunReachedTheAPI fails that run.
//
// No lock: every probe here runs sequentially (no subtest calls t.Parallel),
// so -race has nothing to complain about.
var reached int

// probe applies the verdict table above to one read-only call.
func probe(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		reached++
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
	case apiErr.Status >= 500 && strings.Contains(apiErr.Error(), "bw-serve: Unable to connect"):
		// A vault route that reaches `bw serve` and finds nothing listening has
		// already proved everything conformance cares about: the route exists,
		// auth passed, and the API accepted the request the SDK built. The
		// missing piece is the deployment's storage backend. Narrow on purpose
		// — every other 5xx still fails below.
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

	t.Run("Vault.Status", func(t *testing.T) {
		agentID := liveAgentID(t)
		_, err := client.Vault.Status(ctx, agentID)
		probe(t, err)
	})
}

// An agent asking its owner for a vault or a phone number. The status and
// resource filters are server-side enums: sending the wrong casing earns a 400,
// and no fixture ever would.
func TestLiveProvisioningRequests(t *testing.T) {
	client := liveClient(t)
	ctx := context.Background()

	t.Run("ProvisioningRequests.List", func(t *testing.T) {
		page, err := client.ProvisioningRequests.List(ctx, &anima.ListProvisioningRequestsParams{
			ListParams: anima.ListParams{Limit: 1},
		})
		probe(t, err)
		if err == nil && page.Items == nil {
			t.Error("items was null — the envelope is not what the SDK declares")
		}
	})

	t.Run("ProvisioningRequests.List_StatusFilter", func(t *testing.T) {
		_, err := client.ProvisioningRequests.List(ctx, &anima.ListProvisioningRequestsParams{
			ListParams: anima.ListParams{Limit: 1},
			Status:     anima.ProvisioningRequestPending,
		})
		probe(t, err)
	})

	t.Run("ProvisioningRequests.List_ResourceFilter", func(t *testing.T) {
		_, err := client.ProvisioningRequests.List(ctx, &anima.ListProvisioningRequestsParams{
			ListParams: anima.ListParams{Limit: 1},
			Resource:   anima.ProvisionableResourceVault,
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
		page, err := client.Audit.List(ctx, orgID, &anima.AuditLogListParams{ListParams: anima.ListParams{Limit: 1}})
		probe(t, err)
		// Audit answers the FLAT envelope {items, nextCursor}. Pagination is a
		// value type, so before Page.UnmarshalJSON handled both shapes this
		// decoded to the zero value: HasMore false, and ListAutoPaging stopped
		// after one page returning a nil error. Assert the cursor survived.
		if err == nil && page != nil && len(page.Items) > 0 && page.Pagination.NextCursor == nil {
			t.Error("flat envelope lost its cursor — Page.UnmarshalJSON is not normalizing")
		}
	})

	// UPPERCASE in the contract; these constants were lowercase until
	// 2026-08-04. Sent as FILTERS, so a wrong casing is a 400 rather than an
	// empty list — Go validates nothing on the way in, so a response assertion
	// would pass on an org with no audit rows.
	t.Run("Audit.List_ActorTypeEnum", func(t *testing.T) {
		for _, at := range []anima.AuditActorType{
			anima.AuditActorAPIKey, anima.AuditActorUser,
			anima.AuditActorSystem, anima.AuditActorAgent,
		} {
			_, err := client.Audit.List(ctx, orgID, &anima.AuditLogListParams{
				ListParams: anima.ListParams{Limit: 1}, ActorType: at,
			})
			probe(t, err)
		}
	})

	t.Run("Audit.List_ResultEnum", func(t *testing.T) {
		for _, r := range []anima.AuditResult{
			anima.AuditResultSuccess, anima.AuditResultFailure, anima.AuditResultDenied,
		} {
			_, err := client.Audit.List(ctx, orgID, &anima.AuditLogListParams{
				ListParams: anima.ListParams{Limit: 1}, Result: r,
			})
			probe(t, err)
		}
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

// TestLiveDeclaredEnumsAreAccepted sends every enum value this SDK declares and
// confirms the API accepts it.
//
// This is the probe that would have caught the compliance bug outright: every
// enum shipped lowercase against a contract validating SCREAMING_SNAKE, so each
// of these would have come back 400 — while the mocked tests passed.
//
// The slices name the exported constants rather than string literals, so
// renaming or deleting one is a compile error here. Go cannot enumerate a
// string-constant type, so a value ADDED to the set still has to be added
// below by hand — the one gap the python probe does not have, since it
// iterates its Enum directly.
func TestLiveDeclaredEnumsAreAccepted(t *testing.T) {
	client := liveClient(t)
	orgID := liveOrgID(t)
	ctx := context.Background()

	frameworks := []anima.ComplianceFramework{
		anima.ComplianceFrameworkSOC2,
		anima.ComplianceFrameworkGDPR,
		anima.ComplianceFrameworkPCI,
	}
	for _, framework := range frameworks {
		t.Run("framework "+string(framework), func(t *testing.T) {
			_, err := client.Compliance.ListControls(ctx, orgID, &anima.ComplianceControlListParams{
				ListParams: anima.ListParams{Limit: 1},
				Framework:  framework,
			})
			probe(t, err)
		})
	}

	severities := []anima.SecuritySeverity{
		anima.SecuritySeverityLow,
		anima.SecuritySeverityMedium,
		anima.SecuritySeverityHigh,
		anima.SecuritySeverityCritical,
	}
	for _, severity := range severities {
		t.Run("severity "+string(severity), func(t *testing.T) {
			_, err := client.Security.ListEvents(ctx, anima.SecurityEventsListParams{
				ListParams: anima.ListParams{Limit: 1},
				OrgID:      orgID,
				Severity:   severity,
			})
			probe(t, err)
		})
	}
}

// TestLiveRunReachedTheAPI fails a run in which no probe ever got a 2xx.
//
// Every verdict in probe() is sound on its own, but "401 is a pass" does not
// compose: a key with no scopes is rejected everywhere, each probe passes
// because the route demonstrably exists, and this file goes green having
// verified nothing at all — the same hollow tick the mocks were giving us,
// which is the entire reason the file exists.
//
// Deliberately the last test in the file: go test runs a file's tests in
// source order, so every probe above has already run and settled `reached`.
func TestLiveRunReachedTheAPI(t *testing.T) {
	liveClient(t) // skip along with every other probe when no key is set
	if reached == 0 {
		t.Error("no probe reached the API: every call was rejected (401/403), " +
			"rate limited, or skipped, so this run verified no path, no query " +
			"param and no response shape. Check that ANIMA_LIVE_API_KEY is valid " +
			"and scoped — a green run in this state would prove nothing.")
	}
}
