package anima

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := newAPIError(ErrAuthentication, 401, "AUTH_ERROR", "bad token", nil)
	want := "anima: bad token (status=401, code=AUTH_ERROR)"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestAPIError_ErrorWithoutCode(t *testing.T) {
	err := &APIError{Status: 500, Message: "oops"}
	want := "anima: oops (status=500)"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestAPIError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		target   error
		wantIs   bool
	}{
		{"auth error matches ErrAuthentication", newAuthError("bad", nil), ErrAuthentication, true},
		{"auth error does not match ErrNotFound", newAuthError("bad", nil), ErrNotFound, false},
		{"not found matches ErrNotFound", newNotFoundError("gone", nil), ErrNotFound, true},
		{"validation matches ErrValidation", newValidationError("bad input", nil), ErrValidation, true},
		{"rate limit matches ErrRateLimit", newRateLimitError("slow down", 30, nil), ErrRateLimit, true},
		{"conflict matches ErrConflict", newConflictError("dup", nil), ErrConflict, true},
		{"internal server matches ErrInternalServer", newInternalServerError("boom", 502, nil), ErrInternalServer, true},
		{"timeout matches ErrTimeout", newTimeoutError("timed out"), ErrTimeout, true},
		{"network matches ErrNetwork", newNetworkError("dns fail"), ErrNetwork, true},
		{"retry exhausted matches ErrRetryExhausted", newRetryExhaustedError(), ErrRetryExhausted, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.wantIs {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.err, tt.target, got, tt.wantIs)
			}
		})
	}
}

func TestAPIError_As(t *testing.T) {
	err := newRateLimitError("slow down", 60, map[string]string{"reason": "too many"})

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("errors.As should succeed for *APIError")
	}

	if apiErr.Status != 429 {
		t.Errorf("Status = %d, want 429", apiErr.Status)
	}
	if apiErr.Code != "RATE_LIMIT" {
		t.Errorf("Code = %q, want RATE_LIMIT", apiErr.Code)
	}
	if apiErr.RetryAfter != 60 {
		t.Errorf("RetryAfter = %d, want 60", apiErr.RetryAfter)
	}
	if apiErr.Details == nil {
		t.Error("Details should not be nil")
	}
}

func TestAPIError_As_WrappedInStandardError(t *testing.T) {
	original := newNotFoundError("agent not found", nil)
	wrapped := fmt.Errorf("operation failed: %w", original)

	var apiErr *APIError
	if !errors.As(wrapped, &apiErr) {
		t.Fatal("errors.As should unwrap through fmt.Errorf wrapping")
	}
	if apiErr.Status != 404 {
		t.Errorf("Status = %d, want 404", apiErr.Status)
	}
}
