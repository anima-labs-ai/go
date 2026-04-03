package anima

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Call represents a voice call.
type Call struct {
	ID              string  `json:"id"`
	AgentID         string  `json:"agentId"`
	PhoneIdentityID string  `json:"phoneIdentityId"`
	Direction       string  `json:"direction"`
	Tier            string  `json:"tier"`
	State           string  `json:"state"`
	From            string  `json:"from"`
	To              string  `json:"to"`
	StartedAt       string  `json:"startedAt"`
	AnsweredAt      *string `json:"answeredAt"`
	EndedAt         *string `json:"endedAt"`
	EndReason       *string `json:"endReason"`
	DurationSeconds *int    `json:"durationSeconds"`
	CreatedAt       string  `json:"createdAt"`
}

// CallList wraps a list of calls with total count.
type CallList struct {
	Calls []Call `json:"calls"`
	Total int    `json:"total"`
}

// ListCallsParams contains parameters for listing calls.
type ListCallsParams struct {
	AgentID   string
	Direction string
	State     string
	Limit     int
	Offset    int
}

// CreateCallParams contains parameters for creating an outbound call.
type CreateCallParams struct {
	To         string `json:"to"`
	AgentID    string `json:"agentId,omitempty"`
	Tier       string `json:"tier,omitempty"`
	Greeting   string `json:"greeting,omitempty"`
	FromNumber string `json:"fromNumber,omitempty"`
}

// CreateCallOutput contains the result of creating a call.
type CreateCallOutput struct {
	CallID    string `json:"callId"`
	State     string `json:"state"`
	From      string `json:"from"`
	To        string `json:"to"`
	Tier      string `json:"tier"`
	Direction string `json:"direction"`
}

// TranscriptSegment represents a segment of a call transcript.
type TranscriptSegment struct {
	Speaker    string  `json:"speaker"`
	Text       string  `json:"text"`
	StartTime  float64 `json:"startTime"`
	EndTime    float64 `json:"endTime"`
	Confidence float64 `json:"confidence"`
	IsFinal    bool    `json:"isFinal"`
}

// CallTranscript contains the full transcript for a call.
type CallTranscript struct {
	CallID   string             `json:"callId"`
	Segments []TranscriptSegment `json:"segments"`
}

// CallsService provides methods for managing voice calls.
type CallsService struct {
	client *httpClient
}

// newCallsService creates a new CallsService.
func newCallsService(c *httpClient) *CallsService {
	return &CallsService{client: c}
}

// List returns a list of calls, optionally filtered.
func (s *CallsService) List(ctx context.Context, params ListCallsParams) (*CallList, error) {
	q := url.Values{}
	if params.AgentID != "" {
		q.Set("agentId", params.AgentID)
	}
	if params.Direction != "" {
		q.Set("direction", params.Direction)
	}
	if params.State != "" {
		q.Set("state", params.State)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}
	list, err := Do[CallList](ctx, s.client, http.MethodGet, "/voice/calls", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Get returns a specific call by ID.
func (s *CallsService) Get(ctx context.Context, callID string) (*Call, error) {
	call, err := Do[Call](ctx, s.client, http.MethodGet, "/voice/calls/"+callID, nil, nil)
	if err != nil {
		return nil, err
	}
	return &call, nil
}

// Create creates a new outbound call.
func (s *CallsService) Create(ctx context.Context, params CreateCallParams) (*CreateCallOutput, error) {
	out, err := Do[CreateCallOutput](ctx, s.client, http.MethodPost, "/voice/calls", params, nil)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTranscript returns the transcript for a call.
func (s *CallsService) GetTranscript(ctx context.Context, callID string) (*CallTranscript, error) {
	transcript, err := Do[CallTranscript](ctx, s.client, http.MethodGet, "/voice/calls/"+callID+"/transcript", nil, nil)
	if err != nil {
		return nil, err
	}
	return &transcript, nil
}
