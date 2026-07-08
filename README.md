# find-ai-sdk-go

[![CI](https://github.com/Diferentt/find-ai-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Diferentt/find-ai-sdk-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Diferentt/find-ai-sdk-go.svg)](https://pkg.go.dev/github.com/Diferentt/find-ai-sdk-go)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

A Go client for FindAI Studio's **datasets** (a.k.a. the knowledge module): list your dataset tables, inspect their field schema, and create/read/update/delete rows — plus full-text search, semantic search, and CSV import.

This SDK is deliberately scoped to *data* operations. Creating or editing a dataset's schema (its "table structure") is a dashboard-only action; the SDK works with tables that already exist.

## Install

```sh
go get github.com/Diferentt/find-ai-sdk-go
```

Requires Go 1.22+.

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	findai "github.com/Diferentt/find-ai-sdk-go"
)

func main() {
	client, err := findai.NewClient(
		"fai_your_api_key",
		findai.WithBaseURL("https://api.en-kel.com"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	templates, err := client.ListTemplates(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range templates {
		fmt.Printf("%s: %s (%d fields)\n", t.ID, t.Name, len(t.Fields))
	}
}
```

More runnable examples: [`examples/basic_crud`](examples/basic_crud), [`examples/search`](examples/search), [`examples/csv_import`](examples/csv_import).

## Authentication

Requests are authenticated with a tenant-scoped API key (format `fai_...`) carrying the `dataset:manage` scope. Create one from your FindAI Studio dashboard's API keys section, then pass it to `NewClient`.

Note: `https://api.en-kel.com` is the API host — different from `https://app.en-kel.com`, which is the dashboard web app. `WithBaseURL` must point at the API host.

## Usage

### Templates (read-only)

A "template" is a dataset's schema — its name and field definitions. The SDK only reads templates; create/edit/delete them from the dashboard.

```go
templates, err := client.ListTemplates(ctx)
tmpl, err := client.GetTemplate(ctx, "kt_abc123")
for _, f := range tmpl.Fields {
    fmt.Println(f.Name, f.Type, f.Required)
}
```

### Records (CRUD)

A record is one row. Its values are arbitrary JSON shaped by the owning template's field schema, represented as `map[string]any`.

```go
rec, err := client.CreateRecord(ctx, templateID, map[string]any{
    "company_name": "Acme Corp",
    "founded_year": 1999,
})

rec, err = client.GetRecord(ctx, templateID, rec.ID)
rec, err = client.UpdateRecord(ctx, templateID, rec.ID, map[string]any{"founded_year": 2000})
err = client.DeleteRecord(ctx, templateID, rec.ID)
```

Optional sugar for building/reading values without stringly-typed map literals:

```go
values := findai.NewValuesBuilder().
    Set("company_name", "Acme Corp").
    Set("founded_year", 1999).
    Build()

name, _ := findai.AsString(rec.ValuesData, "company_name")
year, _ := findai.AsNumber(rec.ValuesData, "founded_year")
```

### Pagination

`ListRecords` mirrors the API's `offset`/`limit` contract directly:

```go
page, err := client.ListRecords(ctx, templateID, findai.ListRecordsOptions{Offset: 0, Limit: 50})
for _, r := range page.Records {
    // ...
}
if page.HasMore {
    // fetch the next page
}
```

For walking every record without manually tracking offsets, use the iterator:

```go
it := client.ListRecordsIterator(ctx, templateID, findai.ListRecordsOptions{})
for it.Next() {
    rec := it.Record()
    // ...
}
if err := it.Err(); err != nil {
    log.Fatal(err)
}
```

### Search

```go
resp, err := client.Search(ctx, templateID, findai.SearchRequest{Query: "acme", Limit: 20})

results, err := client.SemanticSearch(ctx, templateID, findai.SemanticSearchRequest{
    Query: "enterprise software companies",
    TopK:  10,
})
```

### CSV import

```go
f, err := os.Open("companies.csv")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

result, err := client.ImportCSV(ctx, templateID, "companies.csv", f)
fmt.Printf("imported %d/%d rows\n", result.Imported, result.TotalRows)
for _, rowErr := range result.Errors {
    fmt.Printf("row %d: %s\n", rowErr.Row, rowErr.Error)
}
```

### Webchat visitor tokens

For webchat connections with `auth_mode="signed"`, generate a JWT for your end-users from your backend:

```go
token, err := findai.GenerateWebchatVisitorToken(
    "your_webhook_secret", // from the webchat connection config
    "end_user_123",        // your app's user identifier
    time.Hour,             // token TTL (default: 1h if <= 0)
)
// send `token` to your frontend; it attaches it to webchat requests
```

The token is a HS256 JWT with claims `{visitor_id, exp}` signed with the connection's `webhook_secret`. No external dependencies required.

## Error handling

Every non-2xx response is returned as `*findai.APIError`:

```go
_, err := client.GetRecord(ctx, templateID, "kr_missing")
if findai.IsNotFound(err) {
    // ...
}

var apiErr *findai.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Detail)
    for _, fe := range apiErr.Errors { // populated for request-body validation errors
        fmt.Println(fe.Loc, fe.Msg)
    }
}
```

Helpers: `IsNotFound`, `IsUnauthorized`, `IsForbidden`, `IsValidationError`, `IsRateLimited`.

## Configuration

```go
client, err := findai.NewClient(apiKey,
    findai.WithBaseURL("https://api.en-kel.com"),                   // required
    findai.WithHTTPClient(customHTTPClient),                         // optional
    findai.WithTimeout(10 * time.Second),                            // optional
    findai.WithRetry(3, 200*time.Millisecond),                       // optional; retries 429/5xx/network errors
    findai.WithUserAgent("my-app/1.0"),                              // optional
)
```

## Versioning

This module follows [SemVer](https://semver.org/). It's currently pre-1.0 (`v0.x`) — the API may still shift release to release. See [CHANGELOG.md](CHANGELOG.md).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Apache-2.0](LICENSE)
