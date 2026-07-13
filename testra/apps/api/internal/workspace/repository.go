package workspace

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, workspace *Workspace) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workspaces (id, organization_id, name, slug, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		workspace.ID, workspace.OrganizationID, workspace.Name, workspace.Slug, workspace.CreatedAt, workspace.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error) {
	var w Workspace
	err := r.db.QueryRowContext(ctx,
		`SELECT id, organization_id, name, slug, created_at, updated_at FROM workspaces WHERE id = $1`,
		id,
	).Scan(&w.ID, &w.OrganizationID, &w.Name, &w.Slug, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *SQLRepository) GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*Workspace, error) {
	var w Workspace
	err := r.db.QueryRowContext(ctx,
		`SELECT id, organization_id, name, slug, created_at, updated_at FROM workspaces
		 WHERE organization_id = $1 AND slug = $2`,
		orgID, slug,
	).Scan(&w.ID, &w.OrganizationID, &w.Name, &w.Slug, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *SQLRepository) ListForOrganization(ctx context.Context, orgID uuid.UUID) ([]Workspace, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, organization_id, name, slug, created_at, updated_at FROM workspaces
		 WHERE organization_id = $1
		 ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.OrganizationID, &w.Name, &w.Slug, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, rows.Err()
}

func (r *SQLRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Workspace, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT w.id, w.organization_id, w.name, w.slug, w.created_at, w.updated_at
		 FROM workspaces w
		 JOIN workspace_members wm ON w.id = wm.workspace_id
		 WHERE wm.user_id = $1
		 ORDER BY w.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.OrganizationID, &w.Name, &w.Slug, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, rows.Err()
}

func (r *SQLRepository) AddMember(ctx context.Context, member *Member) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role, created_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (workspace_id, user_id) DO NOTHING`,
		member.WorkspaceID, member.UserID, member.Role, member.CreatedAt,
	)
	return err
}
