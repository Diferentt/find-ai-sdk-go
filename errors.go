package findai

import (
	"errors"

	"github.com/Diferentt/find-ai-sdk-go/internal/transport"
)

// APIError represents a non-2xx response from the API. The backend does not
// use a single uniform error shape: hand-raised business errors return
// {"detail": "<string>"}, while request-body validation errors return
// {"detail": [{"loc":...,"msg":...,"type":...}]}. APIError normalizes both
// into Detail, while preserving structured field errors (when present) in
// Errors.
type APIError = transport.APIError

// FieldError is one entry of a request-body validation error.
type FieldError = transport.FieldError

// ErrInvalidAPIKey is returned by NewClient when the supplied API key does
// not look like a valid find-ai key (client-side format check only — the
// server is always the source of truth for whether a key is actually valid).
var ErrInvalidAPIKey = errors.New("findai: invalid API key")

// IsNotFound reports whether err is an APIError with status 404.
func IsNotFound(err error) bool { return hasStatus(err, 404) }

// IsUnauthorized reports whether err is an APIError with status 401.
func IsUnauthorized(err error) bool { return hasStatus(err, 401) }

// IsForbidden reports whether err is an APIError with status 403.
func IsForbidden(err error) bool { return hasStatus(err, 403) }

// IsValidationError reports whether err is an APIError with status 422.
func IsValidationError(err error) bool { return hasStatus(err, 422) }

// IsRateLimited reports whether err is an APIError with status 429.
func IsRateLimited(err error) bool { return hasStatus(err, 429) }

func hasStatus(err error, status int) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == status
	}
	return false
}
