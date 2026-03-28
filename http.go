package anima

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// retryDelays are the base delays for exponential backoff.
var retryDelays = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	4 * time.Second,
}

// httpClient is the low-level HTTP transport for the Anima API.
type httpClient struct {
	apiKey     string
	baseURL    string
	maxRetries int
	client     *http.Client
}

// apiErrorEnvelope matches the JSON error shape returned by the Anima API.
type apiErrorEnvelope struct {
	Message string `json:"message,omitempty"`
	Error   *struct {
		Message string `json:"message,omitempty"`
		Code    string `json:"code,omitempty"`
		Details any    `json:"details,omitempty"`
	} `json:"error,omitempty"`
}

// RequestOption allows per-request customization.
type RequestOption func(*http.Request)

// Do executes an HTTP request with retries and returns the parsed response body.
// T is the type to unmarshal the response body into.
func Do[T any](ctx context.Context, hc *httpClient, method, path string, body any, query url.Values) (T, error) {
	var zero T

	reqURL, err := hc.buildURL(path, query)
	if err != nil {
		return zero, newNetworkError(fmt.Sprintf("invalid URL: %v", err))
	}

	var bodyReader func() (io.Reader, error)
	if body != nil {
		encoded, encErr := json.Marshal(body)
		if encErr != nil {
			return zero, newValidationError(fmt.Sprintf("failed to encode request body: %v", encErr), nil)
		}
		bodyReader = func() (io.Reader, error) {
			return bytes.NewReader(encoded), nil
		}
	} else {
		bodyReader = func() (io.Reader, error) {
			return nil, nil
		}
	}

	for attempt := 0; attempt <= hc.maxRetries; attempt++ {
		r, reqErr := bodyReader()
		if reqErr != nil {
			return zero, newNetworkError(fmt.Sprintf("failed to prepare request body: %v", reqErr))
		}

		req, reqErr := http.NewRequestWithContext(ctx, method, reqURL, r)
		if reqErr != nil {
			return zero, newNetworkError(fmt.Sprintf("failed to create request: %v", reqErr))
		}

		req.Header.Set("Authorization", "Bearer "+hc.apiKey)
		req.Header.Set("User-Agent", "anima-go/"+SDKVersion)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, doErr := hc.client.Do(req)
		if doErr != nil {
			// Context cancellation — do not retry.
			if ctx.Err() != nil {
				return zero, newTimeoutError(fmt.Sprintf("request cancelled: %v", ctx.Err()))
			}

			if attempt < hc.maxRetries {
				hc.sleep(ctx, hc.backoff(attempt, nil))
				continue
			}
			return zero, newNetworkError(doErr.Error())
		}

		// Success.
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			defer resp.Body.Close()
			if resp.StatusCode == 204 {
				return zero, nil
			}

			var result T
			if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
				return zero, newNetworkError(fmt.Sprintf("failed to decode response: %v", decErr))
			}
			return result, nil
		}

		// Retryable status codes.
		if hc.shouldRetry(resp.StatusCode) && attempt < hc.maxRetries {
			delay := hc.backoff(attempt, resp)
			resp.Body.Close()
			hc.sleep(ctx, delay)
			continue
		}

		// Non-retryable error — parse and return.
		apiErr := hc.parseError(resp)
		resp.Body.Close()
		return zero, apiErr
	}

	return zero, newRetryExhaustedError()
}

// shouldRetry returns true for status codes that warrant a retry.
func (hc *httpClient) shouldRetry(status int) bool {
	return status == 429 || status >= 500
}

// backoff calculates the delay for the given attempt. If the response includes
// a Retry-After header, that value is used instead.
func (hc *httpClient) backoff(attempt int, resp *http.Response) time.Duration {
	// Check Retry-After header.
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if seconds, err := strconv.Atoi(ra); err == nil && seconds > 0 {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	// Exponential backoff with capped delay table.
	idx := attempt
	if idx >= len(retryDelays) {
		idx = len(retryDelays) - 1
	}
	base := retryDelays[idx]

	// Add jitter: 0.5x to 1.5x.
	jitter := 0.5 + (float64(attempt%7) / 7.0)
	return time.Duration(math.Round(float64(base) * jitter))
}

// sleep pauses for the given duration, respecting context cancellation.
func (hc *httpClient) sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// buildURL constructs the full request URL.
func (hc *httpClient) buildURL(path string, query url.Values) (string, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u, err := url.Parse(hc.baseURL + path)
	if err != nil {
		return "", err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String(), nil
}

// parseError reads the response body and returns a typed APIError.
func (hc *httpClient) parseError(resp *http.Response) *APIError {
	retryAfterHeader := resp.Header.Get("Retry-After")
	retryAfter := 0
	if retryAfterHeader != "" {
		if v, err := strconv.Atoi(retryAfterHeader); err == nil {
			retryAfter = v
		}
	}

	var envelope apiErrorEnvelope
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &envelope)

	message := fmt.Sprintf("request failed with status %d", resp.StatusCode)
	var code string
	var details any

	if envelope.Error != nil {
		if envelope.Error.Message != "" {
			message = envelope.Error.Message
		}
		code = envelope.Error.Code
		details = envelope.Error.Details
	} else if envelope.Message != "" {
		message = envelope.Message
	}

	switch resp.StatusCode {
	case 400, 422:
		return newValidationError(message, details)
	case 401, 403:
		return newAuthError(message, details)
	case 404:
		return newNotFoundError(message, details)
	case 409:
		return newConflictError(message, details)
	case 429:
		return newRateLimitError(message, retryAfter, details)
	default:
		if resp.StatusCode >= 500 {
			return newInternalServerError(message, resp.StatusCode, details)
		}
		return newAPIError(nil, resp.StatusCode, code, message, details)
	}
}
