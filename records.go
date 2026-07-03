package findai

import (
	"context"
	"net/url"
	"strconv"
)

func recordsPath(templateID string) string {
	return "/api/v1/knowledge/templates/" + templateID + "/records"
}

func recordPath(templateID, recordID string) string {
	return recordsPath(templateID) + "/" + recordID
}

// CreateRecord adds a new row to the templateID dataset. valuesData is
// validated server-side against the template's field schema.
func (c *Client) CreateRecord(ctx context.Context, templateID string, valuesData map[string]any) (*RecordResponse, error) {
	body := map[string]any{"values_data": valuesData}
	var out RecordResponse
	if err := c.t.Do(ctx, "POST", recordsPath(templateID), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRecord returns one row by ID.
func (c *Client) GetRecord(ctx context.Context, templateID, recordID string) (*RecordResponse, error) {
	var out RecordResponse
	if err := c.t.Do(ctx, "GET", recordPath(templateID, recordID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateRecord replaces the values of an existing row.
func (c *Client) UpdateRecord(ctx context.Context, templateID, recordID string, valuesData map[string]any) (*RecordResponse, error) {
	body := map[string]any{"values_data": valuesData}
	var out RecordResponse
	if err := c.t.Do(ctx, "PUT", recordPath(templateID, recordID), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteRecord soft-deletes a row by ID.
func (c *Client) DeleteRecord(ctx context.Context, templateID, recordID string) error {
	return c.t.Do(ctx, "DELETE", recordPath(templateID, recordID), nil, nil, nil)
}

// ListRecords returns one page of rows for templateID. Use
// ListRecordsIterator to walk every row across multiple pages.
func (c *Client) ListRecords(ctx context.Context, templateID string, opts ListRecordsOptions) (*RecordListPage, error) {
	q := url.Values{
		"offset": {strconv.Itoa(opts.Offset)},
		"limit":  {strconv.Itoa(opts.normalizedLimit())},
	}
	var out RecordListPage
	if err := c.t.Do(ctx, "GET", recordsPath(templateID), q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
