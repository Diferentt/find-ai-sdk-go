package findai

import "time"

// FieldType is the data type of a template field.
type FieldType string

const (
	FieldTypeText        FieldType = "text"
	FieldTypeLongText    FieldType = "long_text"
	FieldTypeNumber      FieldType = "number"
	FieldTypeBoolean     FieldType = "boolean"
	FieldTypeDate        FieldType = "date"
	FieldTypeTime        FieldType = "time"
	FieldTypeDateTime    FieldType = "datetime"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multi_select"
	FieldTypeURL         FieldType = "url"
	FieldTypeEmail       FieldType = "email"
	FieldTypePhone       FieldType = "phone"
	FieldTypeLocation    FieldType = "location"
)

// FieldStatus is the lifecycle status of a template field definition.
type FieldStatus string

const (
	FieldStatusActive     FieldStatus = "active"
	FieldStatusDeprecated FieldStatus = "deprecated"
	FieldStatusDeleted    FieldStatus = "deleted"
)

// TemplateField describes one field in a dataset's schema.
type TemplateField struct {
	Name         string      `json:"name"`
	DisplayName  string      `json:"display_name"`
	Type         FieldType   `json:"type"`
	Required     bool        `json:"required"`
	Unique       bool        `json:"unique"`
	Searchable   bool        `json:"searchable"`
	Filterable   bool        `json:"filterable"`
	IsPublic     bool        `json:"is_public"`
	IsLLMVisible bool        `json:"is_llm_visible"`
	Options      []string    `json:"options,omitempty"`
	Status       FieldStatus `json:"status"`
}

// TemplateResponse is a dataset ("table") definition: its name and field
// schema. Templates are managed via the dashboard; this SDK only reads them.
type TemplateResponse struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Fields      []TemplateField `json:"fields"`
	Tags        []string        `json:"tags,omitempty"`
	Version     int             `json:"version"`
	IsActive    bool            `json:"is_active"`
	CreatedBy   *string         `json:"created_by,omitempty"`
	UpdatedBy   *string         `json:"updated_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// RecordResponse is one row in a dataset. ValuesData holds the row's field
// values, shaped by the owning template's field schema at the time it was
// written (see TemplateVersion).
type RecordResponse struct {
	ID              string         `json:"id"`
	TemplateID      string         `json:"template_id"`
	TenantID        string         `json:"tenant_id"`
	TemplateVersion int            `json:"template_version"`
	ValuesData      map[string]any `json:"values_data"`
	IsActive        bool           `json:"is_active"`
	CreatedBy       *string        `json:"created_by,omitempty"`
	UpdatedBy       *string        `json:"updated_by,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// LimitsResponse describes the effective knowledge/dataset limits for the
// authenticated tenant.
type LimitsResponse struct {
	MaxTemplates          int `json:"knowledge_max_templates"`
	MaxFieldsPerTemplate  int `json:"knowledge_max_fields_per_template"`
	MaxRecordsPerTemplate int `json:"knowledge_max_records_per_template"`
	MaxSearchableFields   int `json:"knowledge_max_searchable_fields"`
}
