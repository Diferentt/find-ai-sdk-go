package findai

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// CSVImportRowError describes one row that failed to import.
type CSVImportRowError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

// CSVImportResponse summarizes the result of a CSV import.
type CSVImportResponse struct {
	TotalRows int                 `json:"total_rows"`
	Imported  int                 `json:"imported"`
	Errors    []CSVImportRowError `json:"errors"`
}

// ImportCSV bulk-creates records in templateID from a CSV file. filename is
// used only for the multipart upload's declared filename and must end in
// ".csv" — the server rejects anything else. The full content of r is read
// into memory before sending.
func (c *Client) ImportCSV(ctx context.Context, templateID, filename string, r io.Reader) (*CSVImportResponse, error) {
	if !strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return nil, fmt.Errorf("findai: ImportCSV: filename %q must end in .csv", filename)
	}

	var out CSVImportResponse
	path := "/api/v1/knowledge/templates/" + templateID + "/import"
	if err := c.t.DoMultipart(ctx, "POST", path, filename, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
