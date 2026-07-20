package results

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

// ExecuteItem records the full manual execution state for a run item, including
// per-step results, comments, duration, executor, and a history snapshot.
func (s *Service) ExecuteItem(ctx context.Context, itemID uuid.UUID, input ExecuteItemInput) (*TestRunItem, error) {
	if !IsValidRunItemStatus(string(input.Status)) {
		return nil, sharederrors.ErrInvalidInput
	}

	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	item.Status = input.Status
	item.StepResults = input.StepResults
	item.Comment = input.Comment
	item.DurationMs = input.DurationMs
	item.ErrorMessage = input.ErrorMessage
	item.StackTrace = input.StackTrace
	item.ExecutedBy = &input.ExecutedBy
	item.ExecutedAt = &now
	item.UpdatedAt = now

	if err := s.repo.RunInTx(ctx, func(tx Repository) error {
		if err := tx.CreateItemExecution(ctx, item); err != nil {
			return err
		}
		history := &RunItemHistory{
			ID:          uuid.New(),
			RunItemID:   item.ID,
			Status:      item.Status,
			StepResults: item.StepResults,
			Comment:     item.Comment,
			DurationMs:  item.DurationMs,
			ExecutedBy:  item.ExecutedBy,
			CreatedAt:   now,
		}
		return tx.CreateItemHistory(ctx, history)
	}); err != nil {
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
			Status:   string(item.Status),
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

func (s *Service) ListItemHistory(ctx context.Context, itemID uuid.UUID) ([]RunItemHistory, error) {
	return s.repo.ListItemHistory(ctx, itemID)
}

func (s *Service) AttachEvidence(ctx context.Context, input EvidenceInput) (*EvidenceRef, error) {
	if input.RunItemID == uuid.Nil || input.FileName == "" {
		return nil, sharederrors.ErrInvalidInput
	}

	now := time.Now().UTC()
	evidence := &EvidenceRef{
		ID:          uuid.New(),
		RunItemID:   input.RunItemID,
		StepOrder:   input.StepOrder,
		FileName:    input.FileName,
		ContentType: input.ContentType,
		StoragePath: input.StoragePath,
		CreatedAt:   now,
	}
	if input.UploadedBy != uuid.Nil {
		evidence.UploadedBy = &input.UploadedBy
	}
	if err := s.repo.CreateEvidence(ctx, evidence); err != nil {
		return nil, err
	}
	return evidence, nil
}

func (s *Service) ListEvidence(ctx context.Context, itemID uuid.UUID) ([]EvidenceRef, error) {
	return s.repo.ListEvidenceByItem(ctx, itemID)
}

func (s *Service) DeleteEvidence(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteEvidence(ctx, id)
}

func (s *Service) LinkDefect(ctx context.Context, itemID, defectID uuid.UUID) error {
	if itemID == uuid.Nil || defectID == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.CreateRunItemDefect(ctx, itemID, defectID)
}

func (s *Service) UnlinkDefect(ctx context.Context, itemID, defectID uuid.UUID) error {
	if itemID == uuid.Nil || defectID == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteRunItemDefect(ctx, itemID, defectID)
}

func (s *Service) ListItemDefects(ctx context.Context, itemID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.ListRunItemDefects(ctx, itemID)
}

func (s *Service) ListItemsPaged(ctx context.Context, runID uuid.UUID, status, search, cursor string, limit int) ([]TestRunItem, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListItemsByRunPaged(ctx, runID, status, search, cursor, limit)
}

func (s *Service) BulkUpdateItems(ctx context.Context, runID uuid.UUID, itemIDs []uuid.UUID, status RunItemStatus, executedBy uuid.UUID) ([]TestRunItem, error) {
	if !IsValidRunItemStatus(string(status)) {
		return nil, sharederrors.ErrInvalidInput
	}

	var updated []TestRunItem
	now := time.Now().UTC()
	for _, itemID := range itemIDs {
		item, err := s.repo.GetItemByID(ctx, itemID)
		if err != nil {
			return nil, err
		}
		if item.RunID != runID {
			return nil, sharederrors.ErrInvalidInput
		}
		item.Status = status
		item.ExecutedBy = &executedBy
		item.ExecutedAt = &now
		item.UpdatedAt = now
		if err := s.repo.CreateItemExecution(ctx, item); err != nil {
			return nil, err
		}
		history := &RunItemHistory{
			ID:         uuid.New(),
			RunItemID:  item.ID,
			Status:     item.Status,
			Comment:    item.Comment,
			DurationMs: item.DurationMs,
			ExecutedBy: item.ExecutedBy,
			CreatedAt:  now,
		}
		if err := s.repo.CreateItemHistory(ctx, history); err != nil {
			return nil, err
		}
		updated = append(updated, *item)
	}

	run, err := s.repo.GetRunByID(ctx, runID)
	if err == nil {
		if err := s.recalcRunCounts(ctx, run); err != nil {
			return nil, err
		}
	}
	return updated, nil
}

func (s *Service) CloneRun(ctx context.Context, runID uuid.UUID, createdBy uuid.UUID) (*TestRun, error) {
	orig, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListItems(ctx, runID)
	if err != nil {
		return nil, err
	}

	tcIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.TestCaseID != nil {
			tcIDs = append(tcIDs, *item.TestCaseID)
		}
	}

	return s.CreateRun(ctx, CreateRunInput{
		WorkspaceID: orig.WorkspaceID,
		ProjectID:   orig.ProjectID,
		SuiteID:     orig.SuiteID,
		Name:        fmt.Sprintf("%s (clone)", orig.Name),
		Source:      RunSourceManual,
		CreatedBy:   createdBy,
		TestCaseIDs: tcIDs,
	})
}

func (s *Service) RerunRun(ctx context.Context, runID uuid.UUID, createdBy uuid.UUID) (*TestRun, error) {
	orig, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListItems(ctx, runID)
	if err != nil {
		return nil, err
	}

	tcIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.TestCaseID != nil {
			tcIDs = append(tcIDs, *item.TestCaseID)
		}
	}

	newRun, err := s.CreateRun(ctx, CreateRunInput{
		WorkspaceID: orig.WorkspaceID,
		ProjectID:   orig.ProjectID,
		SuiteID:     orig.SuiteID,
		Name:        fmt.Sprintf("%s (rerun)", orig.Name),
		Source:      RunSourceManual,
		CreatedBy:   createdBy,
		TestCaseIDs: tcIDs,
	})
	if err != nil {
		return nil, err
	}

	newItems, err := s.repo.ListItems(ctx, newRun.ID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for i := range newItems {
		if i < len(items) {
			newItems[i].StepResults = items[i].StepResults
			newItems[i].Comment = items[i].Comment
		}
		newItems[i].Status = RunItemStatusPending
		newItems[i].ExecutedBy = nil
		newItems[i].ExecutedAt = nil
		newItems[i].UpdatedAt = now
		if err := s.repo.UpdateItem(ctx, &newItems[i]); err != nil {
			return nil, err
		}
	}
	return newRun, nil
}

// Test plans

func (s *Service) CreatePlan(ctx context.Context, input CreatePlanInput) (*TestPlan, error) {
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	now := time.Now().UTC()
	plan := &TestPlan{
		ID:            uuid.New(),
		WorkspaceID:   input.WorkspaceID,
		ProjectID:     input.ProjectID,
		SuiteID:       input.SuiteID,
		Name:          input.Name,
		Description:   input.Description,
		Status:        TestPlanStatusActive,
		Configuration: input.Configuration,
		CreatedBy:     input.CreatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if plan.Configuration == nil {
		plan.Configuration = make(map[string]interface{})
	}

	if err := s.repo.RunInTx(ctx, func(tx Repository) error {
		if err := tx.CreatePlan(ctx, plan); err != nil {
			return err
		}
		for i, tcID := range input.TestCaseIDs {
			item := &TestPlanItem{
				ID:         uuid.New(),
				PlanID:     plan.ID,
				TestCaseID: tcID,
				SortOrder:  i,
				CreatedAt:  now,
			}
			if err := tx.CreatePlanItem(ctx, item); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) GetPlan(ctx context.Context, id uuid.UUID) (*TestPlan, error) {
	return s.repo.GetPlanByID(ctx, id)
}

func (s *Service) ListPlans(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]TestPlan, error) {
	return s.repo.ListPlans(ctx, projectID, cursor, limit)
}

func (s *Service) UpdatePlan(ctx context.Context, id uuid.UUID, input UpdatePlanInput) (*TestPlan, error) {
	plan, err := s.repo.GetPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		if !validation.IsValidName(input.Name) {
			return nil, sharederrors.ErrInvalidInput
		}
		plan.Name = input.Name
	}
	if input.Description != "" {
		plan.Description = input.Description
	}
	if input.Status != "" {
		if !IsValidTestPlanStatus(string(input.Status)) {
			return nil, sharederrors.ErrInvalidInput
		}
		plan.Status = input.Status
	}
	if input.Configuration != nil {
		plan.Configuration = input.Configuration
	}
	plan.UpdatedAt = time.Now().UTC()

	if err := s.repo.RunInTx(ctx, func(tx Repository) error {
		if err := tx.UpdatePlan(ctx, plan); err != nil {
			return err
		}
		if input.TestCaseIDs != nil {
			if err := tx.DeletePlanItemsByPlanID(ctx, plan.ID); err != nil {
				return err
			}
			for i, tcID := range input.TestCaseIDs {
				item := &TestPlanItem{
					ID:         uuid.New(),
					PlanID:     plan.ID,
					TestCaseID: tcID,
					SortOrder:  i,
					CreatedAt:  time.Now().UTC(),
				}
				if err := tx.CreatePlanItem(ctx, item); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) DeletePlan(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePlan(ctx, id)
}

func (s *Service) GetPlanItems(ctx context.Context, planID uuid.UUID) ([]TestPlanItem, error) {
	return s.repo.ListPlanItems(ctx, planID)
}

func (s *Service) CreateRunFromPlan(ctx context.Context, planID uuid.UUID, createdBy uuid.UUID) (*TestRun, error) {
	plan, err := s.repo.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListPlanItems(ctx, planID)
	if err != nil {
		return nil, err
	}
	tcIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		tcIDs = append(tcIDs, item.TestCaseID)
	}
	return s.CreateRun(ctx, CreateRunInput{
		WorkspaceID: plan.WorkspaceID,
		ProjectID:   plan.ProjectID,
		SuiteID:     plan.SuiteID,
		Name:        plan.Name,
		Source:      RunSourceManual,
		CreatedBy:   createdBy,
		TestCaseIDs: tcIDs,
		PlanID:      &plan.ID,
	})
}
