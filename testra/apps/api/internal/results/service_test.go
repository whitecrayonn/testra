package results

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type fakeRepository struct {
	runs  map[uuid.UUID]*TestRun
	items map[uuid.UUID]*TestRunItem
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		runs:  make(map[uuid.UUID]*TestRun),
		items: make(map[uuid.UUID]*TestRunItem),
	}
}

func (f *fakeRepository) CreateRun(_ context.Context, run *TestRun) error {
	f.runs[run.ID] = run
	return nil
}

func (f *fakeRepository) GetRunByID(_ context.Context, id uuid.UUID) (*TestRun, error) {
	if run, ok := f.runs[id]; ok {
		return run, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListRuns(_ context.Context, projectID uuid.UUID, cursor string, limit int) ([]TestRun, error) {
	var result []TestRun
	for _, run := range f.runs {
		if run.ProjectID == projectID {
			result = append(result, *run)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (f *fakeRepository) UpdateRun(_ context.Context, run *TestRun) error {
	if _, ok := f.runs[run.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.runs[run.ID] = run
	return nil
}

func (f *fakeRepository) DeleteRun(_ context.Context, id uuid.UUID) error {
	if _, ok := f.runs[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.runs, id)
	return nil
}

func (f *fakeRepository) CreateItem(_ context.Context, item *TestRunItem) error {
	f.items[item.ID] = item
	return nil
}

func (f *fakeRepository) GetItemByID(_ context.Context, id uuid.UUID) (*TestRunItem, error) {
	if item, ok := f.items[id]; ok {
		return item, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListItems(_ context.Context, runID uuid.UUID) ([]TestRunItem, error) {
	var result []TestRunItem
	for _, item := range f.items {
		if item.RunID == runID {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (f *fakeRepository) UpdateItem(_ context.Context, item *TestRunItem) error {
	if _, ok := f.items[item.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.items[item.ID] = item
	return nil
}

func (f *fakeRepository) DeleteItemsByRunID(_ context.Context, runID uuid.UUID) error {
	for id, item := range f.items {
		if item.RunID == runID {
			delete(f.items, id)
		}
	}
	return nil
}

func (f *fakeRepository) RunInTx(_ context.Context, fn func(Repository) error) error {
	return fn(f)
}

func TestServiceCreateRun(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	projID := uuid.New()
	uid := uuid.New()

	run, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Nightly Regression",
		Source:      RunSourceManual,
		CreatedBy:   uid,
		TestCaseIDs: []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run.Name != "Nightly Regression" {
		t.Errorf("expected name 'Nightly Regression', got %s", run.Name)
	}
	if run.Status != RunStatusPending {
		t.Errorf("expected status pending, got %s", run.Status)
	}
	if run.Total != 3 {
		t.Errorf("expected total 3, got %d", run.Total)
	}
	if run.Source != RunSourceManual {
		t.Errorf("expected source manual, got %s", run.Source)
	}
}

func TestServiceCreateRunInvalidName(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: uuid.New(),
		ProjectID:   uuid.New(),
		Name:        "",
		CreatedBy:   uuid.New(),
	})
	if err != sharederrors.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceUpdateRunStatus(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	projID := uuid.New()
	uid := uuid.New()

	run, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Run 1",
		CreatedBy:   uid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := svc.UpdateRunStatus(context.Background(), run.ID, RunStatusRunning)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != RunStatusRunning {
		t.Errorf("expected status running, got %s", updated.Status)
	}
	if updated.StartedAt == nil {
		t.Error("expected started_at to be set")
	}
}

func TestServiceUpdateRunStatusConflictOnTerminal(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	projID := uuid.New()
	uid := uuid.New()

	run, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Run 1",
		CreatedBy:   uid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.UpdateRunStatus(context.Background(), run.ID, RunStatusPassed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.UpdateRunStatus(context.Background(), run.ID, RunStatusRunning)
	if err != sharederrors.ErrConflict {
		t.Errorf("expected ErrConflict on terminal run, got %v", err)
	}
}

func TestServiceUpdateRunStatusInvalid(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	projID := uuid.New()

	run, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Run 1",
		CreatedBy:   uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.UpdateRunStatus(context.Background(), run.ID, RunStatus("invalid"))
	if err != sharederrors.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceDeleteRun(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	projID := uuid.New()

	run, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Run 1",
		CreatedBy:   uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := svc.DeleteRun(context.Background(), run.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := svc.DeleteRun(context.Background(), run.ID); err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestServiceUpdateItemStatus(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	projID := uuid.New()
	tcID := uuid.New()

	run, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Run 1",
		CreatedBy:   uuid.New(),
		TestCaseIDs: []uuid.UUID{tcID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, err := svc.ListItems(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	updated, err := svc.UpdateItemStatus(context.Background(), items[0].ID, RunItemStatusPassed, 500, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != RunItemStatusPassed {
		t.Errorf("expected status passed, got %s", updated.Status)
	}
	if updated.DurationMs != 500 {
		t.Errorf("expected duration 500, got %d", updated.DurationMs)
	}
}

func TestServiceUpdateItemStatusInvalid(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	projID := uuid.New()

	run, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Run 1",
		CreatedBy:   uuid.New(),
		TestCaseIDs: []uuid.UUID{uuid.New()},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, err := svc.ListItems(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.UpdateItemStatus(context.Background(), items[0].ID, RunItemStatus("invalid"), 0, "", "")
	if err != sharederrors.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceGetRunNotFound(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.GetRun(context.Background(), uuid.New())
	if err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestIsValidRunStatus(t *testing.T) {
	valid := []string{"pending", "running", "passed", "failed", "skipped", "cancelled"}
	for _, s := range valid {
		if !IsValidRunStatus(s) {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if IsValidRunStatus("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}

func TestIsValidRunItemStatus(t *testing.T) {
	valid := []string{"pending", "running", "passed", "failed", "skipped", "blocked"}
	for _, s := range valid {
		if !IsValidRunItemStatus(s) {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if IsValidRunItemStatus("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}

func TestIsTerminalRunStatus(t *testing.T) {
	terminal := []RunStatus{RunStatusPassed, RunStatusFailed, RunStatusSkipped, RunStatusCancelled}
	for _, s := range terminal {
		if !IsTerminalRunStatus(s) {
			t.Errorf("expected %s to be terminal", s)
		}
	}
	nonTerminal := []RunStatus{RunStatusPending, RunStatusRunning}
	for _, s := range nonTerminal {
		if IsTerminalRunStatus(s) {
			t.Errorf("expected %s to be non-terminal", s)
		}
	}
}

func TestSubscribeRunProgressTerminalRun(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	projID := uuid.New()

	run, err := svc.CreateRun(context.Background(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "Run 1",
		CreatedBy:   uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.UpdateRunStatus(context.Background(), run.ID, RunStatusPassed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ch, err := svc.SubscribeRunProgress(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed for terminal run")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected channel to be closed immediately")
	}
}
