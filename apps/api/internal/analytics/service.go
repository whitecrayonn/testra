package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDashboard(ctx context.Context, input CreateDashboardInput, createdBy uuid.UUID) (*Dashboard, error) {
	if input.WorkspaceID == uuid.Nil || !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}
	now := time.Now().UTC()
	d := &Dashboard{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Type:        input.Type,
		Config:      input.Config,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if d.Type == "" {
		d.Type = "custom"
	}
	if d.Config == nil {
		d.Config = make(map[string]interface{})
	}
	if err := s.repo.CreateDashboard(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) ListDashboards(ctx context.Context, workspaceID uuid.UUID) ([]Dashboard, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListDashboards(ctx, workspaceID)
}

func (s *Service) GetDashboard(ctx context.Context, id uuid.UUID) (*Dashboard, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetDashboard(ctx, id)
}

func (s *Service) UpdateDashboard(ctx context.Context, id uuid.UUID, input UpdateDashboardInput) (*Dashboard, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	d, err := s.repo.GetDashboard(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		d.Name = input.Name
	}
	if input.Type != "" {
		d.Type = input.Type
	}
	if input.Config != nil {
		d.Config = input.Config
	}
	d.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateDashboard(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) DeleteDashboard(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteDashboard(ctx, id)
}

func (s *Service) GetSummary(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) (*Summary, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetRunSummary(ctx, workspaceID, projectID)
}

func (s *Service) GetTrends(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID, start, end *time.Time) ([]TrendPoint, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	sd := time.Now().UTC().AddDate(0, 0, -30).Truncate(24 * time.Hour)
	ed := time.Now().UTC().Truncate(24 * time.Hour)
	if start != nil {
		sd = *start
	}
	if end != nil {
		ed = *end
	}
	return s.repo.AggregateTrends(ctx, workspaceID, projectID, sd, ed)
}

func (s *Service) AggregateMetrics(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) error {
	if workspaceID == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}

	yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	metrics, err := s.repo.AggregateDailyMetrics(ctx, workspaceID, projectID, yesterday)
	if err != nil {
		return err
	}

	var total DailyMetric
	total.ID = uuid.New()
	total.WorkspaceID = workspaceID
	total.MetricDate = yesterday
	total.CreatedAt = time.Now().UTC()
	total.UpdatedAt = total.CreatedAt

	for _, m := range metrics {
		total.TotalRuns += m.TotalRuns
		total.Passed += m.Passed
		total.Failed += m.Failed
		total.Skipped += m.Skipped
		total.Blocked += m.Blocked
		total.DurationMs += m.DurationMs
		if err := s.repo.UpsertDailyMetric(ctx, &m); err != nil {
			return err
		}
	}

	if err := s.repo.UpsertDailyMetric(ctx, &total); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetMetrics(ctx context.Context, filter MetricsFilter) (*Metrics, error) {
	if filter.WorkspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	return s.repo.GetMetrics(ctx, filter)
}

func (s *Service) GetRecentActivity(ctx context.Context, filter MetricsFilter) ([]Activity, error) {
	if filter.WorkspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	return s.repo.GetRecentActivity(ctx, filter)
}

func (s *Service) GetMetricsCSV(ctx context.Context, filter MetricsFilter) ([][]string, error) {
	if filter.WorkspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	return s.repo.GetMetricsCSV(ctx, filter)
}

// Input structs

type CreateDashboardInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Type        string
	Config      map[string]interface{}
}

type UpdateDashboardInput struct {
	Name   string
	Type   string
	Config map[string]interface{}
}
