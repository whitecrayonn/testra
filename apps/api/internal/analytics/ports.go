package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MetricsFilter struct {
	WorkspaceID uuid.UUID
	ProjectID   *uuid.UUID
	Release     string
	Sprint      string
	Environment string
	TesterID    *uuid.UUID
	Source      string
	Start       *time.Time
	End         *time.Time
	Limit       int
}

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

	GetMetrics(ctx context.Context, filter MetricsFilter) (*Metrics, error)
	GetTopFailedTestCases(ctx context.Context, filter MetricsFilter) ([]TopFailedItem, error)
	GetTopFailedSuites(ctx context.Context, filter MetricsFilter) ([]TopFailedSuite, error)
	GetTopFailedAPIs(ctx context.Context, filter MetricsFilter) ([]TopFailedAPI, error)
	GetMostActiveQA(ctx context.Context, filter MetricsFilter) ([]ActiveUser, error)
	GetMostActiveAutomation(ctx context.Context, filter MetricsFilter) ([]ActiveUser, error)
	GetDefectMetrics(ctx context.Context, filter MetricsFilter) (open, closed int64, density float64, aging DefectAging, reopenRate float64, err error)
	GetRecentActivity(ctx context.Context, filter MetricsFilter) ([]Activity, error)
	GetExecutionTimeline(ctx context.Context, filter MetricsFilter) ([]TimelinePoint, error)
	GetWeeklyTrend(ctx context.Context, filter MetricsFilter) ([]TrendPoint, error)
	GetMonthlyTrend(ctx context.Context, filter MetricsFilter) ([]TrendPoint, error)
	GetReleaseQualityTrend(ctx context.Context, filter MetricsFilter) ([]ReleaseQualityPoint, error)
	GetMetricsCSV(ctx context.Context, filter MetricsFilter) ([][]string, error)
}
