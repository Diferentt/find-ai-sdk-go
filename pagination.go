package findai

import "context"

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// ListRecordsOptions controls pagination for ListRecords and
// ListRecordsIterator. Limit is clamped client-side to the API's [1, 100]
// range; a zero value uses the API default of 20.
type ListRecordsOptions struct {
	Offset int
	Limit  int
}

func (o ListRecordsOptions) normalizedLimit() int {
	switch {
	case o.Limit <= 0:
		return defaultListLimit
	case o.Limit > maxListLimit:
		return maxListLimit
	default:
		return o.Limit
	}
}

// RecordListPage is one page of a ListRecords call.
type RecordListPage struct {
	Records []RecordResponse `json:"records"`
	Total   int              `json:"total"`
	Offset  int              `json:"offset"`
	Limit   int              `json:"limit"`
	HasMore bool             `json:"has_more"`
}

// RecordIterator walks every record in a template across as many pages as
// needed. Use it as:
//
//	it := client.ListRecordsIterator(ctx, templateID, findai.ListRecordsOptions{})
//	for it.Next() {
//	    rec := it.Record()
//	    ...
//	}
//	if err := it.Err(); err != nil {
//	    ...
//	}
type RecordIterator struct {
	ctx        context.Context
	client     *Client
	templateID string
	opts       ListRecordsOptions

	buf     []RecordResponse
	pos     int
	fetched int
	done    bool
	err     error
}

// ListRecordsIterator returns an iterator over all records in templateID,
// starting at opts.Offset and auto-paginating with opts.Limit-sized pages.
func (c *Client) ListRecordsIterator(ctx context.Context, templateID string, opts ListRecordsOptions) *RecordIterator {
	return &RecordIterator{ctx: ctx, client: c, templateID: templateID, opts: opts}
}

// Next advances the iterator, fetching additional pages as needed. It
// returns false when iteration is complete or an error occurred (check Err).
func (it *RecordIterator) Next() bool {
	if it.err != nil {
		return false
	}
	for it.pos >= len(it.buf) {
		if it.done {
			return false
		}
		if err := it.fetchNextPage(); err != nil {
			it.err = err
			return false
		}
	}
	it.pos++
	return true
}

// Record returns the record at the iterator's current position. Only valid
// after a call to Next that returned true.
func (it *RecordIterator) Record() RecordResponse {
	return it.buf[it.pos-1]
}

// Err returns the first error encountered during iteration, if any.
func (it *RecordIterator) Err() error { return it.err }

func (it *RecordIterator) fetchNextPage() error {
	limit := it.opts.normalizedLimit()
	page, err := it.client.ListRecords(it.ctx, it.templateID, ListRecordsOptions{
		Offset: it.opts.Offset + it.fetched,
		Limit:  limit,
	})
	if err != nil {
		return err
	}
	it.buf = page.Records
	it.pos = 0
	it.fetched += len(page.Records)
	if !page.HasMore || len(page.Records) == 0 {
		it.done = true
	}
	return nil
}
