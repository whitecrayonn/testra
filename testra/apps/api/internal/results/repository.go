package results

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/pagination"
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

func (r *SQLRepository) CreateRun(ctx context.Context, run *TestRun) error {
	metaJSON, _ := json.Marshal(run.Metadata)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_runs (id, workspace_id, project_id, suite_id, name, status, total, passed, failed, skipped, blocked, duration_ms, source, metadata, created_by, started_at, completed_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		run.ID, run.WorkspaceID, run.ProjectID, run.SuiteID, run.Name, string(run.Status),
		run.Total, run.Passed, run.Failed, run.Skipped, run.Blocked, run.DurationMs,
		string(run.Source), metaJSON, run.CreatedBy, run.StartedAt, run.CompletedAt,
		run.CreatedAt, run.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetRunByID(ctx context.Context, id uuid.UUID) (*TestRun, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, project_id, suite_id, name, status, total, passed, failed, skipped, blocked, duration_ms, source, metadata::text, created_by, started_at, completed_at, created_at, updated_at
		 FROM test_runs WHERE id = $1`, id)
	return scanRun(row)
}

func (r *SQLRepository) ListRuns(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]TestRun, error) {
	if cursor == "" {
		rows, err := r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, suite_id, name, status, total, passed, failed, skipped, blocked, duration_ms, source, metadata::text, created_by, started_at, completed_at, created_at, updated_at
			 FROM test_runs WHERE project_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2`,
			projectID, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanRuns(rows)
	}

	cursorID, err := pagination.DecodeCursor(cursor)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, project_id, suite_id, name, status, total, passed, failed, skipped, blocked, duration_ms, source, metadata::text, created_by, started_at, completed_at, created_at, updated_at
		 FROM test_runs WHERE project_id = $1 AND id < $2::uuid ORDER BY created_at DESC, id DESC LIMIT $3`,
		projectID, cursorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRuns(rows)
}

func (r *SQLRepository) UpdateRun(ctx context.Context, run *TestRun) error {
	metaJSON, _ := json.Marshal(run.Metadata)
	result, err := r.db.ExecContext(ctx,
		`UPDATE test_runs SET name = $2, status = $3, total = $4, passed = $5, failed = $6, skipped = $7, blocked = $8, duration_ms = $9, suite_id = $10, metadata = $11, started_at = $12, completed_at = $13, updated_at = $14
		 WHERE id = $1`,
		run.ID, run.Name, string(run.Status), run.Total, run.Passed, run.Failed, run.Skipped, run.Blocked,
		run.DurationMs, run.SuiteID, metaJSON, run.StartedAt, run.CompletedAt, run.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteRun(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM test_runs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) CreateItem(ctx context.Context, item *TestRunItem) error {
	artifactsJSON, _ := json.Marshal(item.Artifacts)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_run_items (id, run_id, test_case_id, title, status, duration_ms, error_message, stack_trace, artifacts, sort_order, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		item.ID, item.RunID, item.TestCaseID, item.Title, string(item.Status),
		item.DurationMs, item.ErrorMessage, item.StackTrace, artifactsJSON,
		item.SortOrder, item.CreatedAt, item.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*TestRunItem, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, run_id, test_case_id, title, status, duration_ms, error_message, stack_trace, artifacts::text, sort_order, created_at, updated_at
		 FROM test_run_items WHERE id = $1`, id)
	return scanItem(row)
}

func (r *SQLRepository) ListItems(ctx context.Context, runID uuid.UUID) ([]TestRunItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, run_id, test_case_id, title, status, duration_ms, error_message, stack_trace, artifacts::text, sort_order, created_at, updated_at
		 FROM test_run_items WHERE run_id = $1 ORDER BY sort_order ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (r *SQLRepository) UpdateItem(ctx context.Context, item *TestRunItem) error {
	artifactsJSON, _ := json.Marshal(item.Artifacts)
	result, err := r.db.ExecContext(ctx,
		`UPDATE test_run_items SET title = $2, status = $3, duration_ms = $4, error_message = $5, stack_trace = $6, artifacts = $7, updated_at = $8
		 WHERE id = $1`,
		item.ID, item.Title, string(item.Status), item.DurationMs,
		item.ErrorMessage, item.StackTrace, artifactsJSON, item.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteItemsByRunID(ctx context.Context, runID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM test_run_items WHERE run_id = $1`, runID)
	return err
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanRun(row rowScanner) (*TestRun, error) {
	var run TestRun
	var suiteID sql.NullString
	var metaStr string
	var source string
	var status string
	var startedAt, completedAt sql.NullTime
	if err := row.Scan(&run.ID, &run.WorkspaceID, &run.ProjectID, &suiteID, &run.Name, &status,
		&run.Total, &run.Passed, &run.Failed, &run.Skipped, &run.Blocked, &run.DurationMs,
		&source, &metaStr, &run.CreatedBy, &startedAt, &completedAt, &run.CreatedAt, &run.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	run.Status = RunStatus(status)
	run.Source = RunSource(source)
	if suiteID.Valid {
		sid, _ := uuid.Parse(suiteID.String)
		run.SuiteID = &sid
	}
	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}
	if metaStr != "" {
		_ = json.Unmarshal([]byte(metaStr), &run.Metadata)
	}
	if run.Metadata == nil {
		run.Metadata = make(map[string]interface{})
	}
	return &run, nil
}

func scanRuns(rows *sql.Rows) ([]TestRun, error) {
	var runs []TestRun
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}
	return runs, rows.Err()
}

func scanItem(row rowScanner) (*TestRunItem, error) {
	var item TestRunItem
	var testCaseID sql.NullString
	var artifactsStr string
	var status string
	if err := row.Scan(&item.ID, &item.RunID, &testCaseID, &item.Title, &status,
		&item.DurationMs, &item.ErrorMessage, &item.StackTrace, &artifactsStr,
		&item.SortOrder, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	item.Status = RunItemStatus(status)
	if testCaseID.Valid {
		tcid, _ := uuid.Parse(testCaseID.String)
		item.TestCaseID = &tcid
	}
	if artifactsStr != "" && artifactsStr != "[]" {
		_ = json.Unmarshal([]byte(artifactsStr), &item.Artifacts)
	}
	return &item, nil
}

func scanItems(rows *sql.Rows) ([]TestRunItem, error) {
	var items []TestRunItem
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}
