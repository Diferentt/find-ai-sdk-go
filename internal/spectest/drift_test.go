//go:build spec

// Package spectest checks that the SDK's hand-written DTOs (in the
// top-level findai package) haven't silently drifted from the backend's
// OpenAPI spec vendored at docs/openapi.json. It intentionally avoids any
// OpenAPI-aware dependency: this is a structural presence check (does every
// spec property have a matching Go json tag, and vice versa), not a full
// JSON-Schema-to-Go type equivalence checker, so plain encoding/json plus
// reflection is sufficient for both 3.0 and 3.1 specs.
//
// Run with: go test -tags=spec ./internal/spectest/...
package spectest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	findai "github.com/Diferentt/find-ai-sdk-go"
)

func loadSpec(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "openapi.json"))
	if err != nil {
		t.Fatalf("reading vendored spec (run `make sync-spec` if missing): %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parsing spec: %v", err)
	}
	return doc
}

func schemaProps(t *testing.T, doc map[string]any, name string) map[string]bool {
	t.Helper()
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	schema, ok := schemas[name].(map[string]any)
	if !ok {
		t.Fatalf("schema %q not found in vendored spec — was it renamed? update the mapping in this test", name)
	}
	props, _ := schema["properties"].(map[string]any)
	out := make(map[string]bool, len(props))
	for k := range props {
		out[k] = true
	}
	return out
}

func jsonTags(v any) map[string]bool {
	rt := reflect.TypeOf(v)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	out := make(map[string]bool)
	for i := 0; i < rt.NumField(); i++ {
		tag := rt.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		out[strings.Split(tag, ",")[0]] = true
	}
	return out
}

// assertFieldsMatch fails the test if the spec schema has a property with no
// corresponding Go json tag (real drift risk: a new/renamed backend field
// the SDK doesn't know about). Go fields absent from the spec are logged
// only, since the SDK is allowed to model a schema more richly (e.g.
// optional convenience fields) without that being a bug.
func assertFieldsMatch(t *testing.T, specName string, goValue any, doc map[string]any) {
	t.Helper()
	specProps := schemaProps(t, doc, specName)
	goTags := jsonTags(goValue)

	var missingInGo, extraInGo []string
	for p := range specProps {
		if !goTags[p] {
			missingInGo = append(missingInGo, p)
		}
	}
	for g := range goTags {
		if !specProps[g] {
			extraInGo = append(extraInGo, g)
		}
	}
	sort.Strings(missingInGo)
	sort.Strings(extraInGo)

	if len(missingInGo) > 0 {
		t.Errorf("%s: spec has properties with no Go field — update the corresponding struct: %v", specName, missingInGo)
	}
	if len(extraInGo) > 0 {
		t.Logf("%s: Go struct has fields not present in spec (informational, not a failure): %v", specName, extraInGo)
	}
}

func TestDTOsMatchOpenAPISpec(t *testing.T) {
	doc := loadSpec(t)

	cases := []struct {
		specName string
		goValue  any
	}{
		{"src__interfaces__api__v1__knowledge_models__TemplateResponse", findai.TemplateResponse{}},
		{"RecordResponse", findai.RecordResponse{}},
		{"RecordListResponse", findai.RecordListPage{}},
		{"FulltextSearchResponse", findai.SearchResponse{}},
		{"SearchHit", findai.SearchHit{}},
		{"SemanticSearchResponse", findai.SemanticSearchResponse{}},
		{"SemanticSearchHit", findai.SemanticSearchResult{}},
		{"CSVImportResponse", findai.CSVImportResponse{}},
		{"CSVImportError", findai.CSVImportRowError{}},
		{"KnowledgeLimitsResponse", findai.LimitsResponse{}},
		{"FieldDefinition", findai.TemplateField{}},
		{"ScheduledJobResponse", findai.TaskResponse{}},
		{"TaskInvokeResponse", findai.TaskInvokeResponse{}},
	}

	for _, tc := range cases {
		assertFieldsMatch(t, tc.specName, tc.goValue, doc)
	}
}
