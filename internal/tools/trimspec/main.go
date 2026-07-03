// Command trimspec produces the vendored docs/openapi.json used by
// internal/spectest from a full backend OpenAPI spec.
//
// The full backend spec covers every module in FindAI Studio (billing,
// admin, chat, webhooks, ingest, ...), most of which has nothing to do with
// this SDK. Vendoring it verbatim into a public repo would leak internal
// API surface that isn't this SDK's concern. This tool instead extracts
// only the schemas the SDK's DTOs are checked against (see
// internal/spectest/drift_test.go), following $ref chains so nothing
// referenced is silently dropped.
//
// Usage:
//
//	go run ./internal/tools/trimspec <path-to-full-openapi.json> <output-path>
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// wantedSchemas must stay in sync with the specName values used in
// internal/spectest/drift_test.go.
var wantedSchemas = []string{
	"src__interfaces__api__v1__knowledge_models__TemplateResponse",
	"RecordResponse",
	"RecordListResponse",
	"FulltextSearchResponse",
	"FulltextSearchRequest",
	"SearchHit",
	"SemanticSearchResponse",
	"SemanticSearchRequest",
	"SemanticSearchHit",
	"CSVImportResponse",
	"CSVImportError",
	"KnowledgeLimitsResponse",
	"FieldDefinition",
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: trimspec <path-to-full-openapi.json> <output-path>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "trimspec:", err)
		os.Exit(1)
	}
}

func run(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", inputPath, err)
	}

	var full map[string]any
	if err := json.Unmarshal(data, &full); err != nil {
		return fmt.Errorf("parsing %s: %w", inputPath, err)
	}

	components, _ := full["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	if schemas == nil {
		return fmt.Errorf("no components.schemas found in %s", inputPath)
	}

	included := make(map[string]bool)
	for _, name := range wantedSchemas {
		collectRefs(name, schemas, included)
	}

	trimmedSchemas := make(map[string]any, len(included))
	names := make([]string, 0, len(included))
	for name := range included {
		if s, ok := schemas[name]; ok {
			trimmedSchemas[name] = s
			names = append(names, name)
		}
	}
	sort.Strings(names)

	securitySchemes, _ := components["securitySchemes"].(map[string]any)
	trimmedSecurity := map[string]any{}
	if bearer, ok := securitySchemes["BearerAuth"]; ok {
		trimmedSecurity["BearerAuth"] = bearer
	}

	info, _ := full["info"].(map[string]any)
	version := ""
	if info != nil {
		if v, ok := info["version"].(string); ok {
			version = v
		}
	}

	minimal := map[string]any{
		"openapi": full["openapi"],
		"info": map[string]any{
			"title":       "Find AI Studio API (knowledge/datasets subset)",
			"version":     version,
			"description": "Trimmed to only the schemas used by find-ai-sdk-go's spec-drift test (internal/spectest). Not the full backend API surface — see internal/tools/trimspec.",
		},
		"components": map[string]any{
			"schemas":         trimmedSchemas,
			"securitySchemes": trimmedSecurity,
		},
	}

	out, err := json.MarshalIndent(minimal, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding output: %w", err)
	}
	out = append(out, '\n')

	if err := os.WriteFile(outputPath, out, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outputPath, err)
	}

	fmt.Printf("wrote %s: %d schema(s) (%s)\n", outputPath, len(names), strings.Join(names, ", "))
	return nil
}

// collectRefs walks schema and everything it (transitively) $refs, adding
// each schema name to included.
func collectRefs(name string, schemas map[string]any, included map[string]bool) {
	if included[name] {
		return
	}
	schema, ok := schemas[name]
	if !ok {
		return
	}
	included[name] = true
	walkRefs(schema, schemas, included)
}

func walkRefs(node any, schemas map[string]any, included map[string]bool) {
	switch v := node.(type) {
	case map[string]any:
		if ref, ok := v["$ref"].(string); ok {
			const prefix = "#/components/schemas/"
			if strings.HasPrefix(ref, prefix) {
				collectRefs(strings.TrimPrefix(ref, prefix), schemas, included)
			}
			return
		}
		for _, val := range v {
			walkRefs(val, schemas, included)
		}
	case []any:
		for _, item := range v {
			walkRefs(item, schemas, included)
		}
	}
}
