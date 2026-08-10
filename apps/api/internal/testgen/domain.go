// Package testgen implements deterministic, rule-based test case generation
// from an imported OpenAPI spec. Per docs/BIBLICAL_TESTRA.md's "No External
// LLM" / "Transparent ML" principles (also recorded as a permanent fact in
// docs/AI_MEMORY.md: "No external LLM processing; ML must be transparent,
// explainable, and tenant-isolated"), this package never calls a language
// model. Every generated test case is produced by plain code walking the
// OpenAPI spec structure (required fields, enums, types, string/numeric
// boundaries, auth requirements) — the same rule set is documented in
// docs/architecture/adrs and is fully reproducible and explainable.
package testgen

import (
	"time"

	"github.com/google/uuid"
)

// GenerationRun records one generation job: which spec was used, how many
// endpoints and draft cases it produced, and who ran it. It is the audit
// trail for the generation feature, analogous to audit_events but scoped to
// this feature so it can carry generation-specific fields (endpoint/case
// counts) without overloading the generic audit log.
type GenerationRun struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	ProjectID     uuid.UUID
	Source        string
	SpecFilename  string
	EndpointCount int
	CaseCount     int
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
}

// EndpointField describes one query/path/header/body field of a
// single-endpoint quick-generate request. It carries the same information an
// OpenAPI parameter or request body property would (see field in
// openapi_parser.go) so buildEndpointSpec can translate it directly into the
// same map[string]interface{} shape GenerateDraftCases already parses.
type EndpointField struct {
	Name     string
	Location string // query | path | header | body
	Type     string // string | integer | number | boolean
	Required bool
	Enum     []string
}
