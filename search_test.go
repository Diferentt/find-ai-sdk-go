package findai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSearch(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/templates/kt_1/search" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body SearchRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Query != "acme" {
			t.Errorf("query = %q", body.Query)
		}
		_, _ = w.Write([]byte(`{
			"hits": [{"record_id": "kr_1", "values_data": {"company_name": "Acme"}, "ranking_score": 0.9}],
			"query": "acme", "processing_time_ms": 3, "estimated_total_hits": 1
		}`))
	})

	resp, err := c.Search(context.Background(), "kt_1", SearchRequest{Query: "acme", Limit: 10})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(resp.Hits) != 1 || resp.Hits[0].RecordID != "kr_1" {
		t.Fatalf("unexpected hits: %+v", resp.Hits)
	}
	if resp.Hits[0].RankingScore == nil || *resp.Hits[0].RankingScore != 0.9 {
		t.Fatalf("unexpected ranking score: %+v", resp.Hits[0].RankingScore)
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"hits": []}`))
	})

	resp, err := c.Search(context.Background(), "kt_1", SearchRequest{Query: "nothing"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(resp.Hits) != 0 {
		t.Fatalf("expected no hits, got %+v", resp.Hits)
	}
}

func TestSemanticSearch(t *testing.T) {
	c := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/templates/kt_1/semantic-search" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body SemanticSearchRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.TopK != 5 {
			t.Errorf("top_k = %d", body.TopK)
		}
		_, _ = w.Write([]byte(`{
			"results": [{"score": 0.82, "values_data": {"company_name": "Acme"}}],
			"query": "software companies", "total": 1
		}`))
	})

	resp, err := c.SemanticSearch(context.Background(), "kt_1", SemanticSearchRequest{Query: "software companies", TopK: 5})
	if err != nil {
		t.Fatalf("SemanticSearch() error = %v", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].Score != 0.82 {
		t.Fatalf("unexpected results: %+v", resp.Results)
	}
	if resp.Results[0].RecordID != nil {
		t.Fatalf("expected nil RecordID, got %v", *resp.Results[0].RecordID)
	}
}
