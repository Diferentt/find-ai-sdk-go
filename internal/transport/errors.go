// Package transport implements the low-level HTTP plumbing shared by every
// find-ai-sdk-go resource method: request construction, auth, retries, and
// response/error decoding. It is not part of the public API.
package transport

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FieldError is one entry of a FastAPI/Pydantic body-validation error, as
// returned when the request body fails schema validation before it ever
// reaches application code.
type FieldError struct {
	Loc  []string `json:"loc"`
	Msg  string   `json:"msg"`
	Type string   `json:"type"`
}

// APIError represents a non-2xx response from the API.
//
// The backend does not use a single uniform error shape: hand-raised
// business errors return `{"detail": "<string>"}`, while FastAPI's own
// request-body validation errors return `{"detail": [{"loc":...,"msg":...,"type":...}]}`.
// APIError normalizes both into a single Detail string, while preserving the
// structured field errors (when present) in Errors.
type APIError struct {
	StatusCode int
	Detail     string
	Errors     []FieldError
	RequestID  string
	rawBody    []byte
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("findai: %d %s (request_id=%s)", e.StatusCode, e.Detail, e.RequestID)
	}
	return fmt.Sprintf("findai: %d %s", e.StatusCode, e.Detail)
}

// RawBody returns the unparsed response body, for callers that need to
// inspect a shape the SDK doesn't otherwise model.
func (e *APIError) RawBody() []byte { return e.rawBody }

// decodeAPIError builds an APIError from a non-2xx HTTP response body.
func decodeAPIError(statusCode int, body []byte, requestID string) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		RequestID:  requestID,
		rawBody:    body,
	}

	var stringDetail struct {
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(body, &stringDetail); err == nil && stringDetail.Detail != "" {
		apiErr.Detail = stringDetail.Detail
		return apiErr
	}

	var arrayDetail struct {
		Detail []FieldError `json:"detail"`
	}
	if err := json.Unmarshal(body, &arrayDetail); err == nil && len(arrayDetail.Detail) > 0 {
		apiErr.Errors = arrayDetail.Detail
		parts := make([]string, 0, len(arrayDetail.Detail))
		for _, fe := range arrayDetail.Detail {
			loc := strings.Join(fe.Loc, ".")
			if loc != "" {
				parts = append(parts, fmt.Sprintf("%s: %s", loc, fe.Msg))
			} else {
				parts = append(parts, fe.Msg)
			}
		}
		apiErr.Detail = "validation failed: " + strings.Join(parts, "; ")
		return apiErr
	}

	if len(body) == 0 {
		apiErr.Detail = "no response body"
		return apiErr
	}
	apiErr.Detail = string(body)
	return apiErr
}
