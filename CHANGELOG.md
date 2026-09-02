# Changelog

All notable changes to this project are documented in this file.
This project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Tasks: `InvokeTask` / `InvokeTaskWithState` run a flow without a conversation and return its result synchronously (`flows:execute` scope); `GetTask` reports the parameters a task accepts (`Inputs`). New `IsConflict` error helper (409: task disabled).
- `TemplateResponse.Slug` and `LimitsResponse.ReservedFieldNames`, fields the backend added since the spec was last vendored (caught by the spec-drift test on refresh).
- Initial client: `NewClient`, functional options (`WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithRetry`, `WithUserAgent`).
- Read-only template access: `ListTemplates`, `GetTemplate`, `GetLimits`.
- Record CRUD: `CreateRecord`, `GetRecord`, `UpdateRecord`, `DeleteRecord`, `ListRecords`, `ListRecordsIterator`.
- `values_data` helpers: `ValuesBuilder`, `AsString`, `AsNumber`, `AsBool`, `AsStringSlice`, `AsDate`.
- Full-text search (`Search`) and semantic search (`SemanticSearch`).
- CSV import (`ImportCSV`).
- Typed errors (`APIError`, `IsNotFound`/`IsUnauthorized`/`IsForbidden`/`IsValidationError`/`IsRateLimited`).
- Spec-drift test against a vendored copy of the backend's OpenAPI spec.
