package testgen

import (
	"context"
	"database/sql"
	"errors"

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

func (r *SQLRepository) CreateRun(ctx context.Context, run *GenerationRun) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO generation_runs (id, workspace_id, project_id, source, spec_filename, endpoint_count, case_count, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		run.ID, run.WorkspaceID, run.ProjectID, run.Source, run.SpecFilename, run.EndpointCount, run.CaseCount, run.CreatedBy, run.CreatedAt,
	)
	return err
}

func (r *SQLRepository) GetRunByID(ctx context.Context, id uuid.UUID) (*GenerationRun, error) {
	var run GenerationRun
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, project_id, source, spec_filename, endpoint_count, case_count, created_by, created_at
		 FROM generation_runs WHERE id = $1`,
		id,
	).Scan(&run.ID, &run.WorkspaceID, &run.ProjectID, &run.Source, &run.SpecFilename, &run.EndpointCount, &run.CaseCount, &run.CreatedBy, &run.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}
