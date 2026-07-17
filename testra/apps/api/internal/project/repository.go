package project

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

func (r *SQLRepository) Create(ctx context.Context, project *Project) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (id, workspace_id, name, key, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		project.ID, project.WorkspaceID, project.Name, project.Key, project.Description, project.CreatedAt, project.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	var p Project
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, name, key, description, created_at, updated_at FROM projects WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Key, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *SQLRepository) GetByKey(ctx context.Context, workspaceID uuid.UUID, key string) (*Project, error) {
	var p Project
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, name, key, description, created_at, updated_at FROM projects
		 WHERE workspace_id = $1 AND key = $2`,
		workspaceID, key,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Key, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *SQLRepository) ListForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Project, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, name, key, description, created_at, updated_at FROM projects
		 WHERE workspace_id = $1
		 ORDER BY created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Key, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *SQLRepository) ListForWorkspacePaginated(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Project, error) {
	if cursor != "" {
		rows, err := r.db.QueryContext(ctx,
			`SELECT id, workspace_id, name, key, description, created_at, updated_at FROM projects
			 WHERE workspace_id = $1 AND id < $2
			 ORDER BY id DESC
			 LIMIT $3`,
			workspaceID, cursor, limit,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var projects []Project
		for rows.Next() {
			var p Project
			if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Key, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
				return nil, err
			}
			projects = append(projects, p)
		}
		return projects, rows.Err()
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, name, key, description, created_at, updated_at FROM projects
		 WHERE workspace_id = $1
		 ORDER BY id DESC
		 LIMIT $2`,
		workspaceID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Key, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}
