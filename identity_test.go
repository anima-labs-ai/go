package anima

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

// TestIdentityService_ListCredentials asserts the endpoint's actual wire
// shape: a BARE JSON array of credential records (no items envelope). The
// SDK previously expected {"items": [...]} and every real call failed at
// decode — this is the regression test for that fix.
func TestIdentityService_ListCredentials(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agents/agent_123/credentials", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"id":              "vc_1",
				"agentId":         "agent_123",
				"orgId":           "org_1",
				"type":            "AnimaEmailVerified",
				"jwtVc":           "eyJhbGciOiJFZERTQSJ9.e30.sig",
				"issuerDid":       "did:web:agents.useanima.sh:anima:platform",
				"subjectDid":      "did:web:agents.useanima.sh:org1:agent123",
				"issuedAt":        "2026-07-17T10:00:00Z",
				"expiresAt":       nil,
				"revoked":         false,
				"revokedAt":       nil,
				"revocationIndex": 7,
				"metadata":        map[string]interface{}{"source": "platform-auto"},
				"createdAt":       "2026-07-17T10:00:00Z",
				"updatedAt":       "2026-07-17T10:00:00Z",
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	creds, err := client.Identity.ListCredentials(context.Background(), "agent_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}
	vc := creds[0]
	if vc.Type != "AnimaEmailVerified" {
		t.Errorf("expected type 'AnimaEmailVerified', got %q", vc.Type)
	}
	// The signed credential itself is the JWT-VC string — losing it makes
	// the record useless for presentation/verification.
	if vc.JWTVC == "" {
		t.Error("expected jwtVc to round-trip, got empty string")
	}
	if vc.ExpiresAt != nil {
		t.Errorf("expected nil ExpiresAt on a non-expiring credential, got %v", *vc.ExpiresAt)
	}
	if vc.RevocationIndex == nil || *vc.RevocationIndex != 7 {
		t.Errorf("expected RevocationIndex 7, got %v", vc.RevocationIndex)
	}
	if vc.Revoked {
		t.Error("expected Revoked false")
	}
}

func TestIdentityService_IssueCredential(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agents/agent_123/credentials", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body := decodeRawBody(t, r)
		if body["agentId"] != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %v", body["agentId"])
		}
		if body["type"] != "AnimaAddressVerified" {
			t.Errorf("expected type 'AnimaAddressVerified', got %v", body["type"])
		}
		claims, ok := body["claims"].(map[string]interface{})
		if !ok || claims["country"] != "BG" {
			t.Errorf("expected claims.country 'BG', got %v", body["claims"])
		}
		if body["expiresInSeconds"] != float64(3600) {
			t.Errorf("expected expiresInSeconds 3600, got %v", body["expiresInSeconds"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         "vc_2",
			"agentId":    "agent_123",
			"orgId":      "org_1",
			"type":       "AnimaAddressVerified",
			"jwtVc":      "eyJhbGciOiJFZERTQSJ9.e30.sig2",
			"issuerDid":  "did:web:agents.useanima.sh:anima:platform",
			"subjectDid": "did:web:agents.useanima.sh:org1:agent123",
			"issuedAt":   "2026-07-17T11:00:00Z",
			"expiresAt":  "2026-07-17T12:00:00Z",
			"revoked":    false,
			"metadata":   map[string]interface{}{"source": "api"},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	vc, err := client.Identity.IssueCredential(context.Background(), "agent_123", IssueCredentialParams{
		Type:             VCTypeAddressVerified,
		Claims:           map[string]interface{}{"country": "BG"},
		ExpiresInSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vc.ID != "vc_2" {
		t.Errorf("expected ID 'vc_2', got %q", vc.ID)
	}
	if vc.ExpiresAt == nil || *vc.ExpiresAt != "2026-07-17T12:00:00Z" {
		t.Errorf("expected ExpiresAt '2026-07-17T12:00:00Z', got %v", vc.ExpiresAt)
	}
}

// TestIdentityService_IssueCredential_PlatformReserved asserts that the
// platform-reserved types (auto-issued on real verification events) are
// rejected loudly — issuing AnimaEmailVerified by hand would forge the
// trust level an agent card derives from it.
func TestIdentityService_IssueCredential_PlatformReserved(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agents/agent_123/credentials", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "AnimaEmailVerified is platform-reserved and auto-issued on verification events",
				"code":    "FORBIDDEN",
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	_, err := client.Identity.IssueCredential(context.Background(), "agent_123", IssueCredentialParams{
		Type: VCTypeEmailVerified,
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, ErrAuthentication) {
		t.Errorf("expected ErrAuthentication (403), got %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", apiErr.Status)
	}
}

func TestIdentityService_RevokeCredential(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agents/agent_123/credentials/vc_2/revoke", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["agentId"] != "agent_123" {
			t.Errorf("expected agentId 'agent_123', got %v", body["agentId"])
		}
		if body["vcId"] != "vc_2" {
			t.Errorf("expected vcId 'vc_2', got %v", body["vcId"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "vc_2",
			"agentId":   "agent_123",
			"orgId":     "org_1",
			"type":      "AnimaAddressVerified",
			"jwtVc":     "eyJhbGciOiJFZERTQSJ9.e30.sig2",
			"revoked":   true,
			"revokedAt": "2026-07-17T12:30:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	vc, err := client.Identity.RevokeCredential(context.Background(), "agent_123", "vc_2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Revocation must be reflected in the returned record — the caller
	// relies on it to confirm the credential is no longer presentable.
	if !vc.Revoked {
		t.Error("expected Revoked true after revocation")
	}
	if vc.RevokedAt == nil || *vc.RevokedAt != "2026-07-17T12:30:00Z" {
		t.Errorf("expected RevokedAt '2026-07-17T12:30:00Z', got %v", vc.RevokedAt)
	}
}

// VerifyCredential decodes the contract's {valid, credential, errors}.
//
// The fixture below is a real JwtVcPayload (packages/agent-identity/src/vc.ts),
// which is what the handler returns — NOT a flat W3C document. The SDK first
// declared a "checks" field the API has never sent, then replaced it with a
// typed struct of id/issuer/subject/proof: none of those keys exist at the top
// level of a JWT-VC, so every field decoded to its zero value while Valid was
// true and Credential was non-nil. A caller reading credentialSubject got an
// empty map and treated a verified agent as unverified.
func TestIdentityService_VerifyCredential(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/identity/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body struct {
			JWTVC string `json:"jwtVc"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.JWTVC != "eyJhbGciOiJFZERTQSJ9.e30.sig" {
			t.Errorf("unexpected jwtVc %q", body.JWTVC)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": true,
			"credential": map[string]interface{}{
				"iss": "did:web:agents.useanima.sh:anima:platform",
				"sub": "did:web:agents.useanima.sh:org1:agent123",
				"jti": "urn:uuid:vc-1",
				"iat": 1753000000,
				"vc": map[string]interface{}{
					"@context":          []string{"https://www.w3.org/2018/credentials/v1"},
					"type":              []string{"VerifiableCredential", "AnimaEmailVerified"},
					"issuer":            "did:web:agents.useanima.sh:anima:platform",
					"issuanceDate":      "2026-07-17T10:00:00Z",
					"credentialSubject": map[string]interface{}{"id": "did:web:x", "email": "a@example.com"},
				},
			},
			"errors": []string{},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Identity.VerifyCredential(context.Background(), "eyJhbGciOiJFZERTQSJ9.e30.sig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid=true")
	}
	if result.Credential == nil {
		t.Fatal("expected the decoded payload, got nil")
	}
	// The claims live under "vc", not at the top level. A struct modelling a
	// flat W3C document reaches for result.Credential.Issuer and finds nothing.
	if result.Credential["iss"] != "did:web:agents.useanima.sh:anima:platform" {
		t.Errorf("unexpected iss %v", result.Credential["iss"])
	}
	vc, ok := result.Credential["vc"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected a nested vc object, got %T", result.Credential["vc"])
	}
	subject, ok := vc["credentialSubject"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected credentialSubject, got %T", vc["credentialSubject"])
	}
	if subject["email"] != "a@example.com" {
		t.Errorf("credentialSubject did not decode: %v", subject)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %v", result.Errors)
	}
}

// A revoked credential still decodes — Credential is non-nil with valid=false.
// Branching on Credential != nil instead of Valid would accept it.
func TestIdentityService_VerifyCredential_RevokedStillDecodes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/identity/verify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"credential": map[string]interface{}{
				"iss": "did:web:agents.useanima.sh:anima:platform",
				"vc":  map[string]interface{}{"type": []string{"VerifiableCredential"}},
			},
			"errors": []string{"revoked"},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	result, err := client.Identity.VerifyCredential(context.Background(), "eyJ.revoked.sig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected valid=false")
	}
	if result.Credential == nil {
		t.Error("a revoked credential still decodes — Credential must not be nil")
	}
	if len(result.Errors) != 1 || result.Errors[0] != "revoked" {
		t.Errorf("expected [revoked], got %v", result.Errors)
	}
}
