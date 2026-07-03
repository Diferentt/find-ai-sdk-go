package findai

import "context"

// SearchRequest is a full-text search query against a dataset's records.
type SearchRequest struct {
	Query   string            `json:"query"`
	Filters map[string]string `json:"filters,omitempty"`
	Limit   int               `json:"limit,omitempty"`
	Offset  int               `json:"offset,omitempty"`
}

// SearchHit is one full-text search result.
type SearchHit struct {
	RecordID        string         `json:"record_id"`
	ValuesData      map[string]any `json:"values_data"`
	TemplateVersion *int           `json:"template_version,omitempty"`
	RankingScore    *float64       `json:"ranking_score,omitempty"`
	CreatedAt       *string        `json:"created_at,omitempty"`
}

// SearchResponse is the result of a full-text search.
type SearchResponse struct {
	Hits               []SearchHit `json:"hits"`
	Query              *string     `json:"query,omitempty"`
	ProcessingTimeMs   *int        `json:"processing_time_ms,omitempty"`
	EstimatedTotalHits *int        `json:"estimated_total_hits,omitempty"`
}

// SemanticSearchRequest is a semantic (embedding-based) search query.
type SemanticSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

// SemanticSearchResult is one semantic search result.
type SemanticSearchResult struct {
	RecordID        *string        `json:"record_id,omitempty"`
	Score           float64        `json:"score"`
	ValuesData      map[string]any `json:"values_data"`
	TemplateVersion *int           `json:"template_version,omitempty"`
	CreatedAt       *string        `json:"created_at,omitempty"`
}

// SemanticSearchResponse is the result of a semantic search.
type SemanticSearchResponse struct {
	Results []SemanticSearchResult `json:"results"`
	Query   string                 `json:"query"`
	Total   int                    `json:"total"`
}

// Search runs a full-text search over templateID's records.
func (c *Client) Search(ctx context.Context, templateID string, req SearchRequest) (*SearchResponse, error) {
	var out SearchResponse
	path := "/api/v1/knowledge/templates/" + templateID + "/search"
	if err := c.t.Do(ctx, "POST", path, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SemanticSearch runs a semantic (embedding-based) search over templateID's
// records.
func (c *Client) SemanticSearch(ctx context.Context, templateID string, req SemanticSearchRequest) (*SemanticSearchResponse, error) {
	var out SemanticSearchResponse
	path := "/api/v1/knowledge/templates/" + templateID + "/semantic-search"
	if err := c.t.Do(ctx, "POST", path, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
