package findai

import (
	"context"
	"net/http"
	"testing"
)

func TestGetLimits(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/limits" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"knowledge_max_templates": 10,
			"knowledge_max_fields_per_template": 50,
			"knowledge_max_records_per_template": 100000,
			"knowledge_max_searchable_fields": 5
		}`))
	})

	limits, err := c.GetLimits(context.Background())
	if err != nil {
		t.Fatalf("GetLimits() error = %v", err)
	}
	if limits.MaxTemplates != 10 || limits.MaxRecordsPerTemplate != 100000 {
		t.Fatalf("unexpected limits: %+v", limits)
	}
}
