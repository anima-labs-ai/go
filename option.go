package anima

import (
	"net/http"
	"time"
)

// Option configures the Anima client. Use the With* functions to create options.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL    string
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
}

func defaultConfig() clientConfig {
	return clientConfig{
		baseURL:    DefaultBaseURL,
		timeout:    DefaultTimeout,
		maxRetries: DefaultMaxRetries,
	}
}

// WithBaseURL sets a custom API base URL (e.g. for staging or self-hosted).
func WithBaseURL(url string) Option {
	return func(c *clientConfig) {
		c.baseURL = url
	}
}

// WithTimeout sets the per-request timeout. Default is 30 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) {
		c.timeout = d
	}
}

// WithMaxRetries sets the maximum number of retries for failed requests.
// Default is 3. Set to 0 to disable retries.
func WithMaxRetries(n int) Option {
	return func(c *clientConfig) {
		c.maxRetries = n
	}
}

// WithHTTPClient sets a custom *http.Client. When set, the client's Timeout
// field takes precedence over WithTimeout.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = hc
	}
}
