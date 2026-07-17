package automationhub

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/results"
)

type fakeResultsRepo struct {
	runs  map[uuid.UUID]*results.TestRun
	items map[uuid.UUID]*results.TestRunItem
}

func newFakeResultsRepo() *fakeResultsRepo {
	return &fakeResultsRepo{
		runs:  make(map[uuid.UUID]*results.TestRun),
		items: make(map[uuid.UUID]*results.TestRunItem),
	}
}

func (f *fakeResultsRepo) CreateRun(_ context.Context, run *results.TestRun) error {
	f.runs[run.ID] = run
	return nil
}

func (f *fakeResultsRepo) CreateItem(_ context.Context, item *results.TestRunItem) error {
	f.items[item.ID] = item
	return nil
}

func (f *fakeResultsRepo) UpdateRun(_ context.Context, run *results.TestRun) error {
	f.runs[run.ID] = run
	return nil
}

func TestIngestJUnit(t *testing.T) {
	svc := NewService(newFakeResultsRepo())
	wsID := uuid.New()
	projID := uuid.New()
	uid := uuid.New()

	junitXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="Suite 1" tests="3" failures="1" errors="0" skipped="1" time="2.5">
    <testcase name="Test A" classname="ClassA" time="0.5"/>
    <testcase name="Test B" classname="ClassB" time="1.0">
      <failure message="assertion failed" type="AssertionError">stack trace here</failure>
    </testcase>
    <testcase name="Test C" classname="ClassC" time="1.0" status="skipped"/>
  </testsuite>
</testsuites>`)

	result, err := svc.Ingest(context.Background(), IngestInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "CI Build #1",
		Format:      FormatJUnit,
		Body:        junitXML,
		CreatedBy:   uid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
	if result.Passed != 1 {
		t.Errorf("expected passed 1, got %d", result.Passed)
	}
	if result.Failed != 1 {
		t.Errorf("expected failed 1, got %d", result.Failed)
	}
	if result.Skipped != 1 {
		t.Errorf("expected skipped 1, got %d", result.Skipped)
	}
	if result.DurationMs != 2500 {
		t.Errorf("expected duration 2500ms, got %d", result.DurationMs)
	}
}

func TestIngestJUnitEmpty(t *testing.T) {
	svc := NewService(newFakeResultsRepo())

	_, err := svc.Ingest(context.Background(), IngestInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Name:        "Empty",
		Format:      FormatJUnit,
		Body:        []byte(`<testsuites></testsuites>`),
		CreatedBy:   uuid.New(),
	})
	if err == nil {
		t.Error("expected error for empty JUnit XML")
	}
}

func TestIngestPlaywright(t *testing.T) {
	svc := NewService(newFakeResultsRepo())
	wsID := uuid.New()
	projID := uuid.New()
	uid := uuid.New()

	report := []byte(`{
  "suites": [
    {
      "title": "Suite 1",
      "status": "completed",
      "tests": [
        {"title": "Test A", "status": "passed", "duration": 500},
        {"title": "Test B", "status": "failed", "duration": 1000, "error": "timeout"},
        {"title": "Test C", "status": "skipped", "duration": 0}
      ]
    }
  ]
}`)

	result, err := svc.Ingest(context.Background(), IngestInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Playwright Run",
		Format:      FormatPlaywright,
		Body:        report,
		CreatedBy:   uid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
	if result.Passed != 1 {
		t.Errorf("expected passed 1, got %d", result.Passed)
	}
	if result.Failed != 1 {
		t.Errorf("expected failed 1, got %d", result.Failed)
	}
	if result.Skipped != 1 {
		t.Errorf("expected skipped 1, got %d", result.Skipped)
	}
}

func TestIngestUnsupportedFormat(t *testing.T) {
	svc := NewService(newFakeResultsRepo())

	_, err := svc.Ingest(context.Background(), IngestInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Name:        "Bad",
		Format:      IngestionFormat("unsupported"),
		Body:        []byte("{}"),
		CreatedBy:   uuid.New(),
	})
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestIngestInvalidJUnitXML(t *testing.T) {
	svc := NewService(newFakeResultsRepo())

	_, err := svc.Ingest(context.Background(), IngestInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Name:        "Bad",
		Format:      FormatJUnit,
		Body:        []byte(`not xml`),
		CreatedBy:   uuid.New(),
	})
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestIsValidFormat(t *testing.T) {
	valid := []string{"junit", "playwright", "cypress"}
	for _, s := range valid {
		if !IsValidFormat(s) {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if IsValidFormat("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}
