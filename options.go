package findai

import (
	"net/http"
	"strings"
	"time"
)

// Option configures a Client during construction.
type Option func(*Client)

// WithBaseURL sets the API host to send requests to, e.g.
// "https://api.example.com". Required — there is no default.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.t.BaseURL = strings.TrimRight(url, "/")
	}
}

// WithHTTPClient overrides the underlying *http.Client, e.g. to configure a
// custom transport, proxy, or TLS settings. Note this replaces any timeout
// set via WithTimeout if applied after it.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.t.HTTPClient = hc
	}
}

// WithTimeout sets the per-request timeout on the client's HTTP client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.t.HTTPClient.Timeout = d
	}
}

// WithRetry configures retry behavior for transient failures (network
// errors, HTTP 429, and 5xx responses). maxAttempts is the total number of
// attempts including the first; maxAttempts=1 disables retrying.
func WithRetry(maxAttempts int, baseDelay time.Duration) Option {
	return func(c *Client) {
		c.t.Retry.MaxAttempts = maxAttempts
		c.t.Retry.BaseDelay = baseDelay
	}
}

// WithUserAgent overrides the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.t.UserAgent = ua
	}
}
