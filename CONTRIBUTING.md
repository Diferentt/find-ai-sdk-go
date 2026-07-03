# Contributing

## Dev setup

Requires Go 1.22+. No other dependencies — this module has zero third-party
runtime dependencies by design.

```sh
git clone https://github.com/Diferentt/find-ai-sdk-go
cd find-ai-sdk-go
make build
make test
```

## Before opening a PR

```sh
make fmt    # gofmt + goimports
make vet
make lint   # requires golangci-lint: https://golangci-lint.run/welcome/install/
make test
```

All four must be clean. CI enforces the same checks.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) style
(`fix:`, `feat:`, `docs:`, `chore:`, ...) — release notes are generated from
commit history.

## Keeping DTOs in sync with the backend

`docs/openapi.json` is **not** the backend's full OpenAPI spec — it's a
trimmed subset containing only the schemas the SDK's DTOs are checked
against (`internal/spectest/drift_test.go`), produced by
`internal/tools/trimspec`. The full backend spec covers unrelated internal
modules (billing, admin, chat, webhooks, ingest, ...) that have no business
being vendored into this public repo, so never `cp` it in directly.

If you're changing a hand-written DTO in `types.go`, `search.go`, or
`importcsv.go` in response to a backend change, refresh the vendored subset
via the tool, not a raw copy:

```sh
make sync-spec BACKEND_REPO=/path/to/find_ai_studio
make test-spec
```

If you add a new DTO that should be checked, add its schema name to
`wantedSchemas` in `internal/tools/trimspec/main.go` *and* a corresponding
case in `internal/spectest/drift_test.go`, then re-run `make sync-spec`.
The drift test fails loudly if the spec has a field your Go struct doesn't
model yet.

## Adding a new endpoint

1. Add/extend the DTO(s) in the relevant top-level file (`types.go` for
   shared types, or a dedicated file like `search.go` for a new resource
   area).
2. Add the client method, following the existing pattern: `context.Context`
   first, use `c.t.Do` (or `c.t.DoMultipart` for file uploads) from
   `internal/transport`.
3. Add unit tests using `httptest.NewServer` (see `records_test.go` for the
   pattern) covering the success path, at least one error path, and any
   edge cases in the response shape (optional/nullable fields).
4. If it's a new schema, add a case to `internal/spectest/drift_test.go`.
