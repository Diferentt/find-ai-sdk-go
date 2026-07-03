package findai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServerClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient("fai_test_key", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithRetry(1, 0))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return c
}

func TestListTemplates(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/templates" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": "kt_1", "tenant_id": "t_1", "name": "Companies",
				"fields": [{"name": "company_name", "display_name": "Company Name", "type": "text", "required": true, "unique": false, "searchable": true, "filterable": true, "is_public": false, "is_llm_visible": true, "status": "active"}],
				"version": 1, "is_active": true,
				"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"
			}
		]`))
	})

	templates, err := c.ListTemplates(context.Background())
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("len(templates) = %d, want 1", len(templates))
	}
	tmpl := templates[0]
	if tmpl.ID != "kt_1" || tmpl.Name != "Companies" {
		t.Fatalf("unexpected template: %+v", tmpl)
	}
	if len(tmpl.Fields) != 1 || tmpl.Fields[0].Type != FieldTypeText {
		t.Fatalf("unexpected fields: %+v", tmpl.Fields)
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail": "Template not found"}`))
	})

	_, err := c.GetTemplate(context.Background(), "kt_missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true, err = %v", err)
	}
}

func TestGetTemplate(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/templates/kt_1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"id": "kt_1", "tenant_id": "t_1", "name": "Companies", "fields": [],
			"version": 2, "is_active": true,
			"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z"
		}`))
	})

	tmpl, err := c.GetTemplate(context.Background(), "kt_1")
	if err != nil {
		t.Fatalf("GetTemplate() error = %v", err)
	}
	if tmpl.Version != 2 {
		t.Fatalf("Version = %d, want 2", tmpl.Version)
	}
}
