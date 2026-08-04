package anima

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Every path this SDK can call must exist on the API.
//
// The per-service tests assert paths against an httptest mux, so they pass
// whether or not the real server serves the route. That is how
// POST /security/scan and GET /identity/did/{did} stayed green for the whole
// life of those methods, and how two entire services (wallet, pods) shipped
// against products the API had removed. Thirty of this SDK's calls went
// nowhere.
//
// This test reads the service sources and checks each path literal against
// apiRoutes below, generated from the monorepo's
// packages/contracts/src/contracts/*.ts at the commit in .anima-ref. A new
// method reaching for a route that does not exist fails here.
//
// WHAT THIS PROVES, AND WHAT IT DOES NOT. A path in apiRoutes is DECLARED in
// the contracts. That is not the same as "the product is alive". Killed
// surfaces are routinely left in place — anima 894035bc deleted the OAuth
// console pages and says the procedures "remain dormant in packages/contracts
// + apps/api/routes/handlers/vault.ts; full backend cleanup is a separate
// chore" — and some are declared `deprecated: true` and answer 400 by design
// (POST /mcp-auth/sessions). So a gap between this list and the SDK is NOT by
// itself evidence of missing coverage. In 2026-08 six dormant vault OAuth
// routes were read that way and re-added to all three SDKs; they were removed
// again. Check git history for a deliberate removal before filling any gap.
//
// Regenerate when .anima-ref moves, from the monorepo root:
//
//	python3 - <<'EOF'
//	import re, pathlib
//	src = "".join(p.read_text() for p in
//	              sorted(pathlib.Path("packages/contracts/src/contracts").glob("*.ts")))
//	out = set()
//	for m in re.finditer(r'\.route\(\s*\{(.*?)\}\s*\)', src, re.S):
//	    meth = re.search(r'method:\s*"([A-Z]+)"', m.group(1))
//	    path = re.search(r'path:\s*"([^"]+)"', m.group(1))
//	    if meth and path:
//	        out.add(meth.group(1) + " " + re.sub(r'\{[^}]*\}', '*', path.group(1)))
//	print("\n".join(sorted(out)))
//	EOF
//
// It must parse the whole `.route({...})` object, not one line: eight routes
// spread method and path across lines (POST /addresses/*/validate, the four
// mcp-auth ones, three agents/*/email-identities ones). The single-line grep
// this header used to recommend silently dropped all eight.
//
// Then add the three routes registered directly on fastify rather than through
// oRPC, which no contracts scan can see: GET /audit/events, GET /events/stream,
// POST /a2a/inbound.

// apiRoutes is "METHOD /path" for every route the API serves, path params as *.
var apiRoutes = map[string]struct{}{
	"POST /a2a/inbound":                             {},
	"GET /addresses":                                {},
	"POST /addresses":                               {},
	"DELETE /addresses/*":                           {},
	"GET /addresses/*":                              {},
	"PUT /addresses/*":                              {},
	"POST /addresses/*/validate":                    {},
	"POST /agent/elevate":                           {},
	"POST /agent/elevate/request":                   {},
	"POST /agent/sign-up":                           {},
	"GET /agent/status":                             {},
	"POST /agent/verify":                            {},
	"GET /agents":                                   {},
	"POST /agents":                                  {},
	"DELETE /agents/*":                              {},
	"GET /agents/*":                                 {},
	"PATCH /agents/*":                               {},
	"POST /agents/*/a2a/dispatch":                   {},
	"GET /agents/*/a2a/tasks":                       {},
	"POST /agents/*/a2a/tasks":                      {},
	"GET /agents/*/a2a/tasks/*":                     {},
	"POST /agents/*/a2a/tasks/*/cancel":             {},
	"POST /agents/*/a2a/tasks/*/update":             {},
	"GET /agents/*/card":                            {},
	"GET /agents/*/credentials":                     {},
	"POST /agents/*/credentials":                    {},
	"POST /agents/*/credentials/*/revoke":           {},
	"GET /agents/*/did":                             {},
	"POST /agents/*/did/rotate":                     {},
	"GET /agents/*/email-identities":                {},
	"POST /agents/*/email-identities":               {},
	"DELETE /agents/*/email-identities/*":           {},
	"POST /agents/*/email-identities/*/set-primary": {},
	"POST /agents/*/email-identities/*/verify":      {},
	"GET /agents/*/policy":                          {},
	"PUT /agents/*/policy":                          {},
	"POST /agents/*/rotate-key":                     {},
	"GET /api-keys":                                 {},
	"POST /api-keys":                                {},
	"DELETE /api-keys/*":                            {},
	"PATCH /api-keys/*":                             {},
	"POST /api-keys/*/rotate":                       {},
	"GET /api-keys/scopes":                          {},
	"GET /attachments/*/download":                   {},
	"GET /attachments/*/text":                       {},
	"GET /audit/events":                             {},
	"POST /billing/change-plan":                     {},
	"POST /billing/checkout":                        {},
	"POST /billing/contact-enterprise":              {},
	"GET /billing/features/*":                       {},
	"GET /billing/invoices":                         {},
	"GET /billing/overage":                          {},
	"PUT /billing/overage":                          {},
	"GET /billing/plans":                            {},
	"POST /billing/portal":                          {},
	"GET /billing/tier":                             {},
	"GET /billing/usage":                            {},
	"POST /demo/inbox":                              {},
	"GET /demo/inbox/*/messages":                    {},
	"POST /demo/inbox/*/self-test":                  {},
	"GET /domains":                                  {},
	"POST /domains":                                 {},
	"DELETE /domains/*":                             {},
	"GET /domains/*":                                {},
	"PATCH /domains/*":                              {},
	"GET /domains/*/deliverability":                 {},
	"GET /domains/*/dns-records":                    {},
	"POST /domains/*/verify":                        {},
	"GET /domains/*/zone-file":                      {},
	"GET /email":                                    {},
	"GET /email-rules":                              {},
	"POST /email-rules":                             {},
	"DELETE /email-rules/*":                         {},
	"POST /email-rules/evaluate":                    {},
	"GET /email/*":                                  {},
	"POST /email/*/forward":                         {},
	"POST /email/*/reply":                           {},
	"GET /email/drafts":                             {},
	"POST /email/drafts":                            {},
	"DELETE /email/drafts/*":                        {},
	"GET /email/drafts/*":                           {},
	"POST /email/drafts/*/send":                     {},
	"POST /email/send":                              {},
	"GET /email/suppressions":                       {},
	"POST /email/unsuppress":                        {},
	"GET /events/stream":                            {},
	"POST /extension/connect":                       {},
	"POST /extension/exchange":                      {},
	"POST /extension/revoke":                        {},
	"GET /extension/settings":                       {},
	"PATCH /extension/settings":                     {},
	"POST /extension/token":                         {},
	"PATCH /extension/token/*":                      {},
	"GET /feedback":                                 {},
	"POST /feedback":                                {},
	"POST /identities":                              {},
	"POST /identity/verify":                         {},
	"GET /inboxes":                                  {},
	"POST /inboxes":                                 {},
	"DELETE /inboxes/*":                             {},
	"GET /inboxes/*":                                {},
	"PATCH /inboxes/*":                              {},
	"GET /invoices":                                 {},
	"GET /invoices/*":                               {},
	"PATCH /invoices/*":                             {},
	"GET /invoices/export":                          {},
	"POST /invoices/match-receipts":                 {},
	"GET /invoices/reconciliation-summary":          {},
	"POST /mcp-auth/sessions":                       {},
	"POST /mcp-auth/sessions/*/complete":            {},
	"POST /mcp-auth/sessions/*/deny":                {},
	"POST /mcp-auth/sessions/poll":                  {},
	"GET /me/agents":                                {},
	"GET /me/orgs":                                  {},
	"GET /messages":                                 {},
	"GET /messages/*":                               {},
	"POST /messages/*/attachments":                  {},
	"PATCH /messages/*/labels":                      {},
	"POST /messages/email":                          {},
	"POST /messages/search":                         {},
	"POST /messages/search/semantic":                {},
	"POST /messages/sms":                            {},
	"POST /oauth/apps":                              {},
	"GET /oauth/apps/*":                             {},
	"PATCH /oauth/apps/*":                           {},
	"POST /oauth/auth-codes/mint":                   {},
	"POST /oauth/register":                          {},
	"POST /oauth/revoke":                            {},
	"POST /oauth/token":                             {},
	"GET /oauth/userinfo":                           {},
	"GET /openclaw/agents":                          {},
	"GET /openclaw/authorize":                       {},
	"GET /openclaw/callback":                        {},
	"POST /openclaw/signup":                         {},
	"GET /orgs":                                     {},
	"POST /orgs":                                    {},
	"DELETE /orgs/*":                                {},
	"GET /orgs/*":                                   {},
	"PATCH /orgs/*":                                 {},
	"GET /orgs/*/access-reviews":                    {},
	"POST /orgs/*/access-reviews":                   {},
	"POST /orgs/*/access-reviews/*/complete":        {},
	"GET /orgs/*/agents/*/baselines":                {},
	"DELETE /orgs/*/agents/*/quarantine":            {},
	"GET /orgs/*/agents/*/quarantine":               {},
	"POST /orgs/*/agents/*/quarantine":              {},
	"GET /orgs/*/anomaly-alerts":                    {},
	"GET /orgs/*/anomaly-alerts/*":                  {},
	"POST /orgs/*/anomaly-alerts/*/acknowledge":     {},
	"POST /orgs/*/anomaly-alerts/*/false-positive":  {},
	"POST /orgs/*/anomaly-alerts/*/resolve":         {},
	"GET /orgs/*/anomaly-rules":                     {},
	"POST /orgs/*/anomaly-rules":                    {},
	"DELETE /orgs/*/anomaly-rules/*":                {},
	"PATCH /orgs/*/anomaly-rules/*":                 {},
	"GET /orgs/*/audit-logs":                        {},
	"GET /orgs/*/audit-logs/*":                      {},
	"POST /orgs/*/audit-logs/export":                {},
	"POST /orgs/*/claim":                            {},
	"GET /orgs/*/compliance/controls":               {},
	"GET /orgs/*/compliance/controls/*":             {},
	"PATCH /orgs/*/compliance/controls/*":           {},
	"POST /orgs/*/compliance/controls/*/collect":    {},
	"GET /orgs/*/compliance/controls/*/evidence":    {},
	"POST /orgs/*/compliance/controls/*/evidence":   {},
	"GET /orgs/*/compliance/dashboard":              {},
	"GET /orgs/*/compliance/dsars":                  {},
	"POST /orgs/*/compliance/dsars":                 {},
	"GET /orgs/*/compliance/dsars/*":                {},
	"PATCH /orgs/*/compliance/dsars/*":              {},
	"GET /orgs/*/compliance/reports":                {},
	"POST /orgs/*/compliance/reports":               {},
	"DELETE /orgs/*/compliance/reports/*":           {},
	"GET /orgs/*/compliance/reports/*":              {},
	"POST /orgs/*/compliance/reports/*/export":      {},
	"POST /orgs/*/compliance/seed":                  {},
	"GET /orgs/*/compliance/summary":                {},
	"GET /orgs/*/compliance/templates":              {},
	"GET /orgs/*/members":                           {},
	"POST /orgs/*/messages/*/approve":               {},
	"POST /orgs/*/rotate-key":                       {},
	"GET /orgs/*/security/events":                   {},
	"GET /orgs/*/security/scanner-status":           {},
	"GET /orgs/claimable":                           {},
	"POST /orgs/feature-interest":                   {},
	"GET /orgs/me":                                  {},
	"GET /orgs/me/usage":                            {},
	"GET /orgs/me/workspace-health":                 {},
	"GET /phone/numbers":                            {},
	"POST /phone/provision":                         {},
	"POST /phone/release":                           {},
	"GET /phone/requirements":                       {},
	"GET /phone/search":                             {},
	"POST /phone/send-sms":                          {},
	"GET /phone/sms-suppressions":                   {},
	"POST /phone/sms-unsuppress":                    {},
	"GET /phone/sms/threads":                        {},
	"GET /phone/sms/threads/*":                      {},
	// Provisioning requests — how an agent asks its owner for a vault or a
	// phone number, neither of which it can provision itself.
	"GET /provisioning-requests":               {},
	"POST /provisioning-requests":              {},
	"GET /provisioning-requests/*":             {},
	"POST /provisioning-requests/*/approve":    {},
	"POST /provisioning-requests/*/cancel":     {},
	"POST /provisioning-requests/*/decline":    {},
	"POST /registry/agents":                    {},
	"DELETE /registry/agents/*":                {},
	"GET /registry/agents/*":                   {},
	"PUT /registry/agents/*":                   {},
	"GET /registry/agents/search":              {},
	"GET /scoped-tokens":                       {},
	"POST /scoped-tokens":                      {},
	"POST /scoped-tokens/revoke":               {},
	"GET /threads":                             {},
	"GET /vault/audit":                         {},
	"GET /vault/credential-requests":           {},
	"POST /vault/credential-requests":          {},
	"GET /vault/credential-requests/*":         {},
	"POST /vault/credential-requests/*/cancel": {},
	"GET /vault/credentials":                   {},
	"POST /vault/credentials":                  {},
	"DELETE /vault/credentials/*":              {},
	"GET /vault/credentials/*":                 {},
	"PUT /vault/credentials/*":                 {},
	"POST /vault/credentials/*/use":            {},
	"POST /vault/deprovision":                  {},
	"POST /vault/generate-password":            {},
	"GET /vault/identities":                    {},
	"GET /vault/oauth/accounts":                {},
	"DELETE /vault/oauth/accounts/*":           {},
	"GET /vault/oauth/apps":                    {},
	"GET /vault/oauth/apps/*":                  {},
	"POST /vault/oauth/apps/*/custom":          {},
	"DELETE /vault/oauth/apps/*/custom/*":      {},
	"POST /vault/oauth/link":                   {},
	"GET /vault/oauth/link/*":                  {},
	"POST /vault/oauth/require-auth":           {},
	"POST /vault/provision":                    {},
	"GET /vault/search":                        {},
	"POST /vault/share":                        {},
	"POST /vault/share/revoke":                 {},
	"GET /vault/shares":                        {},
	"GET /vault/status":                        {},
	"POST /vault/sync":                         {},
	"POST /vault/token":                        {},
	"POST /vault/token/exchange":               {},
	"POST /vault/token/revoke":                 {},
	"GET /vault/totp/*":                        {},
	"GET /voice/analytics":                     {},
	"GET /voice/calls":                         {},
	"POST /voice/calls":                        {},
	"GET /voice/calls/*":                       {},
	"GET /voice/calls/*/recording":             {},
	"GET /voice/calls/*/score":                 {},
	"GET /voice/calls/*/security":              {},
	"GET /voice/calls/*/summary":               {},
	"GET /voice/calls/*/transcript":            {},
	"GET /voice/catalog":                       {},
	"POST /voice/search":                       {},
	"POST /voice/search/cross-channel":         {},
	"GET /webhooks":                            {},
	"POST /webhooks":                           {},
	"DELETE /webhooks/*":                       {},
	"GET /webhooks/*":                          {},
	"PUT /webhooks/*":                          {},
	"GET /webhooks/*/dead-letters":             {},
	"GET /webhooks/*/deliveries":               {},
	"POST /webhooks/*/reenable":                {},
	"POST /webhooks/*/rotate-secret":           {},
	"GET /webhooks/*/stats":                    {},
	"POST /webhooks/*/test":                    {},
	"POST /webhooks/deliveries/*/replay":       {},
	"GET /webhooks/event-types":                {},
}

// Matches a method constant followed by a path literal. A2AService.Discover is
// deliberately not matched: it builds a URL from a caller-supplied foreign host
// (/.well-known/agent.json) rather than calling an Anima route, so it has no
// business in the allowlist.
var callPattern = regexp.MustCompile(`http\.Method([A-Za-z]+),\s*(?:fmt\.Sprintf\(\s*)?"(/[^"]*)"`)

// pathParams differ per call site; compare shapes, not ids.
var pathParam = regexp.MustCompile(`%[sdv]`)

// Catches a path built by concatenation rather than a single fmt.Sprintf.
// callPattern stops at the closing quote of the first literal, so
// `"/voice/calls/"+id+"/transcript"` was scanned as GET /voice/calls — a real
// route, so it passed while the segment after the id went unchecked entirely.
// Anything after the first `+` is invisible to the guard.
//
// Both forms must be caught: matching only the bare literal leaves
// `fmt.Sprintf("/voice/calls/%s", id)+"/transcript"` free to reopen the hole.
//
// Residual gap: a path that does not start with a literal at all
// (`http.MethodGet, basePath+"/x"`) matches neither this nor callPattern, so
// it is never checked. The call-count floor only catches a wholesale
// regression, not one added method.
var concatPath = regexp.MustCompile(
	`http\.Method[A-Za-z]+,\s*(?:fmt\.Sprintf\(\s*"(/[^"]*)"[^)]*\)|"(/[^"]*)")\s*\+`,
)

type sdkCall struct {
	file  string
	route string
}

func collectSDKCalls(t *testing.T) []sdkCall {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	var calls []sdkCall
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		source, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, m := range callPattern.FindAllStringSubmatch(string(source), -1) {
			path := strings.TrimSuffix(pathParam.ReplaceAllString(m[2], "*"), "/")
			if path == "" {
				path = "/"
			}
			calls = append(calls, sdkCall{file: name, route: strings.ToUpper(m[1]) + " " + path})
		}
	}
	return calls
}

// If callPattern ever stops matching, the check below would pass vacuously.
func TestRouteScanFoundTheServiceCalls(t *testing.T) {
	if got := len(collectSDKCalls(t)); got < 110 {
		t.Fatalf("expected the scan to find >=110 calls, found %d — has the call pattern drifted?", got)
	}
}

// A concatenated path hides everything after the first `+` from the guard, so
// the route below it is never verified. Keep every path a single literal.
func TestNoServiceBuildsAPathByConcatenation(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		source, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, m := range concatPath.FindAllStringSubmatch(string(source), -1) {
			prefix := m[1]
			if prefix == "" {
				prefix = m[2]
			}
			t.Errorf(
				"%s builds a path by concatenating onto %q — the route guard cannot see past the `+`. Build the whole path in one fmt.Sprintf.",
				name, prefix,
			)
		}
	}
}

func TestNoServiceCallsARouteTheAPIDoesNotServe(t *testing.T) {
	for _, call := range collectSDKCalls(t) {
		if _, ok := apiRoutes[call.route]; !ok {
			t.Errorf("%s calls %q, which the API does not serve", call.file, call.route)
		}
	}
}
