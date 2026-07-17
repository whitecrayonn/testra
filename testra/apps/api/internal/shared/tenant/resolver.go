package tenant

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
)

var ErrNotFound = errors.New("tenant resolution failed: resource not found")

type Resolver struct {
	db db.DBTX
}

func NewResolver(db db.DBTX) *Resolver {
	return &Resolver{db: db}
}

func (r *Resolver) ResolveOrgFromWorkspace(ctx context.Context, workspaceID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT organization_id FROM workspaces WHERE id = $1`,
		workspaceID,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromProject(ctx context.Context, projectID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM projects p
		 JOIN workspaces w ON p.workspace_id = w.id
		 WHERE p.id = $1`,
		projectID,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromAPIKey(ctx context.Context, apiKeyID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM api_keys ak
		 JOIN workspaces w ON ak.workspace_id = w.id
		 WHERE ak.id = $1`,
		apiKeyID,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromRunItem(ctx context.Context, itemID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM test_run_items tri
		 JOIN test_runs tr ON tri.run_id = tr.id
		 JOIN workspaces w ON tr.workspace_id = w.id
		 WHERE tri.id = $1`,
		itemID,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromRun(ctx context.Context, runID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM test_runs tr
		 JOIN workspaces w ON tr.workspace_id = w.id
		 WHERE tr.id = $1`,
		runID,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromDefect(ctx context.Context, defectID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM defects d
		 JOIN workspaces w ON d.workspace_id = w.id
		 WHERE d.id = $1`,
		defectID,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) CheckMembership(ctx context.Context, userID, orgID uuid.UUID) error {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM organization_members
			WHERE organization_id = $1 AND user_id = $2
		)`,
		orgID, userID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("forbidden: not a member of this organization")
	}
	return nil
}
