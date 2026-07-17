package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateDashboard(ctx context.Context, d *Dashboard) error
	ListDashboards(ctx context.Context, workspaceID uuid.UUID) ([]Dashboard, error)
	GetDashboard(ctx context.Context, id uuid.UUID) (*Dashboard, error)
	UpdateDashboard(ctx context.Context, d *Dashboard) error
	DeleteDashboard(ctx context.Context, id uuid.UUID) error

	UpsertDailyMetric(ctx context.Context, m *DailyMetric) error
	GetDailyMetrics(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID, start, end time.Time) ([]DailyMetric, error)
	AggregateDailyMetrics(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID, date time.Time) ([]DailyMetric, error)
	GetRunSummary(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) (*Summary, error)
	AggregateTrends(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID, start, end time.Time) ([]TrendPoint, error)
}
