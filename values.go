package findai

import "time"

// ValuesBuilder is optional sugar for constructing a record's values_data
// map without repeating map[string]any{...} literal syntax. It has no
// validation of its own — the backend validates values against the
// template's field schema on write.
type ValuesBuilder struct {
	values map[string]any
}

// NewValuesBuilder returns an empty ValuesBuilder.
func NewValuesBuilder() *ValuesBuilder {
	return &ValuesBuilder{values: make(map[string]any)}
}

// Set assigns a field value and returns the builder for chaining.
func (b *ValuesBuilder) Set(field string, value any) *ValuesBuilder {
	b.values[field] = value
	return b
}

// Build returns the underlying values_data map.
func (b *ValuesBuilder) Build() map[string]any {
	return b.values
}

// AsString reads a TEXT/LONG_TEXT/URL/EMAIL/PHONE/SELECT-typed field value
// out of a record's ValuesData.
func AsString(values map[string]any, field string) (string, bool) {
	v, ok := values[field]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// AsNumber reads a NUMBER-typed field value. JSON numbers always decode to
// float64 in Go, so that is what's returned.
func AsNumber(values map[string]any, field string) (float64, bool) {
	v, ok := values[field]
	if !ok {
		return 0, false
	}
	n, ok := v.(float64)
	return n, ok
}

// AsBool reads a BOOLEAN-typed field value.
func AsBool(values map[string]any, field string) (bool, bool) {
	v, ok := values[field]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// AsStringSlice reads a MULTI_SELECT-typed field value.
func AsStringSlice(values map[string]any, field string) ([]string, bool) {
	v, ok := values[field]
	if !ok {
		return nil, false
	}
	raw, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
}

// AsDate parses a DATE/DATETIME-typed field value using layout. Callers
// working with DATETIME fields should pass time.RFC3339; DATE fields are
// typically "2006-01-02".
func AsDate(values map[string]any, field, layout string) (time.Time, bool) {
	s, ok := AsString(values, field)
	if !ok {
		return time.Time{}, false
	}
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
