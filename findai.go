// Package findai is a Go client for the FindAI Studio "datasets" (knowledge
// module) API: list your dataset tables, inspect their structure, and
// create/read/update/delete rows, plus full-text search, semantic search,
// and CSV import.
package findai

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Diferentt/find-ai-sdk-go/internal/transport"
)

const (
	apiKeyPrefix          = "fai_"
	defaultUserAgent      = "find-ai-sdk-go"
	defaultTimeout        = 30 * time.Second
	defaultRetryAttempts  = 3
	defaultRetryBaseDelay = 200 * time.Millisecond
)

// Client is a FindAI Studio API client. Construct one with NewClient.
type Client struct {
	t *transport.Client
}

// NewClient builds a Client authenticated with apiKey (a tenant-scoped key
// with the "dataset:manage" scope, formatted "fai_...").
//
// A base URL must be supplied via WithBaseURL — the SDK does not assume a
// default API host.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("findai: API key is required")
	}
	if !strings.HasPrefix(apiKey, apiKeyPrefix) {
		return nil, fmt.Errorf("findai: API key must start with %q: %w", apiKeyPrefix, ErrInvalidAPIKey)
	}

	c := &Client{
		t: &transport.Client{
			APIKey:     apiKey,
			HTTPClient: &http.Client{Timeout: defaultTimeout},
			UserAgent:  defaultUserAgent,
			Retry: transport.RetryPolicy{
				MaxAttempts: defaultRetryAttempts,
				BaseDelay:   defaultRetryBaseDelay,
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.t.BaseURL == "" {
		return nil, fmt.Errorf("findai: base URL is required; pass findai.WithBaseURL(...)")
	}

	return c, nil
}
