package anima

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// wantTiers is the exact set of tiers the API can accept or return, mirroring
// TierSchema in packages/contracts/src/schemas/organization.ts and the Prisma
// Tier enum. Written out rather than derived from the constants — a check that
// reads the same declarations it is checking cannot fail when they are wrong.
//
// Note the limitation: Go offers no reflection over package constants, so this
// cannot detect a *newly added* phantom constant the way the Python and Node
// equivalents do. It pins the values and the count; the drift canary is what
// catches the contract moving underneath us.
var wantTiers = map[Tier]string{
	TierFree:       "FREE",
	TierStarter:    "STARTER",
	TierGrowth:     "GROWTH",
	TierEnterprise: "ENTERPRISE",
}

func TestTierConstants(t *testing.T) {
	// Referencing TierStarter at all is half the test: it did not exist, so a
	// caller wanting the Starter plan had to bypass the constants entirely and
	// write Tier("STARTER") by hand.
	for tier, want := range wantTiers {
		if string(tier) != want {
			t.Errorf("tier constant = %q, want %q", string(tier), want)
		}
	}

	if len(wantTiers) != 4 {
		t.Errorf("expected 4 tiers, got %d — the API's tier set changed", len(wantTiers))
	}
}

func TestOrganizationsService_GetDecodesEveryLiveTier(t *testing.T) {
	// The tier lands on Organization.Tier through the JSON decode path, so
	// exercise that rather than the constants in isolation.
	for tier := range wantTiers {
		t.Run(string(tier), func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/orgs/org_123", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{
					"id": "org_123",
					"name": "Acme",
					"slug": "acme",
					"clerkOrgId": null,
					"tier": %q,
					"masterKey": "sk_test_master",
					"settings": {},
					"createdAt": "2025-01-01T00:00:00Z",
					"updatedAt": "2025-01-01T00:00:00Z"
				}`, string(tier))
			})

			client, ts := newTestClient(mux)
			defer ts.Close()

			org, err := client.Organizations.Get(context.Background(), "org_123")
			if err != nil {
				t.Fatalf("Get returned error: %v", err)
			}
			if org.Tier != tier {
				t.Errorf("expected tier %q, got %q", tier, org.Tier)
			}
			if org.ID != "org_123" {
				t.Errorf("expected id org_123, got %q", org.ID)
			}
		})
	}
}

func TestCreateOrganizationParams_MarshalsStarterTier(t *testing.T) {
	// The outbound direction: TierStarter has to survive JSON marshalling as
	// "STARTER" or creating a Starter-plan org silently sends the wrong value.
	body, err := json.Marshal(CreateOrganizationParams{
		Name: "Acme",
		Slug: "acme",
		Tier: TierStarter,
	})
	if err != nil {
		t.Fatalf("marshal returned error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal returned error: %v", err)
	}
	if decoded["tier"] != "STARTER" {
		t.Errorf("expected tier STARTER on the wire, got %v", decoded["tier"])
	}
}
