package transport

import "testing"

func TestDecodeAPIError_StringDetail(t *testing.T) {
	body := []byte(`{"detail": "template not found"}`)
	err := decodeAPIError(404, body, "req-123")

	if err.StatusCode != 404 {
		t.Fatalf("StatusCode = %d, want 404", err.StatusCode)
	}
	if err.Detail != "template not found" {
		t.Fatalf("Detail = %q, want %q", err.Detail, "template not found")
	}
	if err.RequestID != "req-123" {
		t.Fatalf("RequestID = %q, want %q", err.RequestID, "req-123")
	}
	if len(err.Errors) != 0 {
		t.Fatalf("Errors = %v, want empty", err.Errors)
	}
}

func TestDecodeAPIError_ArrayDetail(t *testing.T) {
	body := []byte(`{"detail": [{"loc": ["body", "values_data", "email"], "msg": "value is not a valid email address", "type": "value_error"}]}`)
	err := decodeAPIError(422, body, "")

	if err.StatusCode != 422 {
		t.Fatalf("StatusCode = %d, want 422", err.StatusCode)
	}
	if len(err.Errors) != 1 {
		t.Fatalf("Errors = %v, want 1 entry", err.Errors)
	}
	fe := err.Errors[0]
	if fe.Msg != "value is not a valid email address" {
		t.Fatalf("Errors[0].Msg = %q", fe.Msg)
	}
	want := "validation failed: body.values_data.email: value is not a valid email address"
	if err.Detail != want {
		t.Fatalf("Detail = %q, want %q", err.Detail, want)
	}
}

func TestDecodeAPIError_UnknownShape(t *testing.T) {
	body := []byte(`internal server error, upstream timeout`)
	err := decodeAPIError(502, body, "")

	if err.StatusCode != 502 {
		t.Fatalf("StatusCode = %d, want 502", err.StatusCode)
	}
	if err.Detail != string(body) {
		t.Fatalf("Detail = %q, want raw body %q", err.Detail, string(body))
	}
}

func TestDecodeAPIError_EmptyBody(t *testing.T) {
	err := decodeAPIError(500, nil, "")
	if err.Detail == "" {
		t.Fatalf("Detail should not be empty for an empty body")
	}
}

func TestAPIError_ErrorMessage(t *testing.T) {
	err := decodeAPIError(404, []byte(`{"detail": "not found"}`), "req-1")
	got := err.Error()
	want := "findai: 404 not found (request_id=req-1)"
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
