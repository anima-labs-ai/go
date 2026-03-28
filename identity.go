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

// VerifiableCredential represents a W3C Verifiable Credential.
type VerifiableCredential struct {
	ID                string   `json:"id"`
	Type              []string `json:"type"`
	Issuer            string   `json:"issuer"`
	IssuanceDate      string   `json:"issuanceDate"`
	ExpirationDate    string   `json:"expirationDate,omitempty"`
	CredentialSubject any      `json:"credentialSubject"`
	Proof             any      `json:"proof,omitempty"`
	JWT               string   `json:"jwt,omitempty"`
}

// VerifiableCredentialList wraps a list of verifiable credentials.
type VerifiableCredentialList struct {
	Items []VerifiableCredential `json:"items"`
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

// ListCredentials lists all verifiable credentials for an agent.
func (s *IdentityService) ListCredentials(ctx context.Context, agentID string) (*VerifiableCredentialList, error) {
	list, err := Do[VerifiableCredentialList](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s/credentials", agentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &list, nil
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
