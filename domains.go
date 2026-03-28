package anima

import (
	"context"
	"fmt"
	"net/http"
)

// DomainStatus represents the verification status of a domain.
type DomainStatus string

const (
	DomainStatusNotStarted DomainStatus = "NOT_STARTED"
	DomainStatusPending    DomainStatus = "PENDING"
	DomainStatusVerifying  DomainStatus = "VERIFYING"
	DomainStatusVerified   DomainStatus = "VERIFIED"
	DomainStatusInvalid    DomainStatus = "INVALID"
	DomainStatusFailed     DomainStatus = "FAILED"
)

// VerificationMethod is the method used to verify a domain.
type VerificationMethod string

const (
	VerificationMethodDNSTXT   VerificationMethod = "DNS_TXT"
	VerificationMethodDNSCNAME VerificationMethod = "DNS_CNAME"
)

// DomainRecordStatus represents the status of a DNS record.
type DomainRecordStatus string

const (
	DomainRecordStatusMissing DomainRecordStatus = "MISSING"
	DomainRecordStatusInvalid DomainRecordStatus = "INVALID"
	DomainRecordStatusValid   DomainRecordStatus = "VALID"
)

// DomainStatusRecord represents a DNS record and its verification status.
type DomainStatusRecord struct {
	Type     string             `json:"type"`
	Name     string             `json:"name"`
	Value    string             `json:"value"`
	Priority *int               `json:"priority,omitempty"`
	Status   DomainRecordStatus `json:"status"`
}

// Domain represents an email domain in the Anima platform.
type Domain struct {
	ID                       string               `json:"id"`
	Domain                   string               `json:"domain"`
	Status                   DomainStatus          `json:"status"`
	Verified                 bool                  `json:"verified"`
	VerificationCooldownUntil *string              `json:"verificationCooldownUntil"`
	VerificationToken        string               `json:"verificationToken"`
	VerificationMethod       VerificationMethod   `json:"verificationMethod"`
	DKIMSelector             *string              `json:"dkimSelector"`
	DKIMPublicKey            *string              `json:"dkimPublicKey"`
	SPFConfigured            bool                 `json:"spfConfigured"`
	DMARCConfigured          bool                 `json:"dmarcConfigured"`
	MXConfigured             bool                 `json:"mxConfigured"`
	FeedbackEnabled          bool                 `json:"feedbackEnabled"`
	Records                  []DomainStatusRecord `json:"records"`
	CreatedAt                string               `json:"createdAt"`
}

// AddDomainParams contains the parameters for adding a domain.
type AddDomainParams struct {
	Domain string `json:"domain"`
}

// UpdateDomainParams contains the parameters for updating a domain.
type UpdateDomainParams struct {
	FeedbackEnabled *bool `json:"feedbackEnabled,omitempty"`
}

// DomainDNSRecord represents a single DNS record entry.
type DomainDNSRecord struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// DomainDNSRecordWithPriority is a DNS record with a priority field.
type DomainDNSRecordWithPriority struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority int    `json:"priority"`
}

// DomainMailFrom contains the mail-from DNS configuration.
type DomainMailFrom struct {
	Name string                      `json:"name"`
	MX   DomainDNSRecordWithPriority `json:"mx"`
	SPF  string                      `json:"spf"`
}

// DomainDNSRecords contains all the DNS records needed for a domain.
type DomainDNSRecords struct {
	TXT      DomainDNSRecord             `json:"txt"`
	MailFrom DomainMailFrom              `json:"mailFrom"`
	DKIM     []DomainDNSRecord           `json:"dkim"`
	MX       DomainDNSRecordWithPriority `json:"mx"`
	SPF      string                      `json:"spf"`
	DMARC    string                      `json:"dmarc"`
}

// DeliverabilityStats contains deliverability statistics for a domain.
type DeliverabilityStats struct {
	Domain        string  `json:"domain"`
	Sent          int     `json:"sent"`
	Delivered     int     `json:"delivered"`
	Bounced       int     `json:"bounced"`
	Complained    int     `json:"complained"`
	BounceRate    float64 `json:"bounceRate"`
	ComplaintRate float64 `json:"complaintRate"`
	IsHealthy     bool    `json:"isHealthy"`
}

// DomainZoneFile contains a zone file export.
type DomainZoneFile struct {
	ZoneFile string `json:"zoneFile"`
}

// DomainList wraps a list of domains (non-paginated).
type DomainList struct {
	Items []Domain `json:"items"`
}

// DomainsService provides methods for managing email domains.
type DomainsService struct {
	client *httpClient
}

// newDomainsService creates a new DomainsService.
func newDomainsService(c *httpClient) *DomainsService {
	return &DomainsService{client: c}
}

// Add registers a new domain.
func (s *DomainsService) Add(ctx context.Context, params AddDomainParams) (*Domain, error) {
	d, err := Do[Domain](ctx, s.client, http.MethodPost, "/domains", params, nil)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Get retrieves a domain by ID.
func (s *DomainsService) Get(ctx context.Context, id string) (*Domain, error) {
	d, err := Do[Domain](ctx, s.client, http.MethodGet, fmt.Sprintf("/domains/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// List returns all domains.
func (s *DomainsService) List(ctx context.Context) (*DomainList, error) {
	list, err := Do[DomainList](ctx, s.client, http.MethodGet, "/domains", nil, nil)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Delete removes a domain.
func (s *DomainsService) Delete(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/domains/%s", id), nil, nil)
	return err
}

// Update updates a domain's settings.
func (s *DomainsService) Update(ctx context.Context, id string, params UpdateDomainParams) (*Domain, error) {
	d, err := Do[Domain](ctx, s.client, http.MethodPatch, fmt.Sprintf("/domains/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Verify triggers domain verification.
func (s *DomainsService) Verify(ctx context.Context, id string) (*Domain, error) {
	d, err := Do[Domain](ctx, s.client, http.MethodPost, fmt.Sprintf("/domains/%s/verify", id), map[string]string{"id": id, "domainId": id}, nil)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// DNSRecords retrieves the required DNS records for a domain.
func (s *DomainsService) DNSRecords(ctx context.Context, id string) (*DomainDNSRecords, error) {
	records, err := Do[DomainDNSRecords](ctx, s.client, http.MethodGet, fmt.Sprintf("/domains/%s/dns-records", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &records, nil
}

// Deliverability retrieves deliverability stats for a domain.
func (s *DomainsService) Deliverability(ctx context.Context, id string) (*DeliverabilityStats, error) {
	stats, err := Do[DeliverabilityStats](ctx, s.client, http.MethodGet, fmt.Sprintf("/domains/%s/deliverability", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// ZoneFile exports the zone file for a domain.
func (s *DomainsService) ZoneFile(ctx context.Context, id string) (*DomainZoneFile, error) {
	zf, err := Do[DomainZoneFile](ctx, s.client, http.MethodGet, fmt.Sprintf("/domains/%s/zone-file", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &zf, nil
}
