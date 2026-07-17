package defects

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type SQLRepository struct {
	db db.DBTX
}

func NewSQLRepository(sqlDB *sql.DB) *SQLRepository {
	return &SQLRepository{db: db.Wrap(sqlDB)}
}

func (r *SQLRepository) Create(ctx context.Context, defect *Defect) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO defects (id, workspace_id, project_id, test_run_item_id, title, description, severity, priority, status, assigned_to, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		defect.ID, defect.WorkspaceID, defect.ProjectID, defect.TestRunItemID, defect.Title, defect.Description,
		defect.Severity, defect.Priority, defect.Status, defect.AssignedTo, defect.CreatedBy, defect.CreatedAt, defect.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetByID(ctx context.Context, id uuid.UUID) (*Defect, error) {
	var d Defect
	var itemID, assignedTo uuid.NullUUID
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, project_id, test_run_item_id, title, description, severity, priority, status, assigned_to, created_by, created_at, updated_at
		 FROM defects WHERE id = $1`,
		id,
	).Scan(&d.ID, &d.WorkspaceID, &d.ProjectID, &itemID, &d.Title, &d.Description, &d.Severity, &d.Priority, &d.Status, &assignedTo, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d.TestRunItemID = nullUUIDPtr(itemID)
	d.AssignedTo = nullUUIDPtr(assignedTo)
	return &d, nil
}

func (r *SQLRepository) ListByProject(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]Defect, error) {
	if cursor != "" {
		rows, err := r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, test_run_item_id, title, description, severity, priority, status, assigned_to, created_by, created_at, updated_at
			 FROM defects WHERE project_id = $1 AND id < $2
			 ORDER BY id DESC
			 LIMIT $3`,
			projectID, cursor, limit,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanDefects(rows)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, project_id, test_run_item_id, title, description, severity, priority, status, assigned_to, created_by, created_at, updated_at
		 FROM defects WHERE project_id = $1
		 ORDER BY id DESC
		 LIMIT $2`,
		projectID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDefects(rows)
}

func (r *SQLRepository) Update(ctx context.Context, defect *Defect) error {
	defect.UpdatedAt = time.Now().UTC()
	result, err := r.db.ExecContext(ctx,
		`UPDATE defects
		 SET title = $2, description = $3, severity = $4, priority = $5, status = $6, assigned_to = $7, updated_at = $8
		 WHERE id = $1`,
		defect.ID, defect.Title, defect.Description, defect.Severity, defect.Priority, defect.Status, defect.AssignedTo, defect.UpdatedAt,
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

func (r *SQLRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM defects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func scanDefects(rows *sql.Rows) ([]Defect, error) {
	var defects []Defect
	for rows.Next() {
		var d Defect
		var itemID, assignedTo uuid.NullUUID
		if err := rows.Scan(&d.ID, &d.WorkspaceID, &d.ProjectID, &itemID, &d.Title, &d.Description, &d.Severity, &d.Priority, &d.Status, &assignedTo, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.TestRunItemID = nullUUIDPtr(itemID)
		d.AssignedTo = nullUUIDPtr(assignedTo)
		defects = append(defects, d)
	}
	return defects, rows.Err()
}

func nullUUIDPtr(n uuid.NullUUID) *uuid.UUID {
	if n.Valid {
		return &n.UUID
	}
	return nil
}
