package anima

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// Pins ComplianceService to the API contract.
//
// Every call this service could make was rejected. The enums were declared in
// lowercase against a contract that validates SCREAMING_SNAKE, CreateDSARInput
// sent `requestType` where the API reads `type`, and two routes did not exist:
// DownloadReport issued GET /reports/{id}/download and CompleteDSAR POSTed to
// /dsars/{id}/complete, against an API serving POST /reports/{id}/export and
// PATCH /dsars/{id}.
//
// There was no test file for this service, which is why none of it surfaced.
// The daily drift canary does not cover it either — that fires when the
// monorepo's contracts CHANGE, and these values were wrong from the first
// commit. This is the complement: it catches the SDK drifting from a contract
// that stayed put.
//
// Source of truth, at the commit in .anima-ref:
//
//	packages/contracts/src/schemas/compliance.ts
//	packages/contracts/src/schemas/compliance-controls.ts
//	packages/contracts/src/contracts/compliance.ts

func TestComplianceEnumsAreUppercase(t *testing.T) {
	values := []string{
		string(ComplianceFrameworkSOC2), string(ComplianceFrameworkGDPR), string(ComplianceFrameworkPCI),
		string(ComplianceControlStatusNotStarted), string(ComplianceControlStatusInProgress),
		string(ComplianceControlStatusImplemented), string(ComplianceControlStatusVerified),
		string(ComplianceControlStatusFailed),
		string(ComplianceReportTypeSOC2Summary), string(ComplianceReportTypeActivityReport),
		string(ComplianceReportTypeAccessReview), string(ComplianceReportTypeAuditExport),
		string(ComplianceReportTypeGDPRDSAR),
		string(ComplianceReportStatusPending), string(ComplianceReportStatusGenerating),
		string(ComplianceReportStatusCompleted), string(ComplianceReportStatusFailed),
		string(ComplianceReportFormatJSON), string(ComplianceReportFormatCSV), string(ComplianceReportFormatPDF),
		string(DSARTypeAccess), string(DSARTypeDelete), string(DSARTypeRectify),
		string(DSARTypePortability), string(DSARTypeRestrict),
		string(DSARStatusReceived), string(DSARStatusVerified), string(DSARStatusInProgress),
		string(DSARStatusCompleted), string(DSARStatusDenied), string(DSARStatusOverdue),
	}
	for _, v := range values {
		if v != strings.ToUpper(v) {
			t.Errorf("enum value %q is not uppercase; the API rejects it", v)
		}
	}
}

func TestDSARTypesMatchContract(t *testing.T) {
	got := []DSARType{DSARTypeAccess, DSARTypeDelete, DSARTypeRectify, DSARTypePortability, DSARTypeRestrict}
	want := []string{"ACCESS", "DELETE", "RECTIFY", "PORTABILITY", "RESTRICT"}
	if len(got) != len(want) {
		t.Fatalf("expected %d DSAR types, got %d", len(want), len(got))
	}
	for i := range want {
		if string(got[i]) != want[i] {
			t.Errorf("DSAR type %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestDSARStatusesMatchContract(t *testing.T) {
	// OVERDUE and DENIED were missing entirely; a caller could not represent a
	// request that had blown its GDPR deadline.
	want := []string{"RECEIVED", "VERIFIED", "IN_PROGRESS", "COMPLETED", "DENIED", "OVERDUE"}
	got := []DSARStatus{
		DSARStatusReceived, DSARStatusVerified, DSARStatusInProgress,
		DSARStatusCompleted, DSARStatusDenied, DSARStatusOverdue,
	}
	for i := range want {
		if string(got[i]) != want[i] {
			t.Errorf("DSAR status %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func dsarResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           "dsar_1",
		"orgId":        "org_1",
		"type":         "ACCESS",
		"status":       "RECEIVED",
		"subjectEmail": "subject@example.com",
		"subjectName":  nil,
		"subjectId":    nil,
		"description":  nil,
		"requestedAt":  "2026-07-31T00:00:00Z",
		"verifiedAt":   nil,
		"dueAt":        "2026-08-30T00:00:00Z",
		"completedAt":  nil,
		"processedBy":  nil,
		"response":     nil,
		"metadata":     map[string]interface{}{},
		"createdAt":    "2026-07-31T00:00:00Z",
		"updatedAt":    "2026-07-31T00:00:00Z",
	}
}

func TestComplianceService_CreateDSAR_SendsTypeNotRequestType(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orgs/org_1/compliance/dsars", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["type"] != "ACCESS" {
			t.Errorf("expected type 'ACCESS', got %v", body["type"])
		}
		if _, ok := body["requestType"]; ok {
			t.Error("body carries requestType; the API reads type and rejects the request")
		}
		if body["dueInDays"] != float64(14) {
			t.Errorf("expected dueInDays 14, got %v", body["dueInDays"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dsarResponse())
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	d, err := client.Compliance.CreateDSAR(context.Background(), "org_1", CreateDSARInput{
		Type:         DSARTypeAccess,
		SubjectEmail: "subject@example.com",
		DueInDays:    14,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Type != DSARTypeAccess {
		t.Errorf("expected type ACCESS, got %q", d.Type)
	}
	// DueAt is the GDPR clock; the old model dropped it entirely.
	if d.DueAt != "2026-08-30T00:00:00Z" {
		t.Errorf("expected dueAt to round-trip, got %q", d.DueAt)
	}
}

func TestComplianceService_UpdateDSARStatus_PatchesTheDSAR(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orgs/org_1/compliance/dsars/dsar_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["status"] != "COMPLETED" {
			t.Errorf("expected status 'COMPLETED', got %v", body["status"])
		}
		resp := dsarResponse()
		resp["status"] = "COMPLETED"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	// The old client POSTed here. Registering it proves we no longer do.
	mux.HandleFunc("/v1/orgs/org_1/compliance/dsars/dsar_1/complete", func(w http.ResponseWriter, r *http.Request) {
		t.Error("called /complete, which the API does not serve")
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	d, err := client.Compliance.UpdateDSARStatus(context.Background(), "org_1", "dsar_1", UpdateDSARStatusInput{
		Status: DSARStatusCompleted,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != DSARStatusCompleted {
		t.Errorf("expected status COMPLETED, got %q", d.Status)
	}
}

func TestComplianceService_GetDSAR(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orgs/org_1/compliance/dsars/dsar_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dsarResponse())
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	d, err := client.Compliance.GetDSAR(context.Background(), "org_1", "dsar_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID != "dsar_1" {
		t.Errorf("expected ID 'dsar_1', got %q", d.ID)
	}
}

func TestComplianceService_ExportReport_PostsToExport(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orgs/org_1/compliance/reports/rep_1/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["format"] != "PDF" {
			t.Errorf("expected format 'PDF', got %v", body["format"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":        "e30=",
			"contentType": "application/pdf",
			"filename":    "soc2.pdf",
		})
	})
	mux.HandleFunc("/v1/orgs/org_1/compliance/reports/rep_1/download", func(w http.ResponseWriter, r *http.Request) {
		t.Error("called /download, which the API does not serve")
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	out, err := client.Compliance.ExportReport(context.Background(), "org_1", "rep_1", &ExportReportInput{
		Format: ComplianceReportFormatPDF,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The export arrives inline. The old model expected a signed URL.
	if out.Filename != "soc2.pdf" || out.Data == "" {
		t.Errorf("expected inline export data and filename, got %+v", out)
	}
}

func TestComplianceService_GenerateReport_SendsPeriodFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orgs/org_1/compliance/reports", func(w http.ResponseWriter, r *http.Request) {
		body := decodeRawBody(t, r)
		if body["type"] != "SOC2_SUMMARY" {
			t.Errorf("expected type 'SOC2_SUMMARY', got %v", body["type"])
		}
		if body["periodStart"] != "2026-07-01" {
			t.Errorf("expected periodStart, got %v", body["periodStart"])
		}
		// `from`/`to` were the old invented field names.
		if _, ok := body["from"]; ok {
			t.Error("body carries `from`; the contract field is periodStart")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "rep_1", "orgId": "org_1", "type": "SOC2_SUMMARY",
			"title": "Q3", "description": nil, "status": "COMPLETED", "format": "JSON",
			"parameters": map[string]interface{}{}, "content": nil, "errorMessage": nil,
			"generatedBy": nil, "periodStart": "2026-07-01", "periodEnd": "2026-09-30",
			"completedAt": "2026-07-31T00:00:00Z",
			"createdAt":   "2026-07-31T00:00:00Z", "updatedAt": "2026-07-31T00:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	rep, err := client.Compliance.GenerateReport(context.Background(), "org_1", GenerateReportInput{
		Type:        ComplianceReportTypeSOC2Summary,
		PeriodStart: "2026-07-01",
		PeriodEnd:   "2026-09-30",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Status != ComplianceReportStatusCompleted {
		t.Errorf("expected status COMPLETED, got %q", rep.Status)
	}
	if rep.Format != ComplianceReportFormatJSON {
		t.Errorf("expected format JSON, got %q", rep.Format)
	}
}

func TestComplianceService_UpdateControlStatus_SendsUppercase(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orgs/org_1/compliance/controls/CC6.1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		body := decodeRawBody(t, r)
		if body["status"] != "IMPLEMENTED" {
			t.Errorf("expected status 'IMPLEMENTED', got %v", body["status"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "ctl_1", "orgId": "org_1", "framework": "SOC2", "controlId": "CC6.1",
			"title": "Logical access", "description": "Access is restricted",
			"category": "CC6", "status": "IMPLEMENTED", "owner": nil,
			"lastTestedAt": nil, "nextReviewAt": nil,
			"createdAt": "2026-07-31T00:00:00Z", "updatedAt": "2026-07-31T00:00:00Z",
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	ctrl, err := client.Compliance.UpdateControlStatus(context.Background(), "org_1", "CC6.1", ComplianceControlStatusInput{
		Status: ComplianceControlStatusImplemented,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctrl.Category != ComplianceControlCategoryCC6 {
		t.Errorf("expected category CC6, got %q", ctrl.Category)
	}
}

func TestComplianceService_ListTemplates(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/orgs/org_1/compliance/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"type": "SOC2_SUMMARY", "title": "SOC 2", "description": "Controls and evidence"},
			},
		})
	})

	client, ts := newTestClient(mux)
	defer ts.Close()

	out, err := client.Compliance.ListTemplates(context.Background(), "org_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].Type != "SOC2_SUMMARY" {
		t.Errorf("expected one SOC2_SUMMARY template, got %+v", out.Items)
	}
}
