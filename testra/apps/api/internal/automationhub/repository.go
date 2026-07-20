package automationhub

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/pagination"
)

type SQLRepository struct {
	db db.DBTX
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

// ----------------- Projects -----------------

func (r *SQLRepository) CreateProject(ctx context.Context, p *AutomationProject) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO automation_projects (id, workspace_id, project_id, name, framework, repository_url, branch, command, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		p.ID, p.WorkspaceID, p.ProjectID, p.Name, p.Framework, p.RepositoryURL, p.Branch, p.Command, p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetProject(ctx context.Context, id uuid.UUID) (*AutomationProject, error) {
	var p AutomationProject
	var projectID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, project_id, name, framework, repository_url, branch, command, created_by, created_at, updated_at
		 FROM automation_projects WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.WorkspaceID, &projectID, &p.Name, &p.Framework, &p.RepositoryURL, &p.Branch, &p.Command, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.ProjectID = nullUUID(projectID)
	return &p, nil
}

func (r *SQLRepository) ListProjects(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]AutomationProject, error) {
	var rows *sql.Rows
	var err error
	if cursor != "" {
		cursorID, err := pagination.DecodeCursor(cursor)
		if err != nil {
			return nil, sharederrors.ErrInvalidInput
		}
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, name, framework, repository_url, branch, command, created_by, created_at, updated_at
			 FROM automation_projects WHERE workspace_id = $1 AND id < $2 ORDER BY id DESC LIMIT $3`,
			workspaceID, cursorID, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, name, framework, repository_url, branch, command, created_by, created_at, updated_at
			 FROM automation_projects WHERE workspace_id = $1 ORDER BY id DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProjects(rows)
}

func (r *SQLRepository) UpdateProject(ctx context.Context, p *AutomationProject) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE automation_projects SET name = $2, framework = $3, repository_url = $4, branch = $5, command = $6, updated_at = $7
		 WHERE id = $1`,
		p.ID, p.Name, p.Framework, p.RepositoryURL, p.Branch, p.Command, p.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM automation_projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

// ----------------- Executions -----------------

func (r *SQLRepository) CreateExecution(ctx context.Context, e *AutomationExecution) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO automation_executions (id, project_id, workspace_id, test_run_id, name, status, report_format, report_path, retry_of, duration_ms, total, passed, failed, skipped, blocked, created_by, triggered_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		e.ID, e.ProjectID, e.WorkspaceID, e.TestRunID, e.Name, e.Status, e.ReportFormat, e.ReportPath, e.RetryOf, e.DurationMs, e.Total, e.Passed, e.Failed, e.Skipped, e.Blocked, e.CreatedBy, e.TriggeredBy, e.CreatedAt, e.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetExecution(ctx context.Context, id uuid.UUID) (*AutomationExecution, error) {
	var e AutomationExecution
	var testRunID, retryOf sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, project_id, workspace_id, test_run_id, name, status, report_format, report_path, retry_of, duration_ms, total, passed, failed, skipped, blocked, created_by, triggered_by, created_at, updated_at
		 FROM automation_executions WHERE id = $1`,
		id,
	).Scan(&e.ID, &e.ProjectID, &e.WorkspaceID, &testRunID, &e.Name, &e.Status, &e.ReportFormat, &e.ReportPath, &retryOf, &e.DurationMs, &e.Total, &e.Passed, &e.Failed, &e.Skipped, &e.Blocked, &e.CreatedBy, &e.TriggeredBy, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	e.TestRunID = nullUUID(testRunID)
	e.RetryOf = nullUUID(retryOf)
	return &e, nil
}

func (r *SQLRepository) ListExecutions(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]AutomationExecution, error) {
	var rows *sql.Rows
	var err error
	if cursor != "" {
		cursorID, err := pagination.DecodeCursor(cursor)
		if err != nil {
			return nil, sharederrors.ErrInvalidInput
		}
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, project_id, workspace_id, test_run_id, name, status, report_format, report_path, retry_of, duration_ms, total, passed, failed, skipped, blocked, created_by, triggered_by, created_at, updated_at
			 FROM automation_executions WHERE project_id = $1 AND id < $2 ORDER BY id DESC LIMIT $3`,
			projectID, cursorID, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, project_id, workspace_id, test_run_id, name, status, report_format, report_path, retry_of, duration_ms, total, passed, failed, skipped, blocked, created_by, triggered_by, created_at, updated_at
			 FROM automation_executions WHERE project_id = $1 ORDER BY id DESC LIMIT $2`,
			projectID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanExecutions(rows)
}

func (r *SQLRepository) ListExecutionsByWorkspace(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]AutomationExecution, error) {
	var rows *sql.Rows
	var err error
	if cursor != "" {
		cursorID, err := pagination.DecodeCursor(cursor)
		if err != nil {
			return nil, sharederrors.ErrInvalidInput
		}
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, project_id, workspace_id, test_run_id, name, status, report_format, report_path, retry_of, duration_ms, total, passed, failed, skipped, blocked, created_by, triggered_by, created_at, updated_at
			 FROM automation_executions WHERE workspace_id = $1 AND id < $2 ORDER BY id DESC LIMIT $3`,
			workspaceID, cursorID, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, project_id, workspace_id, test_run_id, name, status, report_format, report_path, retry_of, duration_ms, total, passed, failed, skipped, blocked, created_by, triggered_by, created_at, updated_at
			 FROM automation_executions WHERE workspace_id = $1 ORDER BY id DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanExecutions(rows)
}

func (r *SQLRepository) UpdateExecution(ctx context.Context, e *AutomationExecution) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE automation_executions SET name = $2, status = $3, report_format = $4, report_path = $5, retry_of = $6, duration_ms = $7, total = $8, passed = $9, failed = $10, skipped = $11, blocked = $12, test_run_id = $13, updated_at = $14
		 WHERE id = $1`,
		e.ID, e.Name, e.Status, e.ReportFormat, e.ReportPath, e.RetryOf, e.DurationMs, e.Total, e.Passed, e.Failed, e.Skipped, e.Blocked, e.TestRunID, e.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteExecution(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM automation_executions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

// ----------------- Artifacts -----------------

func (r *SQLRepository) CreateArtifact(ctx context.Context, a *AutomationArtifact) error {
	metaJSON, err := json.Marshal(a.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO automation_artifacts (id, execution_id, workspace_id, test_run_item_id, kind, name, file_path, mime_type, file_size, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		a.ID, a.ExecutionID, a.WorkspaceID, a.TestRunItemID, a.Kind, a.Name, a.FilePath, a.MimeType, a.FileSize, metaJSON, a.CreatedAt,
	)
	return err
}

func (r *SQLRepository) GetArtifact(ctx context.Context, id uuid.UUID) (*AutomationArtifact, error) {
	var a AutomationArtifact
	var itemID sql.NullString
	var metaJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, execution_id, workspace_id, test_run_item_id, kind, name, file_path, mime_type, file_size, metadata::text, created_at
		 FROM automation_artifacts WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.ExecutionID, &a.WorkspaceID, &itemID, &a.Kind, &a.Name, &a.FilePath, &a.MimeType, &a.FileSize, &metaJSON, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.TestRunItemID = nullUUID(itemID)
	if err := json.Unmarshal([]byte(metaJSON), &a.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return &a, nil
}

func (r *SQLRepository) ListArtifacts(ctx context.Context, executionID uuid.UUID, cursor string, limit int) ([]AutomationArtifact, error) {
	var rows *sql.Rows
	var err error
	if cursor != "" {
		cursorID, err := pagination.DecodeCursor(cursor)
		if err != nil {
			return nil, sharederrors.ErrInvalidInput
		}
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, execution_id, workspace_id, test_run_item_id, kind, name, file_path, mime_type, file_size, metadata::text, created_at
			 FROM automation_artifacts WHERE execution_id = $1 AND id < $2 ORDER BY id DESC LIMIT $3`,
			executionID, cursorID, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, execution_id, workspace_id, test_run_item_id, kind, name, file_path, mime_type, file_size, metadata::text, created_at
			 FROM automation_artifacts WHERE execution_id = $1 ORDER BY id DESC LIMIT $2`,
			executionID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArtifacts(rows)
}

func (r *SQLRepository) DeleteArtifact(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM automation_artifacts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

// ----------------- Logs -----------------

func (r *SQLRepository) CreateLog(ctx context.Context, l *AutomationLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO automation_logs (id, execution_id, workspace_id, level, message, logged_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		l.ID, l.ExecutionID, l.WorkspaceID, l.Level, l.Message, l.LoggedAt, l.CreatedAt,
	)
	return err
}

func (r *SQLRepository) ListLogs(ctx context.Context, executionID uuid.UUID, cursor string, limit int) ([]AutomationLog, error) {
	var rows *sql.Rows
	var err error
	if cursor != "" {
		cursorID, err := pagination.DecodeCursor(cursor)
		if err != nil {
			return nil, sharederrors.ErrInvalidInput
		}
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, execution_id, workspace_id, level, message, logged_at, created_at
			 FROM automation_logs WHERE execution_id = $1 AND id < $2 ORDER BY id DESC LIMIT $3`,
			executionID, cursorID, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, execution_id, workspace_id, level, message, logged_at, created_at
			 FROM automation_logs WHERE execution_id = $1 ORDER BY id DESC LIMIT $2`,
			executionID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLogs(rows)
}

// ----------------- scanners -----------------

func scanProjects(rows *sql.Rows) ([]AutomationProject, error) {
	var projects []AutomationProject
	for rows.Next() {
		var p AutomationProject
		var projectID sql.NullString
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &projectID, &p.Name, &p.Framework, &p.RepositoryURL, &p.Branch, &p.Command, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.ProjectID = nullUUID(projectID)
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func scanExecutions(rows *sql.Rows) ([]AutomationExecution, error) {
	var executions []AutomationExecution
	for rows.Next() {
		var e AutomationExecution
		var testRunID, retryOf sql.NullString
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.WorkspaceID, &testRunID, &e.Name, &e.Status, &e.ReportFormat, &e.ReportPath, &retryOf, &e.DurationMs, &e.Total, &e.Passed, &e.Failed, &e.Skipped, &e.Blocked, &e.CreatedBy, &e.TriggeredBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		e.TestRunID = nullUUID(testRunID)
		e.RetryOf = nullUUID(retryOf)
		executions = append(executions, e)
	}
	return executions, rows.Err()
}

func scanArtifacts(rows *sql.Rows) ([]AutomationArtifact, error) {
	var artifacts []AutomationArtifact
	for rows.Next() {
		var a AutomationArtifact
		var itemID sql.NullString
		var metaJSON string
		if err := rows.Scan(&a.ID, &a.ExecutionID, &a.WorkspaceID, &itemID, &a.Kind, &a.Name, &a.FilePath, &a.MimeType, &a.FileSize, &metaJSON, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.TestRunItemID = nullUUID(itemID)
		if err := json.Unmarshal([]byte(metaJSON), &a.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
		artifacts = append(artifacts, a)
	}
	return artifacts, rows.Err()
}

func scanLogs(rows *sql.Rows) ([]AutomationLog, error) {
	var logs []AutomationLog
	for rows.Next() {
		var l AutomationLog
		if err := rows.Scan(&l.ID, &l.ExecutionID, &l.WorkspaceID, &l.Level, &l.Message, &l.LoggedAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func nullUUID(ns sql.NullString) *uuid.UUID {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	id, err := uuid.Parse(ns.String)
	if err != nil {
		return nil
	}
	return &id
}
