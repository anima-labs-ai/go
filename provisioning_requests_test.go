package anima

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// Approve had no test at all, which is how it stayed green through a signature
// change. These cover the two things the permission queue actually needs from
// this SDK: sending a grant, and being able to read a GENERIC row back.

func TestProvisioningRequestsService_ApproveSendsGrant(t *testing.T) {
	var body map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/provisioning-requests/req_1/approve", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"req_1","agentId":"a1","agentName":"A","resource":"GENERIC","reason":"x","status":"APPROVED","options":null,"permission":null,"expiresAt":"2026-01-01T00:00:00Z","decidedAt":null,"decidedNote":null,"provisionedId":"perm_1","createdAt":"2026-01-01T00:00:00Z"}`))
	})
	client, ts := newTestClient(mux)
	defer ts.Close()

	got, err := client.ProvisioningRequests.Approve(context.Background(), "req_1",
		DecideProvisioningRequestParams{Grant: PermissionGrantAlways})
	if err != nil {
		t.Fatalf("Approve: %v", err)
	}
	// The grant reaching the wire is the whole point: approving a permission
	// request without one is a 422, so an SDK that dropped it could not approve
	// a permission request at all.
	if body["grant"] != "always" {
		t.Errorf("expected grant=always on the wire, got %v", body["grant"])
	}
	// RequestID is filled from the path argument, so callers cannot set one
	// that disagrees with the URL.
	if body["requestId"] != "req_1" {
		t.Errorf("expected requestId=req_1, got %v", body["requestId"])
	}
	if got.ProvisionedID == nil || *got.ProvisionedID != "perm_1" {
		t.Errorf("expected provisionedId=perm_1, got %v", got.ProvisionedID)
	}
}

func TestProvisioningRequestsService_ApproveOmitsGrantWhenUnset(t *testing.T) {
	var body map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/provisioning-requests/req_2/approve", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"req_2","agentId":"a1","agentName":"A","resource":"VAULT","reason":"x","status":"APPROVED","options":null,"permission":null,"expiresAt":"2026-01-01T00:00:00Z","decidedAt":null,"decidedNote":null,"provisionedId":"vault_1","createdAt":"2026-01-01T00:00:00Z"}`))
	})
	client, ts := newTestClient(mux)
	defer ts.Close()

	if _, err := client.ProvisioningRequests.Approve(context.Background(), "req_2",
		DecideProvisioningRequestParams{}); err != nil {
		t.Fatalf("Approve: %v", err)
	}
	// The API rejects a grant sent with a resource request, so an empty grant
	// must be absent rather than an empty string.
	if _, present := body["grant"]; present {
		t.Errorf("expected no grant key for a resource request, got %v", body["grant"])
	}
}

func TestProvisioningRequestsService_ParsesGenericRow(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/provisioning-requests/req_3", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"req_3","agentId":"a1","agentName":"A","resource":"GENERIC","reason":"needs approval","status":"PENDING","options":null,"permission":{"procedurePath":"agent.delete","readOnly":false,"argumentPreview":{"agentId":"agent_abc"}},"expiresAt":"2026-01-01T00:00:00Z","decidedAt":null,"decidedNote":null,"provisionedId":null,"createdAt":"2026-01-01T00:00:00Z"}`))
	})
	client, ts := newTestClient(mux)
	defer ts.Close()

	got, err := client.ProvisioningRequests.Get(context.Background(), "req_3")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Resource != ProvisionableResourceGeneric {
		t.Errorf("expected GENERIC, got %v", got.Resource)
	}
	if got.Permission == nil {
		t.Fatal("expected permission detail on a GENERIC row")
	}
	if got.Permission.ProcedurePath != "agent.delete" {
		t.Errorf("expected agent.delete, got %s", got.Permission.ProcedurePath)
	}
	if got.Permission.ReadOnly {
		t.Error("agent.delete is not read-only; Bypass must not be offered for it")
	}
	if got.Permission.ArgumentPreview["agentId"] != "agent_abc" {
		t.Errorf("expected redacted preview, got %v", got.Permission.ArgumentPreview)
	}
}
