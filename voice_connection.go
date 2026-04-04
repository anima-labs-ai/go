package anima

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const pingInterval = 30 * time.Second

// VoiceMessage represents a message on the voice WebSocket.
type VoiceMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
}

// VoiceConnectionOptions configures a voice WebSocket connection.
type VoiceConnectionOptions struct {
	AgentID string
}

// VoiceConnection is a bidirectional WebSocket connection for real-time voice call control.
//
// Send commands (call.create, call.speak, call.hangup) and receive events
// (call.started, call.transcription, call.ended) over a persistent connection.
type VoiceConnection struct {
	wsURL string
	conn  *websocket.Conn
	mu    sync.Mutex

	closed   bool
	closedMu sync.RWMutex

	pingDone chan struct{}

	messageHandlers      []func(VoiceMessage)
	errorHandlers        []func(error)
	connectedHandlers    []func()
	disconnectedHandlers []func()
	handlerMu            sync.RWMutex
}

// newVoiceConnection creates and connects a VoiceConnection.
func newVoiceConnection(wsURL string) (*VoiceConnection, error) {
	vc := &VoiceConnection{
		wsURL: wsURL,
	}
	if err := vc.connect(); err != nil {
		return nil, err
	}
	return vc, nil
}

// OnMessage registers a handler for incoming voice messages.
func (vc *VoiceConnection) OnMessage(handler func(VoiceMessage)) {
	vc.handlerMu.Lock()
	vc.messageHandlers = append(vc.messageHandlers, handler)
	vc.handlerMu.Unlock()
}

// OnError registers a handler for WebSocket errors.
func (vc *VoiceConnection) OnError(handler func(error)) {
	vc.handlerMu.Lock()
	vc.errorHandlers = append(vc.errorHandlers, handler)
	vc.handlerMu.Unlock()
}

// OnConnected registers a handler called when the connection opens.
func (vc *VoiceConnection) OnConnected(handler func()) {
	vc.handlerMu.Lock()
	vc.connectedHandlers = append(vc.connectedHandlers, handler)
	vc.handlerMu.Unlock()
}

// OnDisconnected registers a handler called when the connection closes.
func (vc *VoiceConnection) OnDisconnected(handler func()) {
	vc.handlerMu.Lock()
	vc.disconnectedHandlers = append(vc.disconnectedHandlers, handler)
	vc.handlerMu.Unlock()
}

// Send sends a raw message to the voice WebSocket.
func (vc *VoiceConnection) Send(msgType string, data map[string]interface{}) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.conn == nil {
		return nil
	}

	msg := VoiceMessage{Type: msgType, Data: data}
	return vc.conn.WriteJSON(msg)
}

// CreateCall creates an outbound call.
func (vc *VoiceConnection) CreateCall(to string, opts *CreateCallOptions) error {
	data := map[string]interface{}{"to": to}
	if opts != nil {
		if opts.Tier != "" {
			data["tier"] = opts.Tier
		}
		if opts.Greeting != "" {
			data["greeting"] = opts.Greeting
		}
		if opts.FromNumber != "" {
			data["fromNumber"] = opts.FromNumber
		}
	}
	return vc.Send("call.create", data)
}

// CreateCallOptions contains optional parameters for creating a call via WebSocket.
type CreateCallOptions struct {
	Tier       string
	Greeting   string
	FromNumber string
}

// Speak sends text for TTS playback.
func (vc *VoiceConnection) Speak(callID, text string) error {
	return vc.Send("call.speak", map[string]interface{}{"callId": callID, "text": text})
}

// CancelSpeak cancels in-progress speech.
func (vc *VoiceConnection) CancelSpeak(callID string) error {
	return vc.Send("call.speak.cancel", map[string]interface{}{"callId": callID})
}

// Hangup hangs up a call.
func (vc *VoiceConnection) Hangup(callID string) error {
	return vc.Send("call.hangup", map[string]interface{}{"callId": callID})
}

// Accept accepts an inbound call.
func (vc *VoiceConnection) Accept(callID string) error {
	return vc.Send("call.accept", map[string]interface{}{"callId": callID})
}

// Reject rejects an inbound call.
func (vc *VoiceConnection) Reject(callID string) error {
	return vc.Send("call.reject", map[string]interface{}{"callId": callID})
}

// Hold places a call on hold.
func (vc *VoiceConnection) Hold(callID string) error {
	return vc.Send("call.hold", map[string]interface{}{"callId": callID})
}

// Resume resumes a held call.
func (vc *VoiceConnection) Resume(callID string) error {
	return vc.Send("call.resume", map[string]interface{}{"callId": callID})
}

// DTMF sends DTMF tone(s).
func (vc *VoiceConnection) DTMF(callID, digits string) error {
	return vc.Send("call.dtmf", map[string]interface{}{"callId": callID, "digits": digits})
}

// Close closes the WebSocket connection.
func (vc *VoiceConnection) Close() error {
	vc.closedMu.Lock()
	vc.closed = true
	vc.closedMu.Unlock()

	if vc.pingDone != nil {
		close(vc.pingDone)
	}

	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.conn != nil {
		err := vc.conn.Close()
		vc.conn = nil
		return err
	}
	return nil
}

func (vc *VoiceConnection) connect() error {
	vc.closedMu.RLock()
	if vc.closed {
		vc.closedMu.RUnlock()
		return nil
	}
	vc.closedMu.RUnlock()

	conn, _, err := websocket.DefaultDialer.Dial(vc.wsURL, nil)
	if err != nil {
		return err
	}

	vc.mu.Lock()
	vc.conn = conn
	vc.mu.Unlock()

	// Notify connected handlers.
	vc.handlerMu.RLock()
	for _, h := range vc.connectedHandlers {
		h()
	}
	vc.handlerMu.RUnlock()

	// Start ping goroutine.
	vc.pingDone = make(chan struct{})
	go vc.pingLoop()

	// Start read loop in background goroutine.
	go vc.readLoop()

	return nil
}

func (vc *VoiceConnection) readLoop() {
	for {
		vc.mu.Lock()
		c := vc.conn
		vc.mu.Unlock()

		if c == nil {
			return
		}

		_, raw, err := c.ReadMessage()
		if err != nil {
			vc.closedMu.RLock()
			isClosed := vc.closed
			vc.closedMu.RUnlock()

			if !isClosed {
				vc.handlerMu.RLock()
				for _, h := range vc.errorHandlers {
					h(err)
				}
				vc.handlerMu.RUnlock()
			}

			// Notify disconnected handlers.
			vc.handlerMu.RLock()
			for _, h := range vc.disconnectedHandlers {
				h()
			}
			vc.handlerMu.RUnlock()
			return
		}

		var msg VoiceMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		vc.handlerMu.RLock()
		for _, h := range vc.messageHandlers {
			h(msg)
		}
		vc.handlerMu.RUnlock()
	}
}

func (vc *VoiceConnection) pingLoop() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-vc.pingDone:
			return
		case <-ticker.C:
			_ = vc.Send("ping", nil)
		}
	}
}

// Connect opens a bidirectional WebSocket connection for real-time voice call control.
func (s *CallsService) Connect(opts *VoiceConnectionOptions) (*VoiceConnection, error) {
	baseURL := s.client.baseURL
	wsURL := strings.Replace(baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)

	params := "token=" + s.client.apiKey
	if opts != nil && opts.AgentID != "" {
		params += "&agentId=" + opts.AgentID
	}
	wsURL += "/ws/voice?" + params

	return newVoiceConnection(wsURL)
}
