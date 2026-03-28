package anima

import (
	"errors"
	"fmt"
)

// Sentinel errors for use with errors.Is.
var (
	ErrAuthentication  = errors.New("authentication failed")
	ErrNotFound        = errors.New("resource not found")
	ErrValidation      = errors.New("validation failed")
	ErrRateLimit       = errors.New("rate limit exceeded")
	ErrConflict        = errors.New("resource conflict")
	ErrInternalServer  = errors.New("internal server error")
	ErrTimeout         = errors.New("request timed out")
	ErrNetwork         = errors.New("network error")
	ErrRetryExhausted  = errors.New("request failed after retries")
	ErrWebhookInvalid  = errors.New("invalid webhook signature")
	ErrWebhookExpired  = errors.New("webhook timestamp expired")
)

// APIError represents an error returned by the Anima API. It wraps a sentinel
// error so that callers can use errors.Is to check the category, and
// errors.As to extract detailed information.
type APIError struct {
	// Status is the HTTP status code (0 for network-level errors).
	Status int `json:"status"`
	// Code is a machine-readable error code from the API (e.g. "RATE_LIMIT").
	Code string `json:"code,omitempty"`
	// Message is a human-readable description.
	Message string `json:"message"`
	// Details contains optional structured error information.
	Details any `json:"details,omitempty"`
	// RetryAfter is set on 429 responses if the server provided a Retry-After header (seconds).
	RetryAfter int `json:"retry_after,omitempty"`

	// sentinel is the underlying sentinel error that this wraps.
	sentinel error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("anima: %s (status=%d, code=%s)", e.Message, e.Status, e.Code)
	}
	return fmt.Sprintf("anima: %s (status=%d)", e.Message, e.Status)
}

// Unwrap returns the sentinel error so errors.Is works.
func (e *APIError) Unwrap() error {
	return e.sentinel
}

// newAPIError creates an APIError wrapping the given sentinel.
func newAPIError(sentinel error, status int, code, message string, details any) *APIError {
	return &APIError{
		Status:   status,
		Code:     code,
		Message:  message,
		Details:  details,
		sentinel: sentinel,
	}
}

// Typed constructors for specific error categories.

func newAuthError(message string, details any) *APIError {
	return newAPIError(ErrAuthentication, 401, "AUTH_ERROR", message, details)
}

func newNotFoundError(message string, details any) *APIError {
	return newAPIError(ErrNotFound, 404, "NOT_FOUND", message, details)
}

func newValidationError(message string, details any) *APIError {
	return newAPIError(ErrValidation, 400, "VALIDATION_ERROR", message, details)
}

func newRateLimitError(message string, retryAfter int, details any) *APIError {
	e := newAPIError(ErrRateLimit, 429, "RATE_LIMIT", message, details)
	e.RetryAfter = retryAfter
	return e
}

func newConflictError(message string, details any) *APIError {
	return newAPIError(ErrConflict, 409, "CONFLICT", message, details)
}

func newInternalServerError(message string, status int, details any) *APIError {
	return newAPIError(ErrInternalServer, status, "INTERNAL_ERROR", message, details)
}

func newTimeoutError(message string) *APIError {
	return newAPIError(ErrTimeout, 408, "TIMEOUT", message, nil)
}

func newNetworkError(message string) *APIError {
	return newAPIError(ErrNetwork, 0, "NETWORK_ERROR", message, nil)
}

func newRetryExhaustedError() *APIError {
	return newAPIError(ErrRetryExhausted, 0, "RETRY_EXHAUSTED", "request failed after retries", nil)
}
