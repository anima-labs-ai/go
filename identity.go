package anima

import (
	"context"
	"fmt"
	"net/http"
)

// DidDocument represents a DID (Decentralized Identifier) document for an agent.
type DidDocument struct {
	ID                   string                   `json:"id"`
	Context              []string                 `json:"@context"`
	Controller           string                   `json:"controller,omitempty"`
	VerificationMethod   []DidVerificationMethod   `json:"verificationMethod,omitempty"`
	Authentication       []string                 `json:"authentication,omitempty"`
	AssertionMethod      []string                 `json:"assertionMethod,omitempty"`
	KeyAgreement         []string                 `json:"keyAgreement,omitempty"`
	Service              []DidService             `json:"service,omitempty"`
	Created              string                   `json:"created,omitempty"`
	Updated              string                   `json:"updated,omitempty"`
}

// DidVerificationMethod represents a verification method in a DID document.
type DidVerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`
	PublicKeyJWK       any    `json:"publicKeyJwk,omitempty"`
}

// DidService represents a service endpoint in a DID document.
type DidService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

// DidRotateOutput contains the result of a DID key rotation.
type DidRotateOutput struct {
	DID     string      `json:"did"`
	Document DidDocument `json:"document"`
	Rotated  bool        `json:"rotated"`
}

// VCType is a type of verifiable credential Anima issues. (Distinct from
// CredentialType, which is a vault credential kind.)
type VCType string

const (
	// Platform-issued automatically on verification events; requesting these
	// via IssueCredential returns a 403.
	VCTypeEmailVerified  VCType = "AnimaEmailVerified"
	VCTypePhoneVerified  VCType = "AnimaPhoneVerified"
	VCTypeKYBCompleted   VCType = "AnimaKYBCompleted"
	VCTypePaymentCapable VCType = "AnimaPaymentCapable"
	VCTypeOwnerBound     VCType = "AnimaOwnerBound"

	// Org-attestation types issuable via IssueCredential (master key).
	VCTypeAddressVerified VCType = "AnimaAddressVerified"
	VCTypeTrustScore      VCType = "AnimaTrustScore"
)

// VerifiableCredential is a Verifiable Credential record as stored by the
// platform. The signed W3C credential itself is the JWTVC string; the other
// fields are the platform's record of it (issuance, expiry, revocation).
type VerifiableCredential struct {
	ID         string `json:"id"`
	AgentID    string `json:"agentId"`
	OrgID      string `json:"orgId"`
	Type       string `json:"type"`
	JWTVC      string `json:"jwtVc"`
	IssuerDID  string `json:"issuerDid"`
	SubjectDID string `json:"subjectDid"`
	IssuedAt   string `json:"issuedAt"`
	// ExpiresAt is nil for non-expiring credentials.
	ExpiresAt *string `json:"expiresAt"`
	Revoked   bool    `json:"revoked"`
	RevokedAt *string `json:"revokedAt"`
	// RevocationIndex is the credential's bit position in the issuer's
	// StatusList, nil if none was assigned.
	RevocationIndex *int                   `json:"revocationIndex"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       string                 `json:"createdAt"`
	UpdatedAt       string                 `json:"updatedAt"`
}

// IssueCredentialParams contains the parameters for issuing a verifiable
// credential to an agent. Master-key only.
//
// Only the org-attestation types (VCTypeAddressVerified, VCTypeTrustScore)
// can be issued here; the platform-reserved types are auto-issued on real
// verification events (email OTP, phone provisioning, Stripe checkout) and
// return a 403 from this endpoint.
type IssueCredentialParams struct {
	// Type is the credential type to issue.
	Type VCType `json:"type"`
	// Claims are additional claims for the credential subject. The subject
	// id is always the agent's DID.
	Claims map[string]interface{} `json:"claims,omitempty"`
	// ExpiresInSeconds is the optional credential lifetime; omit (0) for a
	// non-expiring credential.
	ExpiresInSeconds int `json:"expiresInSeconds,omitempty"`
}

// VerifyCredentialOutput contains the result of a credential verification.
type VerifyCredentialOutput struct {
	Valid   bool     `json:"valid"`
	Checks []string `json:"checks,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

// AgentCardOutput represents an agent's public card (machine-readable profile).
type AgentCardOutput struct {
	AgentID      string   `json:"agentId"`
	DID          string   `json:"did"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Endpoints    any      `json:"endpoints,omitempty"`
	CreatedAt    string   `json:"createdAt,omitempty"`
	UpdatedAt    string   `json:"updatedAt,omitempty"`
}

// IdentityService provides methods for managing agent decentralized identity.
type IdentityService struct {
	client *httpClient
}

// newIdentityService creates a new IdentityService.
func newIdentityService(c *httpClient) *IdentityService {
	return &IdentityService{client: c}
}

// GetDID retrieves the DID document for an agent.
func (s *IdentityService) GetDID(ctx context.Context, agentID string) (*DidDocument, error) {
	doc, err := Do[DidDocument](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s/did", agentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// ResolveDID resolves a DID to its DID document.
func (s *IdentityService) ResolveDID(ctx context.Context, did string) (*DidDocument, error) {
	doc, err := Do[DidDocument](ctx, s.client, http.MethodGet, fmt.Sprintf("/identity/did/%s", did), nil, nil)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// RotateKeys rotates the cryptographic keys for an agent's DID.
func (s *IdentityService) RotateKeys(ctx context.Context, agentID string) (*DidRotateOutput, error) {
	result, err := Do[DidRotateOutput](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/did/rotate", agentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListCredentials lists all verifiable credential records for an agent,
// newest first. The API returns a bare JSON array (not an items envelope).
func (s *IdentityService) ListCredentials(ctx context.Context, agentID string) ([]VerifiableCredential, error) {
	list, err := Do[[]VerifiableCredential](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s/credentials", agentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// IssueCredential issues a verifiable credential to an agent (master key).
// See IssueCredentialParams for which types are issuable here.
func (s *IdentityService) IssueCredential(ctx context.Context, agentID string, params IssueCredentialParams) (*VerifiableCredential, error) {
	body := struct {
		AgentID string `json:"agentId"`
		IssueCredentialParams
	}{
		AgentID:               agentID,
		IssueCredentialParams: params,
	}
	vc, err := Do[VerifiableCredential](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/credentials", agentID), body, nil)
	if err != nil {
		return nil, err
	}
	return &vc, nil
}

// RevokeCredential revokes a previously issued verifiable credential and
// returns the updated (revoked) credential record.
func (s *IdentityService) RevokeCredential(ctx context.Context, agentID, vcID string) (*VerifiableCredential, error) {
	body := struct {
		AgentID string `json:"agentId"`
		VCID    string `json:"vcId"`
	}{AgentID: agentID, VCID: vcID}
	vc, err := Do[VerifiableCredential](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/credentials/%s/revoke", agentID, vcID), body, nil)
	if err != nil {
		return nil, err
	}
	return &vc, nil
}

// VerifyCredential verifies a JWT-encoded verifiable credential.
func (s *IdentityService) VerifyCredential(ctx context.Context, jwtVC string) (*VerifyCredentialOutput, error) {
	body := struct {
		JWTVC string `json:"jwtVc"`
	}{JWTVC: jwtVC}
	result, err := Do[VerifyCredentialOutput](ctx, s.client, http.MethodPost, "/identity/verify", body, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAgentCard retrieves the public agent card for an agent.
func (s *IdentityService) GetAgentCard(ctx context.Context, agentID string) (*AgentCardOutput, error) {
	card, err := Do[AgentCardOutput](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s/card", agentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &card, nil
}
