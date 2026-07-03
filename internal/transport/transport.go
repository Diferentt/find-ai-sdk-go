package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RetryPolicy controls how the client retries failed requests.
type RetryPolicy struct {
	// MaxAttempts is the total number of attempts (including the first),
	// so MaxAttempts=1 disables retrying.
	MaxAttempts int
	// BaseDelay is the delay before the first retry; subsequent retries
	// back off exponentially with jitter.
	BaseDelay time.Duration
}

// Client is the shared HTTP engine used by the public findai.Client. It
// knows nothing about specific resources (templates, records, ...) — only
// how to build, sign, send, and decode requests against the API.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
	Retry      RetryPolicy
}

// Do issues a JSON request. body is marshaled as the request body when
// non-nil; out is populated by unmarshaling the response body when non-nil.
// A nil out is valid for responses with no body (e.g. 204 No Content).
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("findai: encoding request body: %w", err)
		}
	}

	return c.doWithRetry(ctx, func() (*http.Request, error) {
		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(payload)
		}
		req, err := c.newRequest(ctx, method, path, query, bodyReader)
		if err != nil {
			return nil, err
		}
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return req, nil
	}, out)
}

// DoMultipart issues a multipart/form-data request with a single file part
// named "file". The full content of r is buffered before sending so the
// request can be safely retried on transient failures.
func (c *Client) DoMultipart(ctx context.Context, method, path, filename string, r io.Reader, out any) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("findai: reading file content: %w", err)
	}

	return c.doWithRetry(ctx, func() (*http.Request, error) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, err := w.CreateFormFile("file", filename)
		if err != nil {
			return nil, fmt.Errorf("findai: building multipart body: %w", err)
		}
		if _, err := part.Write(content); err != nil {
			return nil, fmt.Errorf("findai: building multipart body: %w", err)
		}
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("findai: building multipart body: %w", err)
		}

		req, err := c.newRequest(ctx, method, path, nil, &buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
		return req, nil
	}, out)
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Request, error) {
	u := strings.TrimRight(c.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, fmt.Errorf("findai: building request: %w", err)
	}
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// doWithRetry sends the request built by buildReq, retrying on transient
// network errors and 429/5xx responses per c.Retry. buildReq is invoked
// once per attempt so the body reader is always fresh.
func (c *Client) doWithRetry(ctx context.Context, buildReq func() (*http.Request, error), out any) error {
	maxAttempts := c.Retry.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := backoffDelay(c.Retry.BaseDelay, attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := buildReq()
		if err != nil {
			return err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("findai: request failed: %w", err)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("findai: reading response body: %w", readErr)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil || len(respBody) == 0 {
				return nil
			}
			if err := json.Unmarshal(respBody, out); err != nil {
				return fmt.Errorf("findai: decoding response body: %w", err)
			}
			return nil
		}

		apiErr := decodeAPIError(resp.StatusCode, respBody, resp.Header.Get("X-Request-Id"))
		if !isRetryableStatus(resp.StatusCode) || attempt == maxAttempts-1 {
			return apiErr
		}
		lastErr = apiErr
	}

	return lastErr
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

func backoffDelay(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		base = 200 * time.Millisecond
	}
	// attempt is 1-indexed here (called with attempt >= 1): 1x, 2x, 4x, ... base, plus jitter.
	factor := 1 << (attempt - 1)
	delay := base * time.Duration(factor)
	jitter := time.Duration(rand.Int63n(int64(delay)/2 + 1)) //nolint:gosec // jitter does not need to be cryptographically secure
	return delay + jitter
}

// AsAPIError is a convenience helper mirroring errors.As for callers within
// this package's tests.
func AsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
