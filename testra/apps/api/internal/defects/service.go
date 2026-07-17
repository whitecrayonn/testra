package defects

import (
	"context"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	WorkspaceID   uuid.UUID
	ProjectID     uuid.UUID
	TestRunItemID *uuid.UUID
	Title         string
	Description   string
	Severity      DefectSeverity
	Priority      DefectPriority
	Status        DefectStatus
	AssignedTo    *uuid.UUID
	CreatedBy     uuid.UUID
}

type UpdateInput struct {
	Title         *string
	Description   *string
	Severity      *DefectSeverity
	Priority      *DefectPriority
	Status        *DefectStatus
	AssignedTo    *uuid.UUID
	TestRunItemID *uuid.UUID
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Defect, error) {
	if input.WorkspaceID == uuid.Nil || input.ProjectID == uuid.Nil || input.Title == "" || input.CreatedBy == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	severity := input.Severity
	if severity == "" {
		severity = DefectSeverityMedium
	}
	priority := input.Priority
	if priority == "" {
		priority = DefectPriorityMedium
	}
	status := input.Status
	if status == "" {
		status = DefectStatusOpen
	}

	if !IsValidSeverity(severity) || !IsValidPriority(priority) || !IsValidStatus(status) {
		return nil, sharederrors.ErrInvalidInput
	}

	now := time.Now().UTC()
	d := &Defect{
		ID:            uuid.New(),
		WorkspaceID:   input.WorkspaceID,
		ProjectID:     input.ProjectID,
		TestRunItemID: input.TestRunItemID,
		Title:         input.Title,
		Description:   input.Description,
		Severity:      severity,
		Priority:      priority,
		Status:        status,
		AssignedTo:    input.AssignedTo,
		CreatedBy:     input.CreatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Defect, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]Defect, error) {
	if projectID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListByProject(ctx, projectID, cursor, limit)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*Defect, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		if *input.Title == "" {
			return nil, sharederrors.ErrInvalidInput
		}
		d.Title = *input.Title
	}
	if input.Description != nil {
		d.Description = *input.Description
	}
	if input.Severity != nil {
		if !IsValidSeverity(*input.Severity) {
			return nil, sharederrors.ErrInvalidInput
		}
		d.Severity = *input.Severity
	}
	if input.Priority != nil {
		if !IsValidPriority(*input.Priority) {
			return nil, sharederrors.ErrInvalidInput
		}
		d.Priority = *input.Priority
	}
	if input.Status != nil {
		if !IsValidStatus(*input.Status) {
			return nil, sharederrors.ErrInvalidInput
		}
		d.Status = *input.Status
	}
	if input.AssignedTo != nil {
		d.AssignedTo = input.AssignedTo
	}
	if input.TestRunItemID != nil {
		d.TestRunItemID = input.TestRunItemID
	}

	d.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}
