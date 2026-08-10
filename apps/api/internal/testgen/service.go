package testgen

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/testmanagement"
)

// Service generates draft test cases from an OpenAPI spec and writes them
// through the testmanagement module's Repository port — the same
// cross-module composition pattern automationhub already uses (see
// apps/api/internal/automationhub/service.go, which takes a
// testmanagement.Repository the same way).
type Service struct {
	runs     Repository
	testMgmt testmanagement.Repository
	fileGen  FileGenerationClient
}

// NewService takes fileGen as a variadic argument so every existing 2-arg
// call site (tests, in particular) keeps compiling unchanged. When omitted,
// GenerateFromFile fails loudly with "not configured" rather than silently
// doing nothing.
func NewService(runs Repository, testMgmt testmanagement.Repository, fileGen ...FileGenerationClient) *Service {
	var fg FileGenerationClient = &unconfiguredFileGenerationClient{}
	if len(fileGen) > 0 && fileGen[0] != nil {
		fg = fileGen[0]
	}
	return &Service{runs: runs, testMgmt: testMgmt, fileGen: fg}
}

type GenerateFromSpecInput struct {
	WorkspaceID  uuid.UUID
	ProjectID    uuid.UUID
	Spec         map[string]interface{}
	SpecFilename string
	CreatedBy    uuid.UUID
}

type GenerateFromSpecResult struct {
	Run   *GenerationRun
	Cases []testmanagement.TestCase
}

// GenerateFromSpec parses an OpenAPI document and creates one pending_review
// test case per rule match (see openapi_parser.go for the full rule set).
// Every case it writes is source = generated_spec and status =
// pending_review; a human must call testmanagement's ApproveCase before a
// generated case counts toward coverage.
func (s *Service) GenerateFromSpec(ctx context.Context, input GenerateFromSpecInput) (*GenerateFromSpecResult, error) {
	if input.WorkspaceID == uuid.Nil || input.ProjectID == uuid.Nil || input.CreatedBy == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.Spec == nil {
		return nil, sharederrors.ErrInvalidInput
	}

	drafts, endpointCount := GenerateDraftCases(input.Spec)
	if len(drafts) == 0 {
		return nil, sharederrors.ErrInvalidInput
	}

	return s.persistGeneratedCases(ctx, input.WorkspaceID, input.ProjectID, input.CreatedBy, input.SpecFilename, drafts, endpointCount)
}

// GenerateFromEndpointInput describes a single HTTP endpoint (method, path,
// and its query/path/header/body fields) typed directly into the "quick
// generate" form — no OpenAPI document required.
type GenerateFromEndpointInput struct {
	WorkspaceID  uuid.UUID
	ProjectID    uuid.UUID
	Method       string
	Path         string
	Fields       []EndpointField
	RequiresAuth bool
	CreatedBy    uuid.UUID
}

// GenerateFromEndpoint generates draft cases for one endpoint described
// directly by the caller (method/path/fields), rather than parsed from an
// uploaded OpenAPI spec. It builds the same map[string]interface{} shape an
// OpenAPI document would produce for that single operation (buildEndpointSpec
// in endpoint_spec.go) and reuses GenerateDraftCases unchanged, so the rule
// set — required fields, enums, wrong types, auth — is identical to the
// full-spec path. Boundary-value cases only appear when GenerateDraftCases
// finds min/max constraints on a field's schema; the quick form does not
// collect those, so only the full-spec path produces boundary cases today.
func (s *Service) GenerateFromEndpoint(ctx context.Context, input GenerateFromEndpointInput) (*GenerateFromSpecResult, error) {
	if input.WorkspaceID == uuid.Nil || input.ProjectID == uuid.Nil || input.CreatedBy == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	method := strings.TrimSpace(input.Method)
	path := strings.TrimSpace(input.Path)
	if method == "" || path == "" {
		return nil, sharederrors.ErrInvalidInput
	}

	spec := buildEndpointSpec(method, path, input.Fields, input.RequiresAuth)
	drafts, endpointCount := GenerateDraftCases(spec)
	if len(drafts) == 0 {
		return nil, sharederrors.ErrInvalidInput
	}

	filename := strings.ToUpper(method) + " " + path
	return s.persistGeneratedCases(ctx, input.WorkspaceID, input.ProjectID, input.CreatedBy, filename, drafts, endpointCount)
}

// persistGeneratedCases writes the generation_run audit record and one
// pending_review test case per draft. Shared by GenerateFromSpec and
// GenerateFromEndpoint so both entry points persist identically.
func (s *Service) persistGeneratedCases(ctx context.Context, workspaceID, projectID, createdBy uuid.UUID, specFilename string, drafts []draftCase, endpointCount int) (*GenerateFromSpecResult, error) {
	now := time.Now().UTC()
	run := &GenerationRun{
		ID:            uuid.New(),
		WorkspaceID:   workspaceID,
		ProjectID:     projectID,
		Source:        string(testmanagement.TestCaseSourceGeneratedSpec),
		SpecFilename:  specFilename,
		EndpointCount: endpointCount,
		CaseCount:     len(drafts),
		CreatedBy:     createdBy,
		CreatedAt:     now,
	}

	// The run record is created before the cases so their generation_run_id
	// foreign key is always valid. If the case-creation transaction below
	// fails, this run row still records the attempt (case_count reflects
	// what was *intended*, not necessarily what committed) — an acceptable
	// tradeoff for an audit trail that spans two modules' repositories and
	// therefore can't share a single database transaction.
	if err := s.runs.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	cases := make([]testmanagement.TestCase, len(drafts))
	for i, d := range drafts {
		runID := run.ID
		cases[i] = testmanagement.TestCase{
			ID:            uuid.New(),
			WorkspaceID:   workspaceID,
			ProjectID:     projectID,
			Title:         d.Title,
			Description:   "Generated from an imported OpenAPI spec. Review before approving.",
			Preconditions: d.Preconditions,
			Steps: []testmanagement.TestStep{
				{Order: 1, Action: d.Action, Expected: d.ExpectedResult},
			},
			Status:          testmanagement.TestCaseStatusPendingReview,
			Priority:        testmanagement.TestCasePriority(d.Priority),
			Version:         1,
			CreatedBy:       createdBy,
			CreatedAt:       now,
			UpdatedAt:       now,
			Source:          testmanagement.TestCaseSourceGeneratedSpec,
			GenerationRunID: &runID,
		}
	}

	err := s.testMgmt.RunInTx(ctx, func(tx testmanagement.Repository) error {
		for i := range cases {
			if err := tx.CreateCase(ctx, &cases[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &GenerateFromSpecResult{Run: run, Cases: cases}, nil
}

// GenerateFromFileInput describes an uploaded spreadsheet to run through the
// ML service's LLM-backed generator.
type GenerateFromFileInput struct {
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID
	Filename    string
	Content     []byte
	Context     string
	CreatedBy   uuid.UUID
}

// SkippedRow is a row the LLM could not confidently turn into a test case —
// surfaced to the caller instead of silently dropped or fabricated.
type SkippedRow struct {
	Row    int
	Reason string
}

type GenerateFromFileResult struct {
	Run         *GenerationRun
	Cases       []testmanagement.TestCase
	SkippedRows []SkippedRow
}

// GenerateFromFile is the one path in this package that calls an external
// LLM (via the ML service) rather than generating deterministically — see
// the package doc comment for why. Every case it writes is source =
// generated_file and status = pending_review, the same human-review gate
// GenerateFromSpec/GenerateFromEndpoint use.
func (s *Service) GenerateFromFile(ctx context.Context, input GenerateFromFileInput) (*GenerateFromFileResult, error) {
	if input.WorkspaceID == uuid.Nil || input.ProjectID == uuid.Nil || input.CreatedBy == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if len(input.Content) == 0 {
		return nil, sharederrors.ErrInvalidInput
	}

	resp, err := s.fileGen.GenerateFromFile(ctx, input.Filename, input.Content, input.Context)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	run := &GenerationRun{
		ID:            uuid.New(),
		WorkspaceID:   input.WorkspaceID,
		ProjectID:     input.ProjectID,
		Source:        string(testmanagement.TestCaseSourceGeneratedFile),
		SpecFilename:  input.Filename,
		EndpointCount: resp.RowCount,
		CaseCount:     len(resp.Cases),
		CreatedBy:     input.CreatedBy,
		CreatedAt:     now,
	}
	if err := s.runs.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	cases := make([]testmanagement.TestCase, len(resp.Cases))
	for i, c := range resp.Cases {
		runID := run.ID
		steps := make([]testmanagement.TestStep, len(c.Steps))
		for j, st := range c.Steps {
			steps[j] = testmanagement.TestStep{Order: j + 1, Action: st.Action, Expected: st.Expected, TestData: st.TestData}
		}
		cases[i] = testmanagement.TestCase{
			ID:              uuid.New(),
			WorkspaceID:     input.WorkspaceID,
			ProjectID:       input.ProjectID,
			Title:           c.Title,
			Description:     c.Description,
			Preconditions:   c.Preconditions,
			Steps:           steps,
			Status:          testmanagement.TestCaseStatusPendingReview,
			Priority:        testmanagement.TestCasePriority(normalizePriority(c.Priority)),
			Tags:            c.Tags,
			Version:         1,
			CreatedBy:       input.CreatedBy,
			CreatedAt:       now,
			UpdatedAt:       now,
			Source:          testmanagement.TestCaseSourceGeneratedFile,
			GenerationRunID: &runID,
		}
	}

	if len(cases) > 0 {
		err := s.testMgmt.RunInTx(ctx, func(tx testmanagement.Repository) error {
			for i := range cases {
				if err := tx.CreateCase(ctx, &cases[i]); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	skipped := make([]SkippedRow, len(resp.SkippedRows))
	for i, sr := range resp.SkippedRows {
		skipped[i] = SkippedRow{Row: sr.Row, Reason: sr.Reason}
	}

	return &GenerateFromFileResult{Run: run, Cases: cases, SkippedRows: skipped}, nil
}

// normalizePriority re-validates the ML service's priority value on the Go
// side too (defense in depth) rather than trusting the upstream response.
func normalizePriority(p string) string {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "low", "medium", "high", "critical":
		return strings.ToLower(strings.TrimSpace(p))
	default:
		return "medium"
	}
}
