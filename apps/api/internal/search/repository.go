package search

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
)

type SQLRepository struct {
	db db.DBTX
}

func NewSQLRepository(sqlDB *sql.DB) *SQLRepository {
	return &SQLRepository{db: db.Wrap(sqlDB)}
}

func (r *SQLRepository) Search(ctx context.Context, workspaceID uuid.UUID, query string, limit int) (*Result, error) {
	q := "%" + query + "%"
	res := &Result{WorkspaceID: workspaceID, Query: query}

	err := r.queryProjects(ctx, workspaceID, q, limit, res)
	if err != nil {
		return nil, err
	}
	err = r.queryTestCases(ctx, workspaceID, query, limit, res)
	if err != nil {
		return nil, err
	}
	err = r.queryDefects(ctx, workspaceID, q, limit, res)
	if err != nil {
		return nil, err
	}
	err = r.queryAutomation(ctx, workspaceID, q, limit, res)
	if err != nil {
		return nil, err
	}
	err = r.queryAPICollections(ctx, workspaceID, q, limit, res)
	if err != nil {
		return nil, err
	}
	err = r.queryTestPlans(ctx, workspaceID, q, limit, res)
	if err != nil {
		return nil, err
	}
	err = r.queryUsers(ctx, workspaceID, q, limit, res)
	if err != nil {
		return nil, err
	}

	res.Results = make([]SearchItem, 0, len(res.Projects)+len(res.TestCases)+len(res.Defects)+len(res.Automation)+len(res.APICollections)+len(res.TestPlans)+len(res.Users))
	res.Results = append(res.Results, res.Projects...)
	res.Results = append(res.Results, res.TestCases...)
	res.Results = append(res.Results, res.Defects...)
	res.Results = append(res.Results, res.Automation...)
	res.Results = append(res.Results, res.APICollections...)
	res.Results = append(res.Results, res.TestPlans...)
	res.Results = append(res.Results, res.Users...)

	return res, nil
}

func (r *SQLRepository) queryProjects(ctx context.Context, workspaceID uuid.UUID, q string, limit int, res *Result) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, key, description FROM projects
		 WHERE workspace_id = $1 AND (name ILIKE $2 OR key ILIKE $2 OR description ILIKE $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		workspaceID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name, key, desc string
		if err := rows.Scan(&id, &name, &key, &desc); err != nil {
			return err
		}
		res.Projects = append(res.Projects, SearchItem{
			ID:       id.String(),
			Type:     "project",
			Title:    name,
			Subtitle: key,
			URL:      "/dashboard/projects",
		})
	}
	return rows.Err()
}

func (r *SQLRepository) queryTestCases(ctx context.Context, workspaceID uuid.UUID, q string, limit int, res *Result) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, status, priority FROM test_cases
		 WHERE workspace_id = $1 AND search_tsv @@ plainto_tsquery('english', $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		workspaceID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var title, status, priority string
		if err := rows.Scan(&id, &title, &status, &priority); err != nil {
			return err
		}
		res.TestCases = append(res.TestCases, SearchItem{
			ID:       id.String(),
			Type:     "test_case",
			Title:    title,
			Subtitle: status + " · " + priority,
			URL:      "/dashboard/test-cases",
		})
	}
	return rows.Err()
}

func (r *SQLRepository) queryDefects(ctx context.Context, workspaceID uuid.UUID, q string, limit int, res *Result) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, status, severity FROM defects
		 WHERE workspace_id = $1 AND (title ILIKE $2 OR description ILIKE $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		workspaceID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var title, status, severity string
		if err := rows.Scan(&id, &title, &status, &severity); err != nil {
			return err
		}
		res.Defects = append(res.Defects, SearchItem{
			ID:       id.String(),
			Type:     "defect",
			Title:    title,
			Subtitle: status + " · " + severity,
			URL:      "/dashboard/defects",
		})
	}
	return rows.Err()
}

func (r *SQLRepository) queryAutomation(ctx context.Context, workspaceID uuid.UUID, q string, limit int, res *Result) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, framework FROM automation_projects
		 WHERE workspace_id = $1 AND (name ILIKE $2 OR framework ILIKE $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		workspaceID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name, framework string
		if err := rows.Scan(&id, &name, &framework); err != nil {
			return err
		}
		res.Automation = append(res.Automation, SearchItem{
			ID:       id.String(),
			Type:     "automation",
			Title:    name,
			Subtitle: framework,
			URL:      "/dashboard/automation",
		})
	}
	return rows.Err()
}

func (r *SQLRepository) queryAPICollections(ctx context.Context, workspaceID uuid.UUID, q string, limit int, res *Result) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, COALESCE(description, '') FROM api_collections
		 WHERE workspace_id = $1 AND (name ILIKE $2 OR description ILIKE $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		workspaceID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name, desc string
		if err := rows.Scan(&id, &name, &desc); err != nil {
			return err
		}
		res.APICollections = append(res.APICollections, SearchItem{
			ID:       id.String(),
			Type:     "api_collection",
			Title:    name,
			Subtitle: desc,
			URL:      "/dashboard/api-tests",
		})
	}
	return rows.Err()
}

func (r *SQLRepository) queryTestPlans(ctx context.Context, workspaceID uuid.UUID, q string, limit int, res *Result) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, status FROM test_plans
		 WHERE workspace_id = $1 AND (name ILIKE $2 OR description ILIKE $2)
		 ORDER BY updated_at DESC LIMIT $3`,
		workspaceID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name, status string
		if err := rows.Scan(&id, &name, &status); err != nil {
			return err
		}
		res.TestPlans = append(res.TestPlans, SearchItem{
			ID:       id.String(),
			Type:     "test_plan",
			Title:    name,
			Subtitle: status,
			URL:      "/dashboard/test-plans",
		})
	}
	return rows.Err()
}

func (r *SQLRepository) queryUsers(ctx context.Context, workspaceID uuid.UUID, q string, limit int, res *Result) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT u.id, u.name, u.email FROM users u
		 JOIN organization_members om ON u.id = om.user_id
		 WHERE om.organization_id = (SELECT organization_id FROM workspaces WHERE id = $1)
		   AND (u.name ILIKE $2 OR u.email ILIKE $2)
		 ORDER BY u.name LIMIT $3`,
		workspaceID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			return err
		}
		res.Users = append(res.Users, SearchItem{
			ID:       id.String(),
			Type:     "user",
			Title:    name,
			Subtitle: email,
			URL:      "/dashboard/settings/members",
		})
	}
	return rows.Err()
}
