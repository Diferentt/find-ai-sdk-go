package findai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateRecord(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/templates/kt_1/records" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		vd, ok := body["values_data"].(map[string]any)
		if !ok || vd["company_name"] != "Acme" {
			t.Errorf("unexpected body: %+v", body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "kr_1", "template_id": "kt_1", "tenant_id": "t_1", "template_version": 1,
			"values_data": {"company_name": "Acme"}, "is_active": true,
			"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"
		}`))
	})

	rec, err := c.CreateRecord(context.Background(), "kt_1", map[string]any{"company_name": "Acme"})
	if err != nil {
		t.Fatalf("CreateRecord() error = %v", err)
	}
	if rec.ID != "kr_1" || rec.ValuesData["company_name"] != "Acme" {
		t.Fatalf("unexpected record: %+v", rec)
	}
}

func TestUpdateAndDeleteRecord(t *testing.T) {
	var deleteCalled bool
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			if r.URL.Path != "/api/v1/knowledge/templates/kt_1/records/kr_1" {
				t.Errorf("path = %q", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{
				"id": "kr_1", "template_id": "kt_1", "tenant_id": "t_1", "template_version": 1,
				"values_data": {"company_name": "Acme Corp"}, "is_active": true,
				"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z"
			}`))
		case http.MethodDelete:
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected method %q", r.Method)
		}
	})

	rec, err := c.UpdateRecord(context.Background(), "kt_1", "kr_1", map[string]any{"company_name": "Acme Corp"})
	if err != nil {
		t.Fatalf("UpdateRecord() error = %v", err)
	}
	if rec.ValuesData["company_name"] != "Acme Corp" {
		t.Fatalf("unexpected record: %+v", rec)
	}

	if err := c.DeleteRecord(context.Background(), "kt_1", "kr_1"); err != nil {
		t.Fatalf("DeleteRecord() error = %v", err)
	}
	if !deleteCalled {
		t.Fatal("expected DELETE to be called")
	}
}

func TestListRecords_ClampsLimit(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "100" {
			t.Errorf("limit = %q, want 100 (clamped)", got)
		}
		if got := r.URL.Query().Get("offset"); got != "0" {
			t.Errorf("offset = %q, want 0", got)
		}
		_, _ = w.Write([]byte(`{"records": [], "total": 0, "offset": 0, "limit": 100, "has_more": false}`))
	})

	page, err := c.ListRecords(context.Background(), "kt_1", ListRecordsOptions{Limit: 500})
	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}
	if len(page.Records) != 0 {
		t.Fatalf("expected empty page, got %+v", page)
	}
}

func TestListRecordsIterator_WalksAllPages(t *testing.T) {
	// 3 total records across pages of size 2: [0,1], [2], done.
	pages := [][]string{{"kr_1", "kr_2"}, {"kr_3"}}
	var callCount int

	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if callCount >= len(pages) {
			t.Fatalf("unexpected extra page fetch (call #%d)", callCount+1)
		}
		ids := pages[callCount]
		callCount++

		recs := make([]string, 0, len(ids))
		for _, id := range ids {
			recs = append(recs, fmt.Sprintf(`{
				"id": %q, "template_id": "kt_1", "tenant_id": "t_1", "template_version": 1,
				"values_data": {}, "is_active": true,
				"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"
			}`, id))
		}
		hasMore := callCount < len(pages)
		body := fmt.Sprintf(`{"records": [%s], "total": 3, "offset": 0, "limit": 2, "has_more": %v}`,
			joinJSON(recs), hasMore)
		_, _ = w.Write([]byte(body))
	})

	it := c.ListRecordsIterator(context.Background(), "kt_1", ListRecordsOptions{Limit: 2})
	var got []string
	for it.Next() {
		got = append(got, it.Record().ID)
	}
	if err := it.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	want := []string{"kr_1", "kr_2", "kr_3"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if callCount != 2 {
		t.Fatalf("callCount = %d, want 2 page fetches", callCount)
	}
}

func joinJSON(items []string) string {
	out := ""
	for i, item := range items {
		if i > 0 {
			out += ","
		}
		out += item
	}
	return out
}
