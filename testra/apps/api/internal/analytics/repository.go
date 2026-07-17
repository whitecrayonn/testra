package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type SQLRepository struct {
	db DBTX
}

func NewSQLRepository(sqlDB *sql.DB) *SQLRepository {
	return &SQLRepository{db: db.Wrap(sqlDB)}
}

func (r *SQLRepository) CreateDashboard(ctx context.Context, d *Dashboard) error {
	cfgJSON, _ := json.Marshal(d.Config)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO analytics_dashboards (id, workspace_id, name, type, config, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		d.ID, d.WorkspaceID, d.Name, d.Type, cfgJSON, d.CreatedBy, d.CreatedAt, d.UpdatedAt)
	return err
}

func (r *SQLRepository) ListDashboards(ctx context.Context, workspaceID uuid.UUID) ([]Dashboard, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, name, type, config::text, created_by, created_at, updated_at
		 FROM analytics_dashboards WHERE workspace_id = $1 ORDER BY created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDashboards(rows)
}

func (r *SQLRepository) GetDashboard(ctx context.Context, id uuid.UUID) (*Dashboard, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, name, type, config::text, created_by, created_at, updated_at
		 FROM analytics_dashboards WHERE id = $1`, id)
	return scanDashboard(row)
}

func (r *SQLRepository) UpdateDashboard(ctx context.Context, d *Dashboard) error {
	cfgJSON, _ := json.Marshal(d.Config)
	result, err := r.db.ExecContext(ctx,
		`UPDATE analytics_dashboards SET name = $2, type = $3, config = $4, updated_at = $5 WHERE id = $1`,
		d.ID, d.Name, d.Type, cfgJSON, d.UpdatedAt)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteDashboard(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM analytics_dashboards WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) UpsertDailyMetric(ctx context.Context, m *DailyMetric) error {
	projectID := uuid.NullUUID{}
	if m.ProjectID != nil {
		projectID = uuid.NullUUID{UUID: *m.ProjectID, Valid: true}
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO analytics_daily_metrics (id, workspace_id, project_id, metric_date, total_runs, passed, failed, skipped, blocked, duration_ms, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT (workspace_id, project_id, metric_date) DO UPDATE SET
		   total_runs = EXCLUDED.total_runs, passed = EXCLUDED.passed, failed = EXCLUDED.failed,
		   skipped = EXCLUDED.skipped, blocked = EXCLUDED.blocked, duration_ms = EXCLUDED.duration_ms,
		   updated_at = EXCLUDED.updated_at`,
		m.ID, m.WorkspaceID, projectID, m.MetricDate, m.TotalRuns, m.Passed, m.Failed, m.Skipped, m.Blocked, m.DurationMs, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *SQLRepository) AggregateDailyMetrics(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID, date time.Time) ([]DailyMetric, error) {
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)

	query := `SELECT project_id::text, COALESCE(SUM(total),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(duration_ms),0)
	          FROM test_runs WHERE workspace_id = $1 AND created_at >= $2 AND created_at < $3`
	args := []interface{}{workspaceID, start, end}
	if projectID != nil {
		query += ` AND project_id = $4`
		args = append(args, *projectID)
	}
	query += ` GROUP BY project_id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []DailyMetric
	for rows.Next() {
		var projectIDStr sql.NullString
		var m DailyMetric
		m.ID = uuid.New()
		m.WorkspaceID = workspaceID
		m.MetricDate = start
		m.CreatedAt = time.Now().UTC()
		m.UpdatedAt = m.CreatedAt
		if err := rows.Scan(&projectIDStr, &m.TotalRuns, &m.Passed, &m.Failed, &m.Skipped, &m.Blocked, &m.DurationMs); err != nil {
			return nil, err
		}
		if projectIDStr.Valid && projectIDStr.String != "" {
			pid, _ := uuid.Parse(projectIDStr.String)
			m.ProjectID = &pid
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

func (r *SQLRepository) GetDailyMetrics(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID, start, end time.Time) ([]DailyMetric, error) {
	var rows *sql.Rows
	var err error
	if projectID != nil {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, metric_date, total_runs, passed, failed, skipped, blocked, duration_ms, created_at, updated_at
		     FROM analytics_daily_metrics
		     WHERE workspace_id = $1 AND project_id = $2 AND metric_date BETWEEN $3 AND $4
		     ORDER BY metric_date ASC`,
			workspaceID, *projectID, start, end)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, metric_date, total_runs, passed, failed, skipped, blocked, duration_ms, created_at, updated_at
		     FROM analytics_daily_metrics
		     WHERE workspace_id = $1 AND metric_date BETWEEN $2 AND $3
		     ORDER BY metric_date ASC`,
			workspaceID, start, end)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDailyMetrics(rows)
}

func (r *SQLRepository) GetRunSummary(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID) (*Summary, error) {
	query := `SELECT COALESCE(SUM(total),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(duration_ms),0)
	          FROM test_runs WHERE workspace_id = $1`
	args := []interface{}{workspaceID}
	if projectID != nil {
		query += ` AND project_id = $2`
		args = append(args, *projectID)
	}
	s := &Summary{}
	row := r.db.QueryRowContext(ctx, query, args...)
	err := row.Scan(&s.TotalRuns, &s.Passed, &s.Failed, &s.Skipped, &s.Blocked, &s.DurationMs)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SQLRepository) AggregateTrends(ctx context.Context, workspaceID uuid.UUID, projectID *uuid.UUID, start, end time.Time) ([]TrendPoint, error) {
	// Prioritize materialized daily metrics, fall back to aggregating test_runs by completion date.
	metrics, err := r.GetDailyMetrics(ctx, workspaceID, projectID, start, end)
	if err != nil {
		return nil, err
	}
	if len(metrics) > 0 {
		return metricsToTrends(metrics), nil
	}

	query := `SELECT DATE(created_at), COALESCE(SUM(total),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(duration_ms),0)
	          FROM test_runs WHERE workspace_id = $1 AND created_at BETWEEN $2 AND $3`
	args := []interface{}{workspaceID, start, end}
	if projectID != nil {
		query += ` AND project_id = $4`
		args = append(args, *projectID)
	}
	query += ` GROUP BY DATE(created_at) ORDER BY DATE(created_at)`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []TrendPoint
	for rows.Next() {
		var date time.Time
		var t TrendPoint
		if err := rows.Scan(&date, &t.TotalRuns, &t.Passed, &t.Failed, &t.Skipped, &t.Blocked, &t.DurationMs); err != nil {
			return nil, err
		}
		t.Date = date.Format("2006-01-02")
		trends = append(trends, t)
	}
	return trends, rows.Err()
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanDashboard(row rowScanner) (*Dashboard, error) {
	var d Dashboard
	var createdBy sql.NullString
	var configStr string
	if err := row.Scan(&d.ID, &d.WorkspaceID, &d.Name, &d.Type, &configStr, &createdBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	if configStr != "" {
		_ = json.Unmarshal([]byte(configStr), &d.Config)
	}
	if d.Config == nil {
		d.Config = make(map[string]interface{})
	}
	if createdBy.Valid {
		d.CreatedBy, _ = uuid.Parse(createdBy.String)
	}
	return &d, nil
}

func scanDashboards(rows *sql.Rows) ([]Dashboard, error) {
	var list []Dashboard
	for rows.Next() {
		d, err := scanDashboard(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *d)
	}
	return list, rows.Err()
}

func scanDailyMetric(row rowScanner) (*DailyMetric, error) {
	var m DailyMetric
	var projectID sql.NullString
	if err := row.Scan(&m.ID, &m.WorkspaceID, &projectID, &m.MetricDate, &m.TotalRuns, &m.Passed, &m.Failed, &m.Skipped, &m.Blocked, &m.DurationMs, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	if projectID.Valid {
		pid, _ := uuid.Parse(projectID.String)
		m.ProjectID = &pid
	}
	return &m, nil
}

func scanDailyMetrics(rows *sql.Rows) ([]DailyMetric, error) {
	var list []DailyMetric
	for rows.Next() {
		m, err := scanDailyMetric(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *m)
	}
	return list, rows.Err()
}

func metricsToTrends(metrics []DailyMetric) []TrendPoint {
	trends := make([]TrendPoint, len(metrics))
	for i, m := range metrics {
		trends[i] = TrendPoint{
			Date:       m.MetricDate.Format("2006-01-02"),
			TotalRuns:  m.TotalRuns,
			Passed:     m.Passed,
			Failed:     m.Failed,
			Skipped:    m.Skipped,
			Blocked:    m.Blocked,
			DurationMs: m.DurationMs,
		}
	}
	return trends
}

func (r *SQLRepository) RunInTx(ctx context.Context, fn func(Repository) error) error {
	beginner, ok := r.db.(db.BeginTxer)
	if !ok {
		return fmt.Errorf("database handle does not support transactions")
	}
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if tenantID, ok := db.TenantIDFromContext(ctx); ok {
		_, _ = tx.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", tenantID.String())
	}
	txRepo := &SQLRepository{db: tx}
	if err := fn(txRepo); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
