package results

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type Service struct {
	repo     Repository
	progress *progressHub
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:     repo,
		progress: newProgressHub(),
	}
}

func (s *Service) CreateRun(ctx context.Context, input CreateRunInput) (*TestRun, error) {
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	now := time.Now().UTC()
	metadata := make(map[string]interface{})
	if input.PlanID != nil {
		metadata["plan_id"] = input.PlanID.String()
	}
	run := &TestRun{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		ProjectID:   input.ProjectID,
		SuiteID:     input.SuiteID,
		Name:        input.Name,
		Status:      RunStatusPending,
		Total:       len(input.TestCaseIDs),
		Source:      input.Source,
		Metadata:    metadata,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	eventbus.Default().Publish(ctx, eventbus.Event{
		Type:     "test_run.created",
		TenantID: input.WorkspaceID.String(),
		Payload: map[string]interface{}{
			"run_id":       run.ID.String(),
			"project_id":   input.ProjectID.String(),
			"workspace_id": input.WorkspaceID.String(),
			"status":       string(run.Status),
		},
	})

	for i, tcID := range input.TestCaseIDs {
		item := &TestRunItem{
			ID:         uuid.New(),
			RunID:      run.ID,
			TestCaseID: &tcID,
			Title:      fmt.Sprintf("Test case %d", i+1),
			Status:     RunItemStatusPending,
			SortOrder:  i,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := s.repo.CreateItem(ctx, item); err != nil {
			return nil, err
		}
	}

	return run, nil
}

func (s *Service) GetRun(ctx context.Context, id uuid.UUID) (*TestRun, error) {
	return s.repo.GetRunByID(ctx, id)
}

func (s *Service) ListRuns(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]TestRun, error) {
	return s.repo.ListRuns(ctx, projectID, cursor, limit)
}

func (s *Service) UpdateRunStatus(ctx context.Context, id uuid.UUID, status RunStatus) (*TestRun, error) {
	if !IsValidRunStatus(string(status)) {
		return nil, sharederrors.ErrInvalidInput
	}

	run, err := s.repo.GetRunByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if IsTerminalRunStatus(run.Status) {
		return nil, sharederrors.ErrConflict
	}

	now := time.Now().UTC()
	run.Status = status
	run.UpdatedAt = now

	if status == RunStatusRunning && run.StartedAt == nil {
		run.StartedAt = &now
	}

	if IsTerminalRunStatus(status) {
		run.CompletedAt = &now
	}

	if err := s.repo.UpdateRun(ctx, run); err != nil {
		return nil, err
	}

	s.progress.broadcast(id, RunProgressEvent{
		RunID:    id,
		Status:   string(status),
		Total:    run.Total,
		Passed:   run.Passed,
		Failed:   run.Failed,
		Skipped:  run.Skipped,
		Blocked:  run.Blocked,
		Progress: float64(run.Passed+run.Failed+run.Skipped+run.Blocked) / float64(max(run.Total, 1)),
	})

	eventType := "test_run.status_changed"
	if IsTerminalRunStatus(status) {
		eventType = "test_run.completed"
	}
	eventbus.Default().Publish(ctx, eventbus.Event{
		Type:     eventType,
		TenantID: run.WorkspaceID.String(),
		Payload: map[string]interface{}{
			"run_id":       run.ID.String(),
			"project_id":   run.ProjectID.String(),
			"workspace_id": run.WorkspaceID.String(),
			"status":       string(status),
			"passed":       run.Passed,
			"failed":       run.Failed,
			"skipped":      run.Skipped,
			"blocked":      run.Blocked,
			"total":        run.Total,
		},
	})

	return run, nil
}

func (s *Service) DeleteRun(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRun(ctx, id)
}

func (s *Service) ListItems(ctx context.Context, runID uuid.UUID) ([]TestRunItem, error) {
	return s.repo.ListItems(ctx, runID)
}

func (s *Service) UpdateItemStatus(ctx context.Context, itemID uuid.UUID, status RunItemStatus, durationMs int64, errMsg string, stackTrace string) (*TestRunItem, error) {
	if !IsValidRunItemStatus(string(status)) {
		return nil, sharederrors.ErrInvalidInput
	}

	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	item.Status = status
	item.DurationMs = durationMs
	item.ErrorMessage = errMsg
	item.StackTrace = stackTrace
	item.UpdatedAt = now

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	run, err := s.repo.GetRunByID(ctx, item.RunID)
	if err == nil {
		if err := s.recalcRunCounts(ctx, run); err != nil {
			return nil, err
		}
		s.progress.broadcast(run.ID, RunProgressEvent{
			RunID:    run.ID,
			ItemID:   item.ID,
			Status:   string(status),
			Total:    run.Total,
			Passed:   run.Passed,
			Failed:   run.Failed,
			Skipped:  run.Skipped,
			Blocked:  run.Blocked,
			Progress: float64(run.Passed+run.Failed+run.Skipped+run.Blocked) / float64(max(run.Total, 1)),
		})
	}

	return item, nil
}

func (s *Service) SubscribeRunProgress(ctx context.Context, runID uuid.UUID) (<-chan RunProgressEvent, error) {
	run, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if IsTerminalRunStatus(run.Status) {
		ch := make(chan RunProgressEvent)
		close(ch)
		return ch, nil
	}
	return s.progress.subscribe(runID), nil
}

func (s *Service) recalcRunCounts(ctx context.Context, run *TestRun) error {
	items, err := s.repo.ListItems(ctx, run.ID)
	if err != nil {
		return err
	}
	run.Total = len(items)
	run.Passed = 0
	run.Failed = 0
	run.Skipped = 0
	run.Blocked = 0
	var totalDuration int64
	for _, item := range items {
		totalDuration += item.DurationMs
		switch item.Status {
		case RunItemStatusPassed:
			run.Passed++
		case RunItemStatusFailed:
			run.Failed++
		case RunItemStatusSkipped:
			run.Skipped++
		case RunItemStatusBlocked:
			run.Blocked++
		}
	}
	run.DurationMs = totalDuration
	run.UpdatedAt = time.Now().UTC()
	return s.repo.UpdateRun(ctx, run)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type progressHub struct {
	mu   sync.RWMutex
	subs map[uuid.UUID][]chan RunProgressEvent
}

func newProgressHub() *progressHub {
	return &progressHub{subs: make(map[uuid.UUID][]chan RunProgressEvent)}
}

func (h *progressHub) subscribe(runID uuid.UUID) <-chan RunProgressEvent {
	ch := make(chan RunProgressEvent, 16)
	h.mu.Lock()
	h.subs[runID] = append(h.subs[runID], ch)
	h.mu.Unlock()
	return ch
}

func (h *progressHub) broadcast(runID uuid.UUID, event RunProgressEvent) {
	h.mu.RLock()
	subs := h.subs[runID]
	h.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
	if IsTerminalRunStatus(RunStatus(event.Status)) {
		h.mu.Lock()
		for _, ch := range h.subs[runID] {
			close(ch)
		}
		delete(h.subs, runID)
		h.mu.Unlock()
	}
}
