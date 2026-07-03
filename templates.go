package findai

import "context"

// ListTemplates returns all dataset templates ("tables") owned by the
// authenticated tenant. There is no pagination on this endpoint — it always
// returns the full list.
func (c *Client) ListTemplates(ctx context.Context) ([]TemplateResponse, error) {
	var out []TemplateResponse
	if err := c.t.Do(ctx, "GET", "/api/v1/knowledge/templates", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTemplate returns one dataset template, including its field schema
// (Fields), by ID.
func (c *Client) GetTemplate(ctx context.Context, templateID string) (*TemplateResponse, error) {
	var out TemplateResponse
	path := "/api/v1/knowledge/templates/" + templateID
	if err := c.t.Do(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
