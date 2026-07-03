# Changelog

All notable changes to this project are documented in this file.
This project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Initial client: `NewClient`, functional options (`WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithRetry`, `WithUserAgent`).
- Read-only template access: `ListTemplates`, `GetTemplate`, `GetLimits`.
- Record CRUD: `CreateRecord`, `GetRecord`, `UpdateRecord`, `DeleteRecord`, `ListRecords`, `ListRecordsIterator`.
- `values_data` helpers: `ValuesBuilder`, `AsString`, `AsNumber`, `AsBool`, `AsStringSlice`, `AsDate`.
- Full-text search (`Search`) and semantic search (`SemanticSearch`).
- CSV import (`ImportCSV`).
- Typed errors (`APIError`, `IsNotFound`/`IsUnauthorized`/`IsForbidden`/`IsValidationError`/`IsRateLimited`).
- Spec-drift test against a vendored copy of the backend's OpenAPI spec.
