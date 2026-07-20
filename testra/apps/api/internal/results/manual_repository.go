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

// CreateItemExecution updates the full test run item, including step results and
// execution metadata. It is the persistence side of an execution save.
func (r *SQLRepository) CreateItemExecution(ctx context.Context, item *TestRunItem) error {
	return r.UpdateItem(ctx, item)
}

func (r *SQLRepository) ListItemsByRunPaged(ctx context.Context, runID uuid.UUID, status, search, cursor string, limit int) ([]TestRunItem, error) {
	base := `SELECT id, run_id, test_case_id, title, status, duration_ms, error_message, stack_trace, artifacts::text, step_results::text, comment, executed_by, executed_at, sort_order, created_at, updated_at
		 FROM test_run_items WHERE run_id = $1`
	args := []interface{}{runID}
	argIdx := 1

	if status != "" {
		argIdx++
		base += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
	}
	if search != "" {
		argIdx++
		base += fmt.Sprintf(" AND title ILIKE $%d", argIdx)
		args = append(args, "%"+search+"%")
	}

	if cursor != "" {
		cursorID, err := pagination.DecodeCursor(cursor)
		if err != nil {
			return nil, sharederrors.ErrInvalidInput
		}
		argIdx++
		base += fmt.Sprintf(" AND id < $%d::uuid", argIdx)
		args = append(args, cursorID)
	}

	argIdx++
	base += fmt.Sprintf(" ORDER BY sort_order ASC, id DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (r *SQLRepository) CreateItemHistory(ctx context.Context, history *RunItemHistory) error {
	stepResultsJSON, _ := json.Marshal(history.StepResults)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_run_item_history (id, run_item_id, status, step_results, comment, duration_ms, executed_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		history.ID, history.RunItemID, string(history.Status), stepResultsJSON,
		history.Comment, history.DurationMs, history.ExecutedBy, history.CreatedAt,
	)
	return err
}

func (r *SQLRepository) ListItemHistory(ctx context.Context, itemID uuid.UUID) ([]RunItemHistory, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, run_item_id, status, step_results::text, comment, duration_ms, executed_by, created_at
		 FROM test_run_item_history WHERE run_item_id = $1 ORDER BY created_at DESC`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []RunItemHistory
	for rows.Next() {
		h, err := scanHistory(rows)
		if err != nil {
			return nil, err
		}
		history = append(history, *h)
	}
	return history, rows.Err()
}

func scanHistory(row rowScanner) (*RunItemHistory, error) {
	var h RunItemHistory
	var stepResultsStr string
	var executedBy sql.NullString
	var status string
	if err := row.Scan(&h.ID, &h.RunItemID, &status, &stepResultsStr, &h.Comment, &h.DurationMs, &executedBy, &h.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	h.Status = RunItemStatus(status)
	if stepResultsStr != "" && stepResultsStr != "[]" {
		_ = json.Unmarshal([]byte(stepResultsStr), &h.StepResults)
	}
	if executedBy.Valid {
		ebid, _ := uuid.Parse(executedBy.String)
		h.ExecutedBy = &ebid
	}
	return &h, nil
}

func (r *SQLRepository) CreateEvidence(ctx context.Context, evidence *EvidenceRef) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_run_item_evidence (id, run_item_id, step_order, file_name, content_type, storage_path, uploaded_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		evidence.ID, evidence.RunItemID, evidence.StepOrder, evidence.FileName,
		evidence.ContentType, evidence.StoragePath, evidence.UploadedBy, evidence.CreatedAt,
	)
	return err
}

func (r *SQLRepository) ListEvidenceByItem(ctx context.Context, itemID uuid.UUID) ([]EvidenceRef, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, run_item_id, step_order, file_name, content_type, storage_path, uploaded_by, created_at
		 FROM test_run_item_evidence WHERE run_item_id = $1 ORDER BY step_order ASC, created_at DESC`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidence []EvidenceRef
	for rows.Next() {
		e, err := scanEvidence(rows)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, *e)
	}
	return evidence, rows.Err()
}

func (r *SQLRepository) DeleteEvidence(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM test_run_item_evidence WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func scanEvidence(row rowScanner) (*EvidenceRef, error) {
	var e EvidenceRef
	var uploadedBy sql.NullString
	if err := row.Scan(&e.ID, &e.RunItemID, &e.StepOrder, &e.FileName, &e.ContentType, &e.StoragePath, &uploadedBy, &e.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	if uploadedBy.Valid {
		uid, _ := uuid.Parse(uploadedBy.String)
		e.UploadedBy = &uid
	}
	return &e, nil
}

func (r *SQLRepository) CreateRunItemDefect(ctx context.Context, itemID, defectID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_run_item_defects (run_item_id, defect_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING`,
		itemID, defectID)
	return err
}

func (r *SQLRepository) ListRunItemDefects(ctx context.Context, itemID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT defect_id FROM test_run_item_defects WHERE run_item_id = $1 ORDER BY created_at DESC`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *SQLRepository) DeleteRunItemDefect(ctx context.Context, itemID, defectID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM test_run_item_defects WHERE run_item_id = $1 AND defect_id = $2`, itemID, defectID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) CreatePlan(ctx context.Context, plan *TestPlan) error {
	configJSON, _ := json.Marshal(plan.Configuration)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_plans (id, workspace_id, project_id, suite_id, name, description, status, configuration, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		plan.ID, plan.WorkspaceID, plan.ProjectID, plan.SuiteID, plan.Name, plan.Description,
		string(plan.Status), configJSON, plan.CreatedBy, plan.CreatedAt, plan.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*TestPlan, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, project_id, suite_id, name, description, status, configuration::text, created_by, created_at, updated_at
		 FROM test_plans WHERE id = $1`, id)
	return scanPlan(row)
}

func (r *SQLRepository) ListPlans(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]TestPlan, error) {
	if limit <= 0 {
		limit = 20
	}
	if cursor == "" {
		rows, err := r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, suite_id, name, description, status, configuration::text, created_by, created_at, updated_at
			 FROM test_plans WHERE project_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2`,
			projectID, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanPlans(rows)
	}

	cursorID, err := pagination.DecodeCursor(cursor)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, project_id, suite_id, name, description, status, configuration::text, created_by, created_at, updated_at
		 FROM test_plans WHERE project_id = $1 AND id < $2::uuid ORDER BY created_at DESC, id DESC LIMIT $3`,
		projectID, cursorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlans(rows)
}

func (r *SQLRepository) UpdatePlan(ctx context.Context, plan *TestPlan) error {
	configJSON, _ := json.Marshal(plan.Configuration)
	result, err := r.db.ExecContext(ctx,
		`UPDATE test_plans SET name = $2, description = $3, status = $4, configuration = $5, updated_at = $6 WHERE id = $1`,
		plan.ID, plan.Name, plan.Description, string(plan.Status), configJSON, plan.UpdatedAt,
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

func (r *SQLRepository) DeletePlan(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM test_plans WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) CreatePlanItem(ctx context.Context, item *TestPlanItem) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_plan_items (id, plan_id, test_case_id, sort_order, created_at) VALUES ($1, $2, $3, $4, $5)`,
		item.ID, item.PlanID, item.TestCaseID, item.SortOrder, item.CreatedAt,
	)
	return err
}

func (r *SQLRepository) ListPlanItems(ctx context.Context, planID uuid.UUID) ([]TestPlanItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, plan_id, test_case_id, sort_order, created_at FROM test_plan_items WHERE plan_id = $1 ORDER BY sort_order ASC`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TestPlanItem
	for rows.Next() {
		item, err := scanPlanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) DeletePlanItemsByPlanID(ctx context.Context, planID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM test_plan_items WHERE plan_id = $1`, planID)
	return err
}

func scanPlan(row rowScanner) (*TestPlan, error) {
	var plan TestPlan
	var suiteID sql.NullString
	var configStr string
	var status string
	if err := row.Scan(&plan.ID, &plan.WorkspaceID, &plan.ProjectID, &suiteID, &plan.Name, &plan.Description, &status, &configStr, &plan.CreatedBy, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	plan.Status = TestPlanStatus(status)
	if suiteID.Valid {
		sid, _ := uuid.Parse(suiteID.String)
		plan.SuiteID = &sid
	}
	if configStr != "" {
		_ = json.Unmarshal([]byte(configStr), &plan.Configuration)
	}
	if plan.Configuration == nil {
		plan.Configuration = make(map[string]interface{})
	}
	return &plan, nil
}

func scanPlans(rows *sql.Rows) ([]TestPlan, error) {
	var plans []TestPlan
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, *plan)
	}
	return plans, rows.Err()
}

func scanPlanItem(row rowScanner) (*TestPlanItem, error) {
	var item TestPlanItem
	if err := row.Scan(&item.ID, &item.PlanID, &item.TestCaseID, &item.SortOrder, &item.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// RunInTx in repository.go already sets tenant_id on the transaction. The manual
// methods above reuse that context; this helper is available for service-level
// transaction orchestration when needed.
func (r *SQLRepository) withTenantInTx(ctx context.Context, fn func(Repository) error) error {
	return r.RunInTx(ctx, func(txRepo Repository) error {
		if tenantID, ok := db.TenantIDFromContext(ctx); ok {
			if tx, ok := txRepo.(*SQLRepository); ok {
				_, _ = tx.db.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", tenantID.String())
			}
		}
		return fn(txRepo)
	})
}
