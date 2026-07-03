package findai

import "context"

// GetLimits returns the effective dataset limits (max templates, max fields
// per template, max records per template, max searchable fields) for the
// authenticated tenant's current plan.
func (c *Client) GetLimits(ctx context.Context) (*LimitsResponse, error) {
	var out LimitsResponse
	if err := c.t.Do(ctx, "GET", "/api/v1/knowledge/limits", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
