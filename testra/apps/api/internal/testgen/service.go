package testgen

import (
	"context"
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
}

func NewService(runs Repository, testMgmt testmanagement.Repository) *Service {
	return &Service{runs: runs, testMgmt: testMgmt}
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

	now := time.Now().UTC()
	run := &GenerationRun{
		ID:            uuid.New(),
		WorkspaceID:   input.WorkspaceID,
		ProjectID:     input.ProjectID,
		Source:        string(testmanagement.TestCaseSourceGeneratedSpec),
		SpecFilename:  input.SpecFilename,
		EndpointCount: endpointCount,
		CaseCount:     len(drafts),
		CreatedBy:     input.CreatedBy,
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
			WorkspaceID:   input.WorkspaceID,
			ProjectID:     input.ProjectID,
			Title:         d.Title,
			Description:   "Generated from an imported OpenAPI spec. Review before approving.",
			Preconditions: d.Preconditions,
			Steps: []testmanagement.TestStep{
				{Order: 1, Action: d.Action, Expected: d.ExpectedResult},
			},
			Status:          testmanagement.TestCaseStatusPendingReview,
			Priority:        testmanagement.TestCasePriority(d.Priority),
			Version:         1,
			CreatedBy:       input.CreatedBy,
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
