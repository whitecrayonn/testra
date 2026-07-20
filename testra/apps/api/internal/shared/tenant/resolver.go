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

func (r *Resolver) ResolveOrgFromAPICollection(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM api_collections c
		 JOIN workspaces w ON c.workspace_id = w.id
		 WHERE c.id = $1`,
		id,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromAPIFolder(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM api_folders f
		 JOIN workspaces w ON f.workspace_id = w.id
		 WHERE f.id = $1`,
		id,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromAPIRequest(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM api_requests req
		 JOIN workspaces w ON req.workspace_id = w.id
		 WHERE req.id = $1`,
		id,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromAPIEnvironment(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM api_environments env
		 JOIN workspaces w ON env.workspace_id = w.id
		 WHERE env.id = $1`,
		id,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromAPIRequestHistory(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM api_request_history h
		 JOIN workspaces w ON h.workspace_id = w.id
		 WHERE h.id = $1`,
		id,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromAutomationProject(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM automation_projects p
		 JOIN workspaces w ON p.workspace_id = w.id
		 WHERE p.id = $1`,
		id,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromAutomationExecution(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM automation_executions e
		 JOIN workspaces w ON e.workspace_id = w.id
		 WHERE e.id = $1`,
		id,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromAutomationArtifact(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM automation_artifacts a
		 JOIN workspaces w ON a.workspace_id = w.id
		 WHERE a.id = $1`,
		id,
	).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return orgID, err
}

func (r *Resolver) ResolveOrgFromTestPlan(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT w.organization_id
		 FROM test_plans p
		 JOIN workspaces w ON p.workspace_id = w.id
		 WHERE p.id = $1`,
		id,
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
