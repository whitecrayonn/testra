package testgen

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/testmanagement"
)

// fakeRunsRepo is a minimal in-memory Repository for tests, mirroring the
// fake-repo pattern used throughout this codebase (see
// testmanagement/service_test.go and automationhub/service_test.go).
type fakeRunsRepo struct {
	runs map[uuid.UUID]*GenerationRun
}

func newFakeRunsRepo() *fakeRunsRepo {
	return &fakeRunsRepo{runs: make(map[uuid.UUID]*GenerationRun)}
}

func (f *fakeRunsRepo) CreateRun(_ context.Context, run *GenerationRun) error {
	f.runs[run.ID] = run
	return nil
}

func (f *fakeRunsRepo) GetRunByID(_ context.Context, id uuid.UUID) (*GenerationRun, error) {
	if run, ok := f.runs[id]; ok {
		return run, nil
	}
	return nil, sharederrors.ErrNotFound
}

// fakeTestMgmtRepo mirrors automationhub's fake of the same interface,
// but actually stores created cases so tests can assert on them.
type fakeTestMgmtRepo struct {
	cases map[uuid.UUID]*testmanagement.TestCase
}

func newFakeTestMgmtRepo() *fakeTestMgmtRepo {
	return &fakeTestMgmtRepo{cases: make(map[uuid.UUID]*testmanagement.TestCase)}
}

func (f *fakeTestMgmtRepo) CreateFolder(_ context.Context, _ *testmanagement.TestFolder) error {
	return nil
}
func (f *fakeTestMgmtRepo) GetFolderByID(_ context.Context, _ uuid.UUID) (*testmanagement.TestFolder, error) {
	return nil, fmt.Errorf("not found")
}
func (f *fakeTestMgmtRepo) ListFolders(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ string, _ int) ([]testmanagement.TestFolder, error) {
	return nil, nil
}
func (f *fakeTestMgmtRepo) UpdateFolder(_ context.Context, _ *testmanagement.TestFolder) error {
	return nil
}
func (f *fakeTestMgmtRepo) DeleteFolder(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeTestMgmtRepo) CreateSuite(_ context.Context, _ *testmanagement.TestSuite) error {
	return nil
}
func (f *fakeTestMgmtRepo) GetSuiteByID(_ context.Context, _ uuid.UUID) (*testmanagement.TestSuite, error) {
	return nil, fmt.Errorf("not found")
}
func (f *fakeTestMgmtRepo) ListSuites(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ string, _ int) ([]testmanagement.TestSuite, error) {
	return nil, nil
}
func (f *fakeTestMgmtRepo) UpdateSuite(_ context.Context, _ *testmanagement.TestSuite) error {
	return nil
}
func (f *fakeTestMgmtRepo) DeleteSuite(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeTestMgmtRepo) CreateCase(_ context.Context, tc *testmanagement.TestCase) error {
	f.cases[tc.ID] = tc
	return nil
}
func (f *fakeTestMgmtRepo) GetCaseByID(_ context.Context, id uuid.UUID) (*testmanagement.TestCase, error) {
	if tc, ok := f.cases[id]; ok {
		return tc, nil
	}
	return nil, sharederrors.ErrNotFound
}
func (f *fakeTestMgmtRepo) ListCases(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ string, _ int) ([]testmanagement.TestCase, error) {
	return nil, nil
}
func (f *fakeTestMgmtRepo) SearchCases(_ context.Context, _ uuid.UUID, _, _ string, _ int) ([]testmanagement.TestCase, string, error) {
	return nil, "", nil
}
func (f *fakeTestMgmtRepo) UpdateCase(_ context.Context, tc *testmanagement.TestCase) error {
	if _, ok := f.cases[tc.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.cases[tc.ID] = tc
	return nil
}
func (f *fakeTestMgmtRepo) DeleteCase(_ context.Context, id uuid.UUID) error {
	delete(f.cases, id)
	return nil
}
func (f *fakeTestMgmtRepo) CreateVersion(_ context.Context, _ *testmanagement.TestCaseVersion) error {
	return nil
}
func (f *fakeTestMgmtRepo) ListVersions(_ context.Context, _ uuid.UUID, _ string, _ int) ([]testmanagement.TestCaseVersion, error) {
	return nil, nil
}
func (f *fakeTestMgmtRepo) RunInTx(_ context.Context, fn func(testmanagement.Repository) error) error {
	return fn(f)
}

func simpleSpec() map[string]interface{} {
	return map[string]interface{}{
		"paths": map[string]interface{}{
			"/ping": map[string]interface{}{
				"get": map[string]interface{}{
					"responses": map[string]interface{}{"200": map[string]interface{}{"description": "OK"}},
				},
			},
		},
	}
}

func TestGenerateFromSpecRequiresIdentifiers(t *testing.T) {
	svc := NewService(newFakeRunsRepo(), newFakeTestMgmtRepo())
	_, err := svc.GenerateFromSpec(context.Background(), GenerateFromSpecInput{
		WorkspaceID: uuid.Nil,
		ProjectID:   uuid.New(),
		Spec:        simpleSpec(),
		CreatedBy:   uuid.New(),
	})
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for a nil workspace id, got %v", err)
	}
}

func TestGenerateFromSpecRejectsEmptySpec(t *testing.T) {
	svc := NewService(newFakeRunsRepo(), newFakeTestMgmtRepo())
	_, err := svc.GenerateFromSpec(context.Background(), GenerateFromSpecInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Spec:        map[string]interface{}{},
		CreatedBy:   uuid.New(),
	})
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for a spec with no paths, got %v", err)
	}
}

func TestGenerateFromSpecCreatesCasesAndRun(t *testing.T) {
	runs := newFakeRunsRepo()
	testMgmt := newFakeTestMgmtRepo()
	svc := NewService(runs, testMgmt)

	wsID, projID, userID := uuid.New(), uuid.New(), uuid.New()
	result, err := svc.GenerateFromSpec(context.Background(), GenerateFromSpecInput{
		WorkspaceID:  wsID,
		ProjectID:    projID,
		SpecFilename: "sample.json",
		Spec:         simpleSpec(),
		CreatedBy:    userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Run.EndpointCount != 1 {
		t.Errorf("expected 1 endpoint, got %d", result.Run.EndpointCount)
	}
	if result.Run.CaseCount != len(result.Cases) {
		t.Errorf("run.CaseCount (%d) should match len(cases) (%d)", result.Run.CaseCount, len(result.Cases))
	}
	if len(testMgmt.cases) != len(result.Cases) {
		t.Fatalf("expected all %d generated cases to be persisted, found %d", len(result.Cases), len(testMgmt.cases))
	}
	if _, err := runs.GetRunByID(context.Background(), result.Run.ID); err != nil {
		t.Errorf("expected the generation run to be persisted: %v", err)
	}

	for _, tc := range testMgmt.cases {
		if tc.Status != testmanagement.TestCaseStatusPendingReview {
			t.Errorf("case %q: expected status pending_review, got %q", tc.Title, tc.Status)
		}
		if tc.Source != testmanagement.TestCaseSourceGeneratedSpec {
			t.Errorf("case %q: expected source generated_spec, got %q", tc.Title, tc.Source)
		}
		if tc.GenerationRunID == nil || *tc.GenerationRunID != result.Run.ID {
			t.Errorf("case %q: expected generation_run_id to point at the run", tc.Title)
		}
		if tc.WorkspaceID != wsID || tc.ProjectID != projID {
			t.Errorf("case %q: workspace/project id was not propagated from the request", tc.Title)
		}
	}
}

// Approving a generated case is testmanagement's responsibility (ApproveCase),
// not testgen's — this test documents that boundary: testgen only ever
// creates pending_review cases, never active ones.
func TestGeneratedCasesAreNeverActiveByDefault(t *testing.T) {
	testMgmt := newFakeTestMgmtRepo()
	svc := NewService(newFakeRunsRepo(), testMgmt)

	_, err := svc.GenerateFromSpec(context.Background(), GenerateFromSpecInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Spec:        simpleSpec(),
		CreatedBy:   uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tc := range testMgmt.cases {
		if tc.Status == testmanagement.TestCaseStatusActive {
			t.Errorf("case %q should not be active immediately after generation", tc.Title)
		}
	}
}
