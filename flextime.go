package findai

import (
	"fmt"
	"strings"
	"time"
)

// FlexTime wraps time.Time to handle API timestamps that may lack a timezone
// suffix (e.g. "2026-06-19T13:04:10" instead of RFC3339).
type FlexTime struct {
	time.Time
}

var flexFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
}

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		ft.Time = time.Time{}
		return nil
	}
	for _, layout := range flexFormats {
		if t, err := time.Parse(layout, s); err == nil {
			ft.Time = t
			return nil
		}
	}
	return fmt.Errorf("findai: cannot parse time %q", s)
}

func (ft FlexTime) MarshalJSON() ([]byte, error) {
	if ft.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + ft.Format(time.RFC3339) + `"`), nil
}
