package findai

import "testing"

func TestValuesBuilder(t *testing.T) {
	v := NewValuesBuilder().
		Set("company_name", "Acme").
		Set("founded_year", 1999).
		Build()

	if v["company_name"] != "Acme" || v["founded_year"] != 1999 {
		t.Fatalf("unexpected values: %+v", v)
	}
}

func TestAsHelpers(t *testing.T) {
	values := map[string]any{
		"name":    "Acme",
		"count":   float64(42),
		"active":  true,
		"tags":    []any{"a", "b"},
		"founded": "2026-01-01",
	}

	if s, ok := AsString(values, "name"); !ok || s != "Acme" {
		t.Fatalf("AsString = %q, %v", s, ok)
	}
	if n, ok := AsNumber(values, "count"); !ok || n != 42 {
		t.Fatalf("AsNumber = %v, %v", n, ok)
	}
	if b, ok := AsBool(values, "active"); !ok || !b {
		t.Fatalf("AsBool = %v, %v", b, ok)
	}
	if tags, ok := AsStringSlice(values, "tags"); !ok || len(tags) != 2 || tags[1] != "b" {
		t.Fatalf("AsStringSlice = %v, %v", tags, ok)
	}
	if _, ok := AsString(values, "missing"); ok {
		t.Fatalf("expected AsString to fail for missing field")
	}
	if d, ok := AsDate(values, "founded", "2006-01-02"); !ok || d.Year() != 2026 {
		t.Fatalf("AsDate = %v, %v", d, ok)
	}
}
