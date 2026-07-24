package anima

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestReadLoopForwardsEagerAndUnknownFrames locks the live-call read loop's
// forward-compat contract: the speculative eager end-of-turn hint
// (call.transcription.eager) and any unknown / newer frame type are BOTH
// forwarded to message handlers unchanged, and neither ends the read loop nor
// trips the error handler — the committed call.transcription still arrives
// after them. Consumers filter by type (+ isFinal); the transparent forwarder
// must not editorialize.
//
// This is the regression guard for the class of bug that crashed the MCP server
// (a type-switch with no default that returned undefined on an unrecognized
// frame). The Go client stays resilient because it never dispatches on type.
func TestReadLoopForwardsEagerAndUnknownFrames(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Barrier: wait for the client to register handlers and signal
		// readiness before sending frames, so no frame races handler setup.
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}

		frames := []string{
			`{"type":"call.transcription.eager","data":{"text":"hel"}}`,
			`{"type":"call.some.future.frame","data":{"x":1}}`,
			`{"type":"call.transcription","data":{"text":"hello","isFinal":true}}`,
		}
		for _, f := range frames {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(f)); err != nil {
				return
			}
		}

		// Keep the connection open until the client disconnects (vc.Close) so
		// the read loop can process every frame above.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	vc, err := newVoiceConnection(wsURL)
	if err != nil {
		t.Fatalf("newVoiceConnection: %v", err)
	}
	defer vc.Close()

	got := make(chan string, 8)
	vc.OnMessage(func(m VoiceMessage) { got <- m.Type })

	errCh := make(chan error, 4)
	vc.OnError(func(e error) { errCh <- e })

	// Release the server barrier now that handlers are registered.
	if err := vc.Send("client.ready", nil); err != nil {
		t.Fatalf("send ready: %v", err)
	}

	// All three frames are forwarded, in order — the eager and unknown frames
	// must neither end the loop nor be swallowed.
	want := []string{"call.transcription.eager", "call.some.future.frame", "call.transcription"}
	for i, expected := range want {
		select {
		case typ := <-got:
			if typ != expected {
				t.Fatalf("message %d: got type %q, want %q", i, typ, expected)
			}
		case e := <-errCh:
			t.Fatalf("unexpected error handler invocation: %v", e)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for message %d (%q); read loop may have ended", i, expected)
		}
	}
}
