package findai

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
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	// Slug is the stable machine name flows reference the template by;
	// derived from Name when not set explicitly.
	Slug        *string         `json:"slug,omitempty"`
	Description *string         `json:"description,omitempty"`
	Fields      []TemplateField `json:"fields"`
	Tags        []string        `json:"tags,omitempty"`
	Version     int             `json:"version"`
	IsActive    bool            `json:"is_active"`
	CreatedBy   *string         `json:"created_by,omitempty"`
	UpdatedBy   *string         `json:"updated_by,omitempty"`
	CreatedAt   FlexTime        `json:"created_at"`
	UpdatedAt   FlexTime        `json:"updated_at"`
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
	CreatedAt       FlexTime       `json:"created_at"`
	UpdatedAt       FlexTime       `json:"updated_at"`
}

// LimitsResponse describes the effective knowledge/dataset limits for the
// authenticated tenant.
type LimitsResponse struct {
	MaxTemplates          int `json:"knowledge_max_templates"`
	MaxFieldsPerTemplate  int `json:"knowledge_max_fields_per_template"`
	MaxRecordsPerTemplate int `json:"knowledge_max_records_per_template"`
	MaxSearchableFields   int `json:"knowledge_max_searchable_fields"`
	// ReservedFieldNames are field names a template cannot use: the document
	// pipeline claims them for its own keys.
	ReservedFieldNames []string `json:"reserved_field_names"`
}

// TaskResponse is one task: a flow executed without a conversation, run by
// its cron schedule (when it has one), by InvokeTask, or both. Served under
// "scheduled-jobs" paths for historical reasons.
type TaskResponse struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	FlowID   string `json:"flow_id"`
	// VersionID pins the flow version; nil runs whatever version is active
	// at execution time.
	VersionID *int `json:"version_id"`
	// ScheduleType and ScheduleExpression are both nil for an API-only task
	// (no cron), and always set together otherwise.
	ScheduleType       *string        `json:"schedule_type"`
	ScheduleExpression *string        `json:"schedule_expression"`
	Timezone           string         `json:"timezone"`
	InitialState       map[string]any `json:"initial_state"`
	Enabled            bool           `json:"enabled"`
	CreatedAt          FlexTime       `json:"created_at"`
	UpdatedAt          FlexTime       `json:"updated_at"`
	LastRunAt          *FlexTime      `json:"last_run_at"`
	LastRunStatus      *string        `json:"last_run_status"`
	EventbridgeStatus  string         `json:"eventbridge_status"`
	EventbridgeError   *string        `json:"eventbridge_error"`
	Warnings           []string       `json:"warnings"`
	// Inputs are the parameter names the task accepts in InvokeTask (nested
	// under initial_state.params), derived from the flow's nodes. Only the
	// single-task GET reports them; list endpoints leave this nil.
	Inputs []string `json:"inputs"`
}

// TaskInvokeResponse is the result of InvokeTask.
type TaskInvokeResponse struct {
	JobID      string `json:"job_id"`
	Success    bool   `json:"success"`
	DurationMS int    `json:"duration_ms"`
	// Output is what the flow wrote under the "output" state bucket — the
	// task's result by convention (a generate node in JSON mode writes there
	// by default). Nil when the flow wrote nothing under "output"; the full
	// state then comes back in FinalState instead.
	Output map[string]any `json:"output"`
	// FinalState is the flow's complete final state, only populated when
	// Output is nil (kept for flows that predate the output convention).
	FinalState    map[string]any `json:"final_state"`
	NodesExecuted []string       `json:"nodes_executed"`
	// Error carries the failure reason when Success is false; the HTTP call
	// itself still succeeds (a flow that ran and failed is a 200 with
	// Success=false, not an APIError).
	Error *string `json:"error"`
}
