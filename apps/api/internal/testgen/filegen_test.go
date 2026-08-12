package testgen

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/testmanagement"
)

// fakeFileGenerationClient stands in for the ML service in tests, mirroring
// the fakeRunsRepo/fakeTestMgmtRepo pattern above.
type fakeFileGenerationClient struct {
	resp *fileGenerationResponse
	err  error
}

func (f *fakeFileGenerationClient) GenerateFromFile(_ context.Context, _ string, _ []byte, _ string) (*fileGenerationResponse, error) {
	return f.resp, f.err
}

func TestGenerateFromFileRequiresIdentifiers(t *testing.T) {
	svc := NewService(newFakeRunsRepo(), newFakeTestMgmtRepo())
	_, err := svc.GenerateFromFile(context.Background(), GenerateFromFileInput{
		WorkspaceID: uuid.Nil,
		ProjectID:   uuid.New(),
		Filename:    "cases.csv",
		Content:     []byte("Title\nSomething"),
		CreatedBy:   uuid.New(),
	})
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for a nil workspace id, got %v", err)
	}
}

func TestGenerateFromFileRejectsEmptyContent(t *testing.T) {
	svc := NewService(newFakeRunsRepo(), newFakeTestMgmtRepo())
	_, err := svc.GenerateFromFile(context.Background(), GenerateFromFileInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Filename:    "cases.csv",
		Content:     nil,
		CreatedBy:   uuid.New(),
	})
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for empty file content, got %v", err)
	}
}

// TestGenerateFromFileWithoutConfiguredClientFailsLoudly asserts the
// "no external LLM configured" default (NewService with no fileGen arg)
// returns a clear 503-shaped error rather than silently doing nothing or
// fabricating cases.
func TestGenerateFromFileWithoutConfiguredClientFailsLoudly(t *testing.T) {
	svc := NewService(newFakeRunsRepo(), newFakeTestMgmtRepo())
	_, err := svc.GenerateFromFile(context.Background(), GenerateFromFileInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Filename:    "cases.csv",
		Content:     []byte("Title\nLogin works"),
		CreatedBy:   uuid.New(),
	})
	var mlErr *MLServiceError
	if !errors.As(err, &mlErr) {
		t.Fatalf("expected an *MLServiceError, got %v (%T)", err, err)
	}
	if mlErr.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", mlErr.StatusCode)
	}
}

func TestGenerateFromFileCreatesCasesAndRun(t *testing.T) {
	runs := newFakeRunsRepo()
	testMgmt := newFakeTestMgmtRepo()
	fakeML := &fakeFileGenerationClient{
		resp: &fileGenerationResponse{
			RowCount: 2,
			Cases: []generatedCaseFromFile{
				{
					Title:         "Login succeeds with valid credentials",
					Description:   "Verify login works with correct credentials.",
					Preconditions: "Account exists",
					Priority:      "HIGH", // deliberately mixed case, to exercise normalizePriority
					Tags:          []string{"auth"},
					Steps: []generatedStepResponse{
						{Action: "Submit valid email/password", Expected: "Dashboard loads", TestData: "user@example.com"},
					},
				},
				{
					Title:    "Case with a bogus priority",
					Priority: "urgent!!", // not one of the 4 allowed values
					Steps: []generatedStepResponse{
						{Action: "Do something", Expected: "Something happens"},
					},
				},
			},
			SkippedRows: []skippedRowResponse{
				{Row: 2, Reason: "Row is empty and has no test intent."},
			},
		},
	}
	svc := NewService(runs, testMgmt, fakeML)

	wsID, projID, userID := uuid.New(), uuid.New(), uuid.New()
	result, err := svc.GenerateFromFile(context.Background(), GenerateFromFileInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Filename:    "cases.csv",
		Content:     []byte("Title,Steps\nLogin succeeds,...\n,\n"),
		Context:     "Login feature",
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Run.CaseCount != 2 {
		t.Errorf("expected run.CaseCount 2, got %d", result.Run.CaseCount)
	}
	if len(testMgmt.cases) != 2 {
		t.Fatalf("expected 2 persisted cases, found %d", len(testMgmt.cases))
	}
	if len(result.SkippedRows) != 1 || result.SkippedRows[0].Reason != "Row is empty and has no test intent." {
		t.Errorf("expected the skipped row to be surfaced, got %+v", result.SkippedRows)
	}
	if _, err := runs.GetRunByID(context.Background(), result.Run.ID); err != nil {
		t.Errorf("expected the generation run to be persisted: %v", err)
	}

	var sawHigh, sawFallbackMedium bool
	for _, tc := range testMgmt.cases {
		if tc.Status != testmanagement.TestCaseStatusPendingReview {
			t.Errorf("case %q: expected status pending_review, got %q", tc.Title, tc.Status)
		}
		if tc.Source != testmanagement.TestCaseSourceGeneratedFile {
			t.Errorf("case %q: expected source generated_file, got %q", tc.Title, tc.Source)
		}
		if tc.GenerationRunID == nil || *tc.GenerationRunID != result.Run.ID {
			t.Errorf("case %q: expected generation_run_id to point at the run", tc.Title)
		}
		switch tc.Priority {
		case testmanagement.TestCasePriorityHigh:
			sawHigh = true
		case testmanagement.TestCasePriorityMedium:
			sawFallbackMedium = true
		}
	}
	if !sawHigh {
		t.Error("expected the mixed-case 'HIGH' priority to normalize to 'high'")
	}
	if !sawFallbackMedium {
		t.Error("expected the invalid 'urgent!!' priority to fall back to 'medium'")
	}
}

// TestGenerateFromFileNeverFabricatesSkippedRows asserts that when the ML
// service reports zero usable cases, the service persists zero test cases
// (not placeholders) and still surfaces the skip reasons.
func TestGenerateFromFileNeverFabricatesSkippedRows(t *testing.T) {
	runs := newFakeRunsRepo()
	testMgmt := newFakeTestMgmtRepo()
	fakeML := &fakeFileGenerationClient{
		resp: &fileGenerationResponse{
			RowCount:    1,
			Cases:       nil,
			SkippedRows: []skippedRowResponse{{Row: 0, Reason: "no discernible test intent"}},
		},
	}
	svc := NewService(runs, testMgmt, fakeML)

	result, err := svc.GenerateFromFile(context.Background(), GenerateFromFileInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Filename:    "cases.csv",
		Content:     []byte("Title\n\n"),
		CreatedBy:   uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Cases) != 0 || len(testMgmt.cases) != 0 {
		t.Fatalf("expected zero fabricated cases, got %d in result / %d persisted", len(result.Cases), len(testMgmt.cases))
	}
	if result.Run.CaseCount != 0 {
		t.Errorf("expected run.CaseCount 0, got %d", result.Run.CaseCount)
	}
	if len(result.SkippedRows) != 1 {
		t.Fatalf("expected 1 skipped row, got %d", len(result.SkippedRows))
	}
}

func TestGenerateFromFilePropagatesMLServiceError(t *testing.T) {
	runs := newFakeRunsRepo()
	testMgmt := newFakeTestMgmtRepo()
	fakeML := &fakeFileGenerationClient{
		err: &MLServiceError{StatusCode: http.StatusBadGateway, Message: "AI generation is rate-limited right now. Try again shortly."},
	}
	svc := NewService(runs, testMgmt, fakeML)

	_, err := svc.GenerateFromFile(context.Background(), GenerateFromFileInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Filename:    "cases.csv",
		Content:     []byte("Title\nSomething"),
		CreatedBy:   uuid.New(),
	})
	var mlErr *MLServiceError
	if !errors.As(err, &mlErr) {
		t.Fatalf("expected an *MLServiceError, got %v (%T)", err, err)
	}
	if mlErr.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", mlErr.StatusCode)
	}
	if len(runs.runs) != 0 {
		t.Error("expected no generation run to be persisted when the ML call itself fails")
	}
}
