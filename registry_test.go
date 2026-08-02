package anima

// Tests for RegistryService, centred on DID path encoding.
//
// Lookup, Update and Unlist all interpolate a DID into the path. DIDs contain
// colons, which are legal in a path segment — so raw interpolation looks fine
// and works for the common case. It breaks on the one that matters: a did:web
// carrying a port percent-encodes that colon, so the DID string itself
// contains "%3A". Interpolated raw, the server decodes it back to ":" and
// looks up a DIFFERENT DID.
//
// This SDK was interpolating raw. There was no registry test file in any of
// the three SDKs, which is why nobody noticed that node encoded and go and
// python did not — the same DID produced three different URLs.
//
// Note url.PathEscape leaves ":" alone (legal in a path segment) where node
// and python escape it to %3A. Both reach the server as the same DID; what
// matters is that "%" and "/" are escaped, and all three do that.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// A did:web with a port — the spec percent-encodes the colon before it.
const didWithPort = "did:web:localhost%3A3000:agents:a1"

// capturePath answers every request and records the raw path the client sent,
// before any router decoding.
func capturePath(t *testing.T, got *string) (*Client, func()) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		*got = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"did":     "did:web:x",
			"agentId": "agent_1",
			"name":    "Test",
		})
	})
	client, ts := newTestClient(mux)
	return client, ts.Close
}

func TestRegistryService_LookupEncodesTheDID(t *testing.T) {
	var path string
	client, done := capturePath(t, &path)
	defer done()

	if _, err := client.Registry.Lookup(context.Background(), didWithPort); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// %3A must arrive as %253A. Otherwise the router decodes it to ":" and
	// resolves did:web:localhost:3000:agents:a1 — a DID nobody registered.
	if !strings.Contains(path, "%253A3000") {
		t.Errorf("port colon was not escaped; sent %q", path)
	}
	segment := strings.TrimPrefix(path, "/v1/registry/agents/")
	decoded, err := url.PathUnescape(segment)
	if err != nil {
		t.Fatalf("path segment does not decode: %v", err)
	}
	if decoded != didWithPort {
		t.Errorf("DID did not survive the round trip: got %q, want %q", decoded, didWithPort)
	}
}

func TestRegistryService_UpdateAndUnlistEncodeTheDID(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*Client) error
	}{
		{"Update", func(c *Client) error {
			_, err := c.Registry.Update(context.Background(), didWithPort, UpdateRegistryAgentParams{Name: "x"})
			return err
		}},
		{"Unlist", func(c *Client) error {
			return c.Registry.Unlist(context.Background(), didWithPort)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var path string
			client, done := capturePath(t, &path)
			defer done()

			if err := tc.call(client); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(path, "%253A3000") {
				t.Errorf("%s did not escape the port colon; sent %q", tc.name, path)
			}
		})
	}
}

// A stray slash would otherwise split the path and hit a different route.
func TestRegistryService_SlashCannotSplitThePath(t *testing.T) {
	var path string
	client, done := capturePath(t, &path)
	defer done()

	if _, err := client.Registry.Lookup(context.Background(), "did:web:a/b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Count(path, "/") != 4 {
		t.Errorf("slash was not escaped — path gained a segment: %q", path)
	}
	if !strings.HasSuffix(path, "%2Fb") {
		t.Errorf("expected the slash escaped as %%2F, got %q", path)
	}
}
