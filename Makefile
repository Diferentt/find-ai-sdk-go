.PHONY: build test test-integration test-spec lint vet fmt sync-spec

build:
	go build ./...

test:
	go test ./... -race -cover

test-integration:
	go test -tags=integration ./... -run Integration -v

test-spec:
	go test -tags=spec ./internal/spectest/... -v

lint:
	golangci-lint run

vet:
	go vet ./...

fmt:
	gofmt -l -w .
	goimports -w .

# Refresh the vendored docs/openapi.json from a sibling checkout of the
# backend repo. This does NOT copy the full backend spec — it extracts only
# the schemas internal/spectest checks against (see
# internal/tools/trimspec), since the full spec covers unrelated internal
# modules (billing, admin, chat, webhooks, ...) that have no place in this
# public repo. Override BACKEND_REPO to point elsewhere, e.g.:
#   make sync-spec BACKEND_REPO=../find_ai_studio
BACKEND_REPO ?= ../find_ai_studio
sync-spec:
	go run ./internal/tools/trimspec $(BACKEND_REPO)/docs/openapi.json docs/openapi.json
