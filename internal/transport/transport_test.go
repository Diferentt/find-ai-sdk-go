package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Client{
		BaseURL:    srv.URL,
		APIKey:     "fai_test",
		HTTPClient: srv.Client(),
		UserAgent:  "test-agent",
		Retry:      RetryPolicy{MaxAttempts: 3, BaseDelay: 5 * time.Millisecond},
	}, srv
}

func TestDo_SendsAuthAndDecodesResponse(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer fai_test" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "test-agent" {
			t.Errorf("User-Agent = %q", got)
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit query = %q", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"kr_1"}`))
	})

	var out struct {
		ID string `json:"id"`
	}
	q := url.Values{"limit": {"10"}}
	err := c.Do(context.Background(), "GET", "/records", q, nil, &out)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if out.ID != "kr_1" {
		t.Fatalf("ID = %q, want kr_1", out.ID)
	}
}

func TestDo_MarshalsRequestBody(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if body["values_data"] == nil {
			t.Errorf("expected values_data in request body, got %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"kr_2"}`))
	})

	var out struct {
		ID string `json:"id"`
	}
	reqBody := map[string]any{"values_data": map[string]any{"name": "Acme"}}
	if err := c.Do(context.Background(), "POST", "/records", nil, reqBody, &out); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}

func TestDo_NoContentResponse(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Do(context.Background(), "DELETE", "/records/kr_1", nil, nil, nil); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}

func TestDo_DecodesErrorBody(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail": "record not found"}`))
	})

	err := c.Do(context.Background(), "GET", "/records/missing", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := AsAPIError(err)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Fatalf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Detail != "record not found" {
		t.Fatalf("Detail = %q", apiErr.Detail)
	}
}

func TestDo_RetriesOn5xxThenSucceeds(t *testing.T) {
	var attempts int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"detail": "temporarily unavailable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"kr_3"}`))
	})

	var out struct {
		ID string `json:"id"`
	}
	if err := c.Do(context.Background(), "GET", "/records/kr_3", nil, nil, &out); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	if out.ID != "kr_3" {
		t.Fatalf("ID = %q", out.ID)
	}
}

func TestDo_DoesNotRetryOn404(t *testing.T) {
	var attempts int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail": "not found"}`))
	})

	err := c.Do(context.Background(), "GET", "/records/missing", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1 (404 should not be retried)", attempts)
	}
}

func TestDo_GivesUpAfterMaxAttempts(t *testing.T) {
	var attempts int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"detail": "boom"}`))
	})

	err := c.Do(context.Background(), "GET", "/records", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3 (MaxAttempts)", attempts)
	}
}

func TestDoMultipart_SendsFileAndDecodesResponse(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		f, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile: %v", err)
		}
		defer f.Close()
		if header.Filename != "data.csv" {
			t.Errorf("Filename = %q", header.Filename)
		}
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		if string(buf[:n]) != "a,b\n1,2\n" {
			t.Errorf("file content = %q", string(buf[:n]))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_rows": 1, "imported": 1, "errors": []}`))
	})

	var out struct {
		TotalRows int `json:"total_rows"`
		Imported  int `json:"imported"`
	}
	r := strings.NewReader("a,b\n1,2\n")
	if err := c.DoMultipart(context.Background(), "POST", "/import", "data.csv", r, &out); err != nil {
		t.Fatalf("DoMultipart() error = %v", err)
	}
	if out.Imported != 1 {
		t.Fatalf("Imported = %d, want 1", out.Imported)
	}
}
