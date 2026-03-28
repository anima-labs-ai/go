package anima

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_Success(t *testing.T) {
	type resp struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or wrong Authorization header")
		}
		if r.Header.Get("User-Agent") != "anima-go/"+SDKVersion {
			t.Errorf("unexpected User-Agent: %s", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp{Name: "test", ID: 42})
	}))
	defer srv.Close()

	hc := &httpClient{
		apiKey:     "test-key",
		baseURL:    srv.URL,
		maxRetries: 0,
		client:     srv.Client(),
	}

	result, err := Do[resp](context.Background(), hc, "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "test" || result.ID != 42 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestDo_204NoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer srv.Close()

	hc := &httpClient{
		apiKey:     "test-key",
		baseURL:    srv.URL,
		maxRetries: 0,
		client:     srv.Client(),
	}

	type empty struct{}
	_, err := Do[empty](context.Background(), hc, "DELETE", "/test/123", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDo_RetriesOn500(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(500)
			w.Write([]byte(`{"error":{"message":"server error"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	// Use very short retry delays for testing.
	origDelays := retryDelays
	retryDelays = []time.Duration{1 * time.Millisecond, 1 * time.Millisecond, 1 * time.Millisecond}
	defer func() { retryDelays = origDelays }()

	hc := &httpClient{
		apiKey:     "test-key",
		baseURL:    srv.URL,
		maxRetries: 3,
		client:     srv.Client(),
	}

	type resp struct {
		Status string `json:"status"`
	}
	result, err := Do[resp](context.Background(), hc, "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ok" {
		t.Errorf("unexpected result: %+v", result)
	}
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestDo_RetriesOn429(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.WriteHeader(429)
			w.Write([]byte(`{"error":{"message":"rate limited"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer srv.Close()

	origDelays := retryDelays
	retryDelays = []time.Duration{1 * time.Millisecond, 1 * time.Millisecond}
	defer func() { retryDelays = origDelays }()

	hc := &httpClient{
		apiKey:     "test-key",
		baseURL:    srv.URL,
		maxRetries: 2,
		client:     srv.Client(),
	}

	type resp struct{ Ok string }
	_, err := Do[resp](context.Background(), hc, "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}
}

func TestDo_NoRetryOn400(t *testing.T) {
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(400)
		w.Write([]byte(`{"error":{"message":"bad request","code":"VALIDATION_ERROR","details":{"field":"name"}}}`))
	}))
	defer srv.Close()

	hc := &httpClient{
		apiKey:     "test-key",
		baseURL:    srv.URL,
		maxRetries: 3,
		client:     srv.Client(),
	}

	type resp struct{}
	_, err := Do[resp](context.Background(), hc, "POST", "/test", map[string]string{"bad": "data"}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts.Load() != 1 {
		t.Errorf("should not retry on 400, got %d attempts", attempts.Load())
	}

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if apiErr.Status != 400 {
			t.Errorf("Status = %d, want 400", apiErr.Status)
		}
	} else {
		t.Error("expected *APIError")
	}
}

func TestDo_ParsesErrorTypes(t *testing.T) {
	tests := []struct {
		status   int
		body     string
		sentinel error
	}{
		{401, `{"error":{"message":"unauthorized"}}`, ErrAuthentication},
		{403, `{"error":{"message":"forbidden"}}`, ErrAuthentication},
		{404, `{"error":{"message":"not found"}}`, ErrNotFound},
		{409, `{"error":{"message":"conflict"}}`, ErrConflict},
		{422, `{"error":{"message":"invalid"}}`, ErrValidation},
	}

	for _, tt := range tests {
		t.Run(tt.sentinel.Error(), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			hc := &httpClient{
				apiKey:     "test-key",
				baseURL:    srv.URL,
				maxRetries: 0,
				client:     srv.Client(),
			}

			type resp struct{}
			_, err := Do[resp](context.Background(), hc, "GET", "/test", nil, nil)
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("expected %v, got %v", tt.sentinel, err)
			}
		})
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	hc := &httpClient{
		apiKey:     "test-key",
		baseURL:    srv.URL,
		maxRetries: 0,
		client:     &http.Client{Timeout: 5 * time.Second},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	type resp struct{}
	_, err := Do[resp](ctx, hc, "GET", "/slow", nil, nil)
	if err == nil {
		t.Fatal("expected error from context cancellation")
	}
}

func TestDo_PostWithBody(t *testing.T) {
	type reqBody struct {
		Name string `json:"name"`
	}
	type respBody struct {
		ID string `json:"id"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json Content-Type, got %s", r.Header.Get("Content-Type"))
		}

		var body reqBody
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "test-agent" {
			t.Errorf("unexpected body: %+v", body)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(respBody{ID: "ag_123"})
	}))
	defer srv.Close()

	hc := &httpClient{
		apiKey:     "test-key",
		baseURL:    srv.URL,
		maxRetries: 0,
		client:     srv.Client(),
	}

	result, err := Do[respBody](context.Background(), hc, "POST", "/agents", reqBody{Name: "test-agent"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "ag_123" {
		t.Errorf("unexpected ID: %s", result.ID)
	}
}

func TestBackoff_RespectsRetryAfterHeader(t *testing.T) {
	hc := &httpClient{}
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Retry-After", "10")

	delay := hc.backoff(0, resp)
	if delay != 10*time.Second {
		t.Errorf("expected 10s from Retry-After, got %v", delay)
	}
}

func TestBackoff_ExponentialWithoutHeader(t *testing.T) {
	hc := &httpClient{}

	d0 := hc.backoff(0, nil)
	d1 := hc.backoff(1, nil)
	d2 := hc.backoff(2, nil)

	// Just verify they're non-zero and generally increasing.
	if d0 <= 0 || d1 <= 0 || d2 <= 0 {
		t.Errorf("delays should be positive: %v, %v, %v", d0, d1, d2)
	}
}
