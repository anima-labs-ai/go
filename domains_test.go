package anima

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestDomainsService_Add(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ak_test_key" {
			t.Errorf("expected Authorization: Bearer ak_test_key, got %s", got)
		}

		body := decodeRawBody(t, r)
		if body["domain"] != "mail.example.com" {
			t.Errorf("expected domain 'mail.example.com', got %v", body["domain"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                 "dom_1",
			"domain":             "mail.example.com",
			"status":             "PENDING",
			"verified":           false,
			"verificationToken":  "anima-verify-abc123",
			"verificationMethod": "DNS_TXT",
			"records": []map[string]interface{}{
				{"type": "TXT", "name": "_anima.mail.example.com", "value": "anima-verify-abc123", "status": "MISSING"},
			},
			"createdAt": "2026-07-17T10:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	d, err := client.Domains.Add(context.Background(), AddDomainParams{Domain: "mail.example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID != "dom_1" {
		t.Errorf("expected ID 'dom_1', got %q", d.ID)
	}
	// A freshly added domain must come back unverified with the DNS
	// challenge the caller has to publish — that token is the whole point
	// of the add step.
	if d.Verified {
		t.Error("expected a fresh domain to be unverified")
	}
	if d.Status != DomainStatusPending {
		t.Errorf("expected status PENDING, got %q", d.Status)
	}
	if d.VerificationToken != "anima-verify-abc123" {
		t.Errorf("expected the verification token to round-trip, got %q", d.VerificationToken)
	}
	if len(d.Records) != 1 || d.Records[0].Status != DomainRecordStatusMissing {
		t.Errorf("expected one MISSING record, got %+v", d.Records)
	}
}

func TestDomainsService_Add_Conflict(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Domain already exists",
				"code":    "CONFLICT",
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	_, err := client.Domains.Add(context.Background(), AddDomainParams{Domain: "mail.example.com"})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestDomainsService_Get(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains/dom_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       "dom_1",
			"domain":   "mail.example.com",
			"status":   "VERIFIED",
			"verified": true,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	d, err := client.Domains.Get(context.Background(), "dom_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Domain != "mail.example.com" {
		t.Errorf("expected domain 'mail.example.com', got %q", d.Domain)
	}
	if !d.Verified || d.Status != DomainStatusVerified {
		t.Errorf("expected a verified domain, got verified=%v status=%q", d.Verified, d.Status)
	}
}

func TestDomainsService_List(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		// The domains list is a plain items envelope WITHOUT pagination —
		// asserting this shape keeps the SDK honest about the contract.
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "dom_1", "domain": "a.example.com"},
				{"id": "dom_2", "domain": "b.example.com"},
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	list, err := client.Domains.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(list.Items))
	}
	if list.Items[1].Domain != "b.example.com" {
		t.Errorf("expected second domain 'b.example.com', got %q", list.Items[1].Domain)
	}
}

// TestDomainsService_Verify asserts the wire contract of the verify call:
// the server's VerifyDomainInput requires `domainId` in the request body
// (the URL id alone is not enough) — dropping it would 400 in production.
func TestDomainsService_Verify(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains/dom_1/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["domainId"] != "dom_1" {
			t.Errorf("expected body domainId 'dom_1' (required by VerifyDomainInput), got %v", body["domainId"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       "dom_1",
			"domain":   "mail.example.com",
			"status":   "VERIFYING",
			"verified": false,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	d, err := client.Domains.Verify(context.Background(), "dom_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != DomainStatusVerifying {
		t.Errorf("expected status VERIFYING, got %q", d.Status)
	}
}

func TestDomainsService_Update(t *testing.T) {
	enabled := true

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains/dom_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["feedbackEnabled"] != true {
			t.Errorf("expected feedbackEnabled true, got %v", body["feedbackEnabled"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              "dom_1",
			"domain":          "mail.example.com",
			"feedbackEnabled": true,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	d, err := client.Domains.Update(context.Background(), "dom_1", UpdateDomainParams{FeedbackEnabled: &enabled})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.FeedbackEnabled {
		t.Error("expected FeedbackEnabled true after update")
	}
}

// TestDomainsService_Update_OmitsUnsetFields asserts pointer semantics on
// PATCH: a nil field means "leave unchanged" and must be absent from the
// wire — sending false would silently flip the setting off.
func TestDomainsService_Update_OmitsUnsetFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains/dom_1", func(w http.ResponseWriter, r *http.Request) {
		body := decodeRawBody(t, r)
		if _, present := body["feedbackEnabled"]; present {
			t.Errorf("expected feedbackEnabled to be omitted when nil, but it was present: %v", body["feedbackEnabled"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "dom_1", "domain": "mail.example.com"})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	if _, err := client.Domains.Update(context.Background(), "dom_1", UpdateDomainParams{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDomainsService_Delete(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains/dom_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	if err := client.Domains.Delete(context.Background(), "dom_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDomainsService_DNSRecords(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains/dom_1/dns-records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"txt": map[string]interface{}{"name": "_anima.mail.example.com", "value": "anima-verify-abc123"},
			"mailFrom": map[string]interface{}{
				"name": "bounce.mail.example.com",
				"mx":   map[string]interface{}{"name": "bounce.mail.example.com", "value": "feedback-smtp.useanima.sh", "priority": 10},
				"spf":  "v=spf1 include:useanima.sh ~all",
			},
			"dkim": []map[string]interface{}{
				{"name": "anima._domainkey.mail.example.com", "value": "p=MIGfMA0..."},
			},
			"mx":    map[string]interface{}{"name": "mail.example.com", "value": "mx.useanima.sh", "priority": 10},
			"spf":   "v=spf1 include:useanima.sh ~all",
			"dmarc": "v=DMARC1; p=quarantine",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	records, err := client.Domains.DNSRecords(context.Background(), "dom_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The DNS record set is what a customer copies into their DNS provider —
	// every section must survive decoding, or domain setup breaks.
	if records.TXT.Value != "anima-verify-abc123" {
		t.Errorf("expected TXT value 'anima-verify-abc123', got %q", records.TXT.Value)
	}
	if records.MailFrom.MX.Priority != 10 {
		t.Errorf("expected mailFrom MX priority 10, got %d", records.MailFrom.MX.Priority)
	}
	if len(records.DKIM) != 1 || records.DKIM[0].Name != "anima._domainkey.mail.example.com" {
		t.Errorf("expected one DKIM record, got %+v", records.DKIM)
	}
	if records.MX.Value != "mx.useanima.sh" {
		t.Errorf("expected MX value 'mx.useanima.sh', got %q", records.MX.Value)
	}
	if records.DMARC != "v=DMARC1; p=quarantine" {
		t.Errorf("expected DMARC policy to round-trip, got %q", records.DMARC)
	}
}

func TestDomainsService_Deliverability(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains/dom_1/deliverability", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"domain":        "mail.example.com",
			"sent":          1000,
			"delivered":     978,
			"bounced":       20,
			"complained":    2,
			"bounceRate":    0.02,
			"complaintRate": 0.002,
			"isHealthy":     true,
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	stats, err := client.Domains.Deliverability(context.Background(), "dom_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Sent != 1000 || stats.Delivered != 978 {
		t.Errorf("expected sent=1000 delivered=978, got sent=%d delivered=%d", stats.Sent, stats.Delivered)
	}
	if stats.BounceRate != 0.02 {
		t.Errorf("expected bounceRate 0.02, got %v", stats.BounceRate)
	}
	if !stats.IsHealthy {
		t.Error("expected IsHealthy true")
	}
}

func TestDomainsService_ZoneFile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/domains/dom_1/zone-file", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"zoneFile": "; Anima DNS records for mail.example.com\n_anima.mail.example.com. IN TXT \"anima-verify-abc123\"\n",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	zf, err := client.Domains.ZoneFile(context.Background(), "dom_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zf.ZoneFile == "" || zf.ZoneFile[0] != ';' {
		t.Errorf("expected a zone-file export, got %q", zf.ZoneFile)
	}
}
