package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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
	          FROM test_runs WHERE workspace_id = $1 AND COALESCE(completed_at, created_at) >= $2 AND COALESCE(completed_at, created_at) < $3`
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

	query := `SELECT DATE(COALESCE(completed_at, created_at)), COALESCE(SUM(total),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(duration_ms),0)
	          FROM test_runs WHERE workspace_id = $1 AND COALESCE(completed_at, created_at) BETWEEN $2 AND $3`
	args := []interface{}{workspaceID, start, end}
	if projectID != nil {
		query += ` AND project_id = $4`
		args = append(args, *projectID)
	}
	query += ` GROUP BY DATE(COALESCE(completed_at, created_at)) ORDER BY DATE(COALESCE(completed_at, created_at))`

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

func (r *SQLRepository) whereClause(args *[]interface{}, filter MetricsFilter, startIdx int) string {
	clauses := []string{fmt.Sprintf("workspace_id = $%d", startIdx)}
	*args = append(*args, filter.WorkspaceID)
	next := startIdx + 1

	if filter.ProjectID != nil {
		clauses = append(clauses, fmt.Sprintf("project_id = $%d", next))
		*args = append(*args, *filter.ProjectID)
		next++
	}
	if filter.Release != "" {
		clauses = append(clauses, fmt.Sprintf("metadata->>'release' = $%d", next))
		*args = append(*args, filter.Release)
		next++
	}
	if filter.Sprint != "" {
		clauses = append(clauses, fmt.Sprintf("metadata->>'sprint' = $%d", next))
		*args = append(*args, filter.Sprint)
		next++
	}
	if filter.Environment != "" {
		clauses = append(clauses, fmt.Sprintf("metadata->>'environment' = $%d", next))
		*args = append(*args, filter.Environment)
		next++
	}
	if filter.TesterID != nil {
		clauses = append(clauses, fmt.Sprintf("created_by = $%d", next))
		*args = append(*args, *filter.TesterID)
		next++
	}
	if filter.Source != "" {
		clauses = append(clauses, fmt.Sprintf("source = $%d", next))
		*args = append(*args, filter.Source)
		next++
	}
	if filter.Start != nil {
		clauses = append(clauses, fmt.Sprintf("COALESCE(completed_at, created_at) >= $%d", next))
		*args = append(*args, *filter.Start)
		next++
	}
	if filter.End != nil {
		clauses = append(clauses, fmt.Sprintf("COALESCE(completed_at, created_at) <= $%d", next))
		*args = append(*args, *filter.End)
		next++
	}
	return "WHERE " + strings.Join(clauses, " AND ")
}

func (r *SQLRepository) countWhere(args *[]interface{}, filter MetricsFilter, startIdx int) string {
	clauses := []string{fmt.Sprintf("r.workspace_id = $%d", startIdx)}
	*args = append(*args, filter.WorkspaceID)
	next := startIdx + 1

	if filter.ProjectID != nil {
		clauses = append(clauses, fmt.Sprintf("r.project_id = $%d", next))
		*args = append(*args, *filter.ProjectID)
		next++
	}
	if filter.Release != "" {
		clauses = append(clauses, fmt.Sprintf("r.metadata->>'release' = $%d", next))
		*args = append(*args, filter.Release)
		next++
	}
	if filter.Sprint != "" {
		clauses = append(clauses, fmt.Sprintf("r.metadata->>'sprint' = $%d", next))
		*args = append(*args, filter.Sprint)
		next++
	}
	if filter.Environment != "" {
		clauses = append(clauses, fmt.Sprintf("r.metadata->>'environment' = $%d", next))
		*args = append(*args, filter.Environment)
		next++
	}
	if filter.TesterID != nil {
		clauses = append(clauses, fmt.Sprintf("r.created_by = $%d", next))
		*args = append(*args, *filter.TesterID)
		next++
	}
	if filter.Source != "" {
		clauses = append(clauses, fmt.Sprintf("r.source = $%d", next))
		*args = append(*args, filter.Source)
		next++
	}
	if filter.Start != nil {
		clauses = append(clauses, fmt.Sprintf("COALESCE(r.completed_at, r.created_at) >= $%d", next))
		*args = append(*args, *filter.Start)
		next++
	}
	if filter.End != nil {
		clauses = append(clauses, fmt.Sprintf("COALESCE(r.completed_at, r.created_at) <= $%d", next))
		*args = append(*args, *filter.End)
		next++
	}
	return "WHERE " + strings.Join(clauses, " AND ")
}

func (r *SQLRepository) itemFilter(args *[]interface{}, filter MetricsFilter, startIdx int, alias string) (string, int) {
	clauses := []string{fmt.Sprintf("%s.workspace_id = $%d", alias, startIdx)}
	*args = append(*args, filter.WorkspaceID)
	next := startIdx + 1

	if filter.ProjectID != nil {
		clauses = append(clauses, fmt.Sprintf("%s.project_id = $%d", alias, next))
		*args = append(*args, *filter.ProjectID)
		next++
	}
	if filter.Release != "" {
		clauses = append(clauses, fmt.Sprintf("%s.metadata->>'release' = $%d", alias, next))
		*args = append(*args, filter.Release)
		next++
	}
	if filter.Sprint != "" {
		clauses = append(clauses, fmt.Sprintf("%s.metadata->>'sprint' = $%d", alias, next))
		*args = append(*args, filter.Sprint)
		next++
	}
	if filter.Environment != "" {
		clauses = append(clauses, fmt.Sprintf("%s.metadata->>'environment' = $%d", alias, next))
		*args = append(*args, filter.Environment)
		next++
	}
	if filter.Source != "" {
		clauses = append(clauses, fmt.Sprintf("%s.source = $%d", alias, next))
		*args = append(*args, filter.Source)
		next++
	}
	if filter.Start != nil {
		clauses = append(clauses, fmt.Sprintf("COALESCE(%s.completed_at, %s.created_at) >= $%d", alias, alias, next))
		*args = append(*args, *filter.Start)
		next++
	}
	if filter.End != nil {
		clauses = append(clauses, fmt.Sprintf("COALESCE(%s.completed_at, %s.created_at) <= $%d", alias, alias, next))
		*args = append(*args, *filter.End)
		next++
	}
	return "WHERE " + strings.Join(clauses, " AND "), next
}

func (r *SQLRepository) GetMetrics(ctx context.Context, filter MetricsFilter) (*Metrics, error) {
	m := &Metrics{
		TopFailedTestCases:   make([]TopFailedItem, 0),
		TopFailedSuites:      make([]TopFailedSuite, 0),
		TopFailedAPIs:        make([]TopFailedAPI, 0),
		MostActiveQA:         make([]ActiveUser, 0),
		MostActiveAutomation: make([]ActiveUser, 0),
		RecentActivity:       make([]Activity, 0),
		ExecutionTimeline:    make([]TimelinePoint, 0),
		WeeklyTrend:          make([]TrendPoint, 0),
		MonthlyTrend:         make([]TrendPoint, 0),
		ReleaseQualityTrend:  make([]ReleaseQualityPoint, 0),
	}

	// Total test cases, plans, runs.
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM test_cases WHERE workspace_id = $1`, filter.WorkspaceID).Scan(&m.TotalTestCases)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM test_plans WHERE workspace_id = $1`, filter.WorkspaceID).Scan(&m.TotalTestPlans)

	runArgs := []interface{}{}
	runWhere := r.whereClause(&runArgs, filter, 1)
	runQuery := fmt.Sprintf(`SELECT COUNT(*), COALESCE(SUM(total),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(duration_ms),0) FROM test_runs %s`, runWhere)
	row := r.db.QueryRowContext(ctx, runQuery, runArgs...)
	var totalRuns int64
	var totalItems, passed, failed, skipped, blocked, duration int64
	if err := row.Scan(&totalRuns, &totalItems, &passed, &failed, &skipped, &blocked, &duration); err != nil {
		return nil, err
	}
	m.TotalTestRuns = totalRuns
	m.ExecutionDurationMs = duration
	if totalRuns > 0 {
		m.AverageExecutionTimeMs = duration / totalRuns
	}
	if totalItems > 0 {
		m.ExecutionProgress = float64(passed+failed+skipped+blocked) / float64(totalItems)
		m.PassRate = float64(passed) / float64(totalItems)
		m.FailRate = float64(failed) / float64(totalItems)
	}
	m.Blocked = blocked
	m.Skipped = skipped

	// Retest / not_executed from run items.
	itemArgs := []interface{}{}
	itemWhere, _ := r.itemFilter(&itemArgs, filter, 1, "r")
	retestQuery := fmt.Sprintf(`SELECT COUNT(*) FROM test_run_items i JOIN test_runs r ON i.run_id = r.id %s AND i.status = 'retest'`, itemWhere)
	notExecQuery := fmt.Sprintf(`SELECT COUNT(*) FROM test_run_items i JOIN test_runs r ON i.run_id = r.id %s AND i.status = 'not_executed'`, itemWhere)
	var retest, notExec int64
	_ = r.db.QueryRowContext(ctx, retestQuery, itemArgs...).Scan(&retest)
	_ = r.db.QueryRowContext(ctx, notExecQuery, itemArgs...).Scan(&notExec)
	m.Retest = retest

	// Coverage based on run source.
	var automationRuns, apiRuns, manualRuns int64
	autoArgs := []interface{}{filter.WorkspaceID}
	autoWhere := "workspace_id = $1"
	next := 2
	if filter.ProjectID != nil {
		autoWhere += fmt.Sprintf(" AND project_id = $%d", next)
		autoArgs = append(autoArgs, *filter.ProjectID)
		next++
	}
	if filter.Start != nil {
		autoWhere += fmt.Sprintf(" AND COALESCE(completed_at, created_at) >= $%d", next)
		autoArgs = append(autoArgs, *filter.Start)
		next++
	}
	if filter.End != nil {
		autoWhere += fmt.Sprintf(" AND COALESCE(completed_at, created_at) <= $%d", next)
		autoArgs = append(autoArgs, *filter.End)
		next++
	}
	if filter.Source == "" {
		_ = r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM test_runs WHERE %s AND source = 'automation'", autoWhere), autoArgs...).Scan(&automationRuns)
		_ = r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM test_runs WHERE %s AND source = 'api'", autoWhere), autoArgs...).Scan(&apiRuns)
		_ = r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM test_runs WHERE %s AND source = 'manual'", autoWhere), autoArgs...).Scan(&manualRuns)
	} else {
		_ = r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM test_runs WHERE %s", autoWhere), autoArgs...).Scan(&totalRuns)
	}

	if totalRuns > 0 {
		m.AutomationCoverage = float64(automationRuns) / float64(totalRuns) * 100
		m.APITestCoverage = float64(apiRuns) / float64(totalRuns) * 100
	}

	if totalItems == 1 && retest == 0 && notExec == 0 {
		// no-op to avoid unused variable warning in older Go versions
		_ = notExec
	}

	// Defects.
	open, closed, density, aging, reopenRate, err := r.GetDefectMetrics(ctx, filter)
	if err != nil {
		return nil, err
	}
	m.OpenDefects = open
	m.ClosedDefects = closed
	m.DefectDensity = density
	m.DefectAging = aging
	m.BugReopenRate = reopenRate

	// Top lists and active users.
	m.TopFailedTestCases, _ = r.GetTopFailedTestCases(ctx, filter)
	m.TopFailedSuites, _ = r.GetTopFailedSuites(ctx, filter)
	m.TopFailedAPIs, _ = r.GetTopFailedAPIs(ctx, filter)
	m.MostActiveQA, _ = r.GetMostActiveQA(ctx, filter)
	m.MostActiveAutomation, _ = r.GetMostActiveAutomation(ctx, filter)
	m.RecentActivity, _ = r.GetRecentActivity(ctx, filter)
	m.ExecutionTimeline, _ = r.GetExecutionTimeline(ctx, filter)
	m.WeeklyTrend, _ = r.GetWeeklyTrend(ctx, filter)
	m.MonthlyTrend, _ = r.GetMonthlyTrend(ctx, filter)
	m.ReleaseQualityTrend, _ = r.GetReleaseQualityTrend(ctx, filter)

	return m, nil
}

func (r *SQLRepository) GetTopFailedTestCases(ctx context.Context, filter MetricsFilter) ([]TopFailedItem, error) {
	limit := filter.Limit
	if limit == 0 {
		limit = 10
	}
	args := []interface{}{}
	where, _ := r.itemFilter(&args, filter, 1, "r")
	q := fmt.Sprintf(`SELECT i.test_case_id::text, t.title, COUNT(*) AS failures
		FROM test_run_items i
		JOIN test_runs r ON i.run_id = r.id
		LEFT JOIN test_cases t ON t.id = i.test_case_id
		%s AND i.status = 'failed'
		GROUP BY i.test_case_id, t.title
		ORDER BY failures DESC
		LIMIT %d`, where, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TopFailedItem
	for rows.Next() {
		var it TopFailedItem
		var title sql.NullString
		if err := rows.Scan(&it.TestCaseID, &title, &it.Failures); err != nil {
			return nil, err
		}
		it.Title = title.String
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *SQLRepository) GetTopFailedSuites(ctx context.Context, filter MetricsFilter) ([]TopFailedSuite, error) {
	limit := filter.Limit
	if limit == 0 {
		limit = 10
	}
	args := []interface{}{}
	where, _ := r.itemFilter(&args, filter, 1, "r")
	q := fmt.Sprintf(`SELECT t.suite_id::text, s.name, COUNT(*) AS failures
		FROM test_run_items i
		JOIN test_runs r ON i.run_id = r.id
		JOIN test_cases t ON t.id = i.test_case_id
		LEFT JOIN test_suites s ON s.id = t.suite_id
		%s AND i.status = 'failed'
		GROUP BY t.suite_id, s.name
		ORDER BY failures DESC
		LIMIT %d`, where, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TopFailedSuite
	for rows.Next() {
		var it TopFailedSuite
		var name sql.NullString
		if err := rows.Scan(&it.SuiteID, &name, &it.Failures); err != nil {
			return nil, err
		}
		it.Name = name.String
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *SQLRepository) GetTopFailedAPIs(ctx context.Context, filter MetricsFilter) ([]TopFailedAPI, error) {
	limit := filter.Limit
	if limit == 0 {
		limit = 10
	}
	args := []interface{}{filter.WorkspaceID}
	where := "WHERE workspace_id = $1"
	next := 2
	if filter.ProjectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", next)
		args = append(args, *filter.ProjectID)
		next++
	}
	if filter.Start != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", next)
		args = append(args, *filter.Start)
		next++
	}
	if filter.End != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", next)
		args = append(args, *filter.End)
		next++
	}
	q := fmt.Sprintf(`SELECT request_id::text, name, COUNT(*) AS failures
		FROM api_request_history
		%s AND COALESCE(response_status, 0) >= 400
		GROUP BY request_id, name
		ORDER BY failures DESC
		LIMIT %d`, where, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TopFailedAPI
	for rows.Next() {
		var it TopFailedAPI
		if err := rows.Scan(&it.RequestID, &it.Name, &it.Failures); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *SQLRepository) GetMostActiveQA(ctx context.Context, filter MetricsFilter) ([]ActiveUser, error) {
	limit := filter.Limit
	if limit == 0 {
		limit = 10
	}
	args := []interface{}{}
	where, _ := r.itemFilter(&args, filter, 1, "r")
	q := fmt.Sprintf(`SELECT u.id::text, COALESCE(u.name, u.email), COUNT(*) AS executions
		FROM test_run_items i
		JOIN test_runs r ON i.run_id = r.id
		JOIN users u ON u.id = i.executed_by
		%s AND i.executed_by IS NOT NULL
		GROUP BY u.id, u.name, u.email
		ORDER BY executions DESC
		LIMIT %d`, where, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []ActiveUser
	for rows.Next() {
		var u ActiveUser
		if err := rows.Scan(&u.UserID, &u.Name, &u.Count); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *SQLRepository) GetMostActiveAutomation(ctx context.Context, filter MetricsFilter) ([]ActiveUser, error) {
	limit := filter.Limit
	if limit == 0 {
		limit = 10
	}
	args := []interface{}{filter.WorkspaceID}
	where := "WHERE workspace_id = $1"
	next := 2
	if filter.ProjectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", next)
		args = append(args, *filter.ProjectID)
		next++
	}
	if filter.Start != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", next)
		args = append(args, *filter.Start)
		next++
	}
	if filter.End != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", next)
		args = append(args, *filter.End)
		next++
	}
	q := fmt.Sprintf(`SELECT COALESCE(u.id::text, ''), COALESCE(u.name, u.email), COUNT(*) AS runs
		FROM automation_executions e
		LEFT JOIN users u ON u.id = e.created_by
		%s
		GROUP BY u.id, u.name, u.email
		ORDER BY runs DESC
		LIMIT %d`, where, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []ActiveUser
	for rows.Next() {
		var u ActiveUser
		if err := rows.Scan(&u.UserID, &u.Name, &u.Count); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *SQLRepository) GetDefectMetrics(ctx context.Context, filter MetricsFilter) (open, closed int64, density float64, aging DefectAging, reopenRate float64, err error) {
	args := []interface{}{filter.WorkspaceID}
	where := "WHERE workspace_id = $1"
	next := 2
	if filter.ProjectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", next)
		args = append(args, *filter.ProjectID)
		next++
	}
	if filter.Start != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", next)
		args = append(args, *filter.Start)
		next++
	}
	if filter.End != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", next)
		args = append(args, *filter.End)
		next++
	}

	err = r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FILTER (WHERE status = 'open'), COUNT(*) FILTER (WHERE status <> 'open') FROM defects %s", where), args...).Scan(&open, &closed)
	if err != nil {
		return
	}

	var testCases int64
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_cases WHERE workspace_id = $1", filter.WorkspaceID).Scan(&testCases)
	if testCases > 0 {
		density = float64(open) / float64(testCases) * 100
	}

	err = r.db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - created_at))/86400),0), COALESCE(MAX(EXTRACT(EPOCH FROM (NOW() - created_at))/86400),0)
		FROM defects %s AND status = 'open'`, where), args...).Scan(&aging.AverageDays, &aging.MaxDays)
	if err != nil {
		return
	}

	// Reopen rate: defects linked to multiple run items / total distinct defects with links.
	var reopens, linked int64
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM (
		SELECT defect_id FROM test_run_item_defects
		GROUP BY defect_id HAVING COUNT(*) > 1
	) x`).Scan(&reopens)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT defect_id) FROM test_run_item_defects`).Scan(&linked)
	if linked > 0 {
		reopenRate = float64(reopens) / float64(linked) * 100
	}
	return
}

func (r *SQLRepository) GetRecentActivity(ctx context.Context, filter MetricsFilter) ([]Activity, error) {
	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	projectFilter := ""
	args := []interface{}{filter.WorkspaceID, limit}
	if filter.ProjectID != nil {
		projectFilter = "AND project_id = $3"
		args = append(args, *filter.ProjectID)
	}
	q := fmt.Sprintf(`SELECT id::text, 'run', name, COALESCE(created_by::text, ''), created_at FROM test_runs WHERE workspace_id = $1 %s
		UNION ALL
		SELECT id::text, 'defect', title, COALESCE(created_by::text, ''), created_at FROM defects WHERE workspace_id = $1 %s
		UNION ALL
		SELECT id::text, 'test_case', title, COALESCE(created_by::text, ''), created_at FROM test_cases WHERE workspace_id = $1 %s
		UNION ALL
		SELECT id::text, 'api_request', name, COALESCE(created_by::text, ''), created_at FROM api_request_history WHERE workspace_id = $1 %s
		UNION ALL
		SELECT id::text, 'automation', name, COALESCE(created_by::text, ''), created_at FROM automation_executions WHERE workspace_id = $1 %s
		ORDER BY created_at DESC
		LIMIT $2`, projectFilter, projectFilter, projectFilter, projectFilter, projectFilter)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var activity []Activity
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.ID, &a.Type, &a.Title, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		activity = append(activity, a)
	}
	return activity, rows.Err()
}

func (r *SQLRepository) GetExecutionTimeline(ctx context.Context, filter MetricsFilter) ([]TimelinePoint, error) {
	args := []interface{}{}
	where := r.whereClause(&args, filter, 1)
	q := fmt.Sprintf(`SELECT DATE(COALESCE(completed_at, created_at)), COALESCE(SUM(total),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(duration_ms),0)
		FROM test_runs %s
		GROUP BY DATE(COALESCE(completed_at, created_at))
		ORDER BY DATE(COALESCE(completed_at, created_at)) DESC
		LIMIT 30`, where)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []TimelinePoint
	for rows.Next() {
		var p TimelinePoint
		var date time.Time
		if err := rows.Scan(&date, &p.TotalRuns, &p.Passed, &p.Failed, &p.Skipped, &p.Blocked, &p.DurationMs); err != nil {
			return nil, err
		}
		p.Date = date.Format("2006-01-02")
		points = append(points, p)
	}
	return points, rows.Err()
}

func (r *SQLRepository) GetWeeklyTrend(ctx context.Context, filter MetricsFilter) ([]TrendPoint, error) {
	args := []interface{}{}
	where := r.whereClause(&args, filter, 1)
	q := fmt.Sprintf(`SELECT DATE_TRUNC('week', COALESCE(completed_at, created_at))::date, COALESCE(SUM(total),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(duration_ms),0)
		FROM test_runs %s
		GROUP BY DATE_TRUNC('week', COALESCE(completed_at, created_at))::date
		ORDER BY DATE_TRUNC('week', COALESCE(completed_at, created_at))::date DESC
		LIMIT 12`, where)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTrendRows(rows)
}

func (r *SQLRepository) GetMonthlyTrend(ctx context.Context, filter MetricsFilter) ([]TrendPoint, error) {
	args := []interface{}{}
	where := r.whereClause(&args, filter, 1)
	q := fmt.Sprintf(`SELECT DATE_TRUNC('month', COALESCE(completed_at, created_at))::date, COALESCE(SUM(total),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(duration_ms),0)
		FROM test_runs %s
		GROUP BY DATE_TRUNC('month', COALESCE(completed_at, created_at))::date
		ORDER BY DATE_TRUNC('month', COALESCE(completed_at, created_at))::date DESC
		LIMIT 12`, where)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTrendRows(rows)
}

func scanTrendRows(rows *sql.Rows) ([]TrendPoint, error) {
	var points []TrendPoint
	for rows.Next() {
		var p TrendPoint
		var date time.Time
		if err := rows.Scan(&date, &p.TotalRuns, &p.Passed, &p.Failed, &p.Skipped, &p.Blocked, &p.DurationMs); err != nil {
			return nil, err
		}
		p.Date = date.Format("2006-01-02")
		points = append(points, p)
	}
	return points, rows.Err()
}

func (r *SQLRepository) GetReleaseQualityTrend(ctx context.Context, filter MetricsFilter) ([]ReleaseQualityPoint, error) {
	args := []interface{}{}
	where := r.whereClause(&args, filter, 1)
	q := fmt.Sprintf(`SELECT COALESCE(metadata->>'release','N/A'), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0), COALESCE(SUM(skipped),0), COALESCE(SUM(blocked),0), COALESCE(SUM(total),0)
		FROM test_runs %s
		GROUP BY COALESCE(metadata->>'release','N/A')
		ORDER BY COALESCE(metadata->>'release','N/A')`, where)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []ReleaseQualityPoint
	for rows.Next() {
		var p ReleaseQualityPoint
		if err := rows.Scan(&p.Release, &p.Passed, &p.Failed, &p.Skipped, &p.Blocked, &p.Total); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (r *SQLRepository) GetMetricsCSV(ctx context.Context, filter MetricsFilter) ([][]string, error) {
	m, err := r.GetMetrics(ctx, filter)
	if err != nil {
		return nil, err
	}
	rows := [][]string{
		{"Metric", "Value"},
		{"Total Test Cases", fmt.Sprintf("%d", m.TotalTestCases)},
		{"Total Test Plans", fmt.Sprintf("%d", m.TotalTestPlans)},
		{"Total Test Runs", fmt.Sprintf("%d", m.TotalTestRuns)},
		{"Execution Progress", fmt.Sprintf("%.2f", m.ExecutionProgress*100)},
		{"Pass Rate", fmt.Sprintf("%.2f", m.PassRate*100)},
		{"Fail Rate", fmt.Sprintf("%.2f", m.FailRate*100)},
		{"Blocked", fmt.Sprintf("%d", m.Blocked)},
		{"Retest", fmt.Sprintf("%d", m.Retest)},
		{"Skipped", fmt.Sprintf("%d", m.Skipped)},
		{"Automation Coverage", fmt.Sprintf("%.2f", m.AutomationCoverage)},
		{"API Test Coverage", fmt.Sprintf("%.2f", m.APITestCoverage)},
		{"Average Execution Time (ms)", fmt.Sprintf("%d", m.AverageExecutionTimeMs)},
		{"Defect Density", fmt.Sprintf("%.2f", m.DefectDensity)},
		{"Open Defects", fmt.Sprintf("%d", m.OpenDefects)},
		{"Closed Defects", fmt.Sprintf("%d", m.ClosedDefects)},
		{"Bug Reopen Rate", fmt.Sprintf("%.2f", m.BugReopenRate)},
	}
	rows = append(rows, []string{"Top Failed Test Cases"})
	for _, it := range m.TopFailedTestCases {
		rows = append(rows, []string{it.Title, fmt.Sprintf("%d", it.Failures)})
	}
	return rows, nil
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
