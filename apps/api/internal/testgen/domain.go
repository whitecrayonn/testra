// Package testgen implements test case generation. GenerateFromSpec and
// GenerateFromEndpoint are deterministic and rule-based: per
// docs/BIBLICAL_TESTRA.md's "No External LLM" / "Transparent ML" principles,
// they never call a language model — every case they produce comes from
// plain code walking the OpenAPI spec structure (required fields, enums,
// types, string/numeric boundaries, auth requirements).
//
// GenerateFromFile (mlclient.go) is the one intentional, opt-in exception:
// it sends an uploaded spreadsheet's rows to the ML service, which calls an
// external LLM (Gemini) to draft test cases from loosely-structured input a
// rule engine can't parse. It only runs when ML_SERVICE_URL is configured,
// and — like every other path in this package — writes cases as
// pending_review so a human reviews them before they count toward coverage.
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
