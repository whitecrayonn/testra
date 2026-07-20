package testmanagement

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
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

func (r *SQLRepository) CreateFolder(ctx context.Context, folder *TestFolder) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_folders (id, workspace_id, parent_id, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		folder.ID, folder.WorkspaceID, folder.ParentID, folder.Name, folder.CreatedAt, folder.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetFolderByID(ctx context.Context, id uuid.UUID) (*TestFolder, error) {
	var f TestFolder
	var parentID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, parent_id, name, created_at, updated_at FROM test_folders WHERE id = $1`,
		id,
	).Scan(&f.ID, &f.WorkspaceID, &parentID, &f.Name, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if parentID.Valid {
		pid, err := uuid.Parse(parentID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid stored parent_id: %w", err)
		}
		f.ParentID = &pid
	}
	return &f, nil
}

func (r *SQLRepository) ListFolders(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID, cursor string, limit int) ([]TestFolder, error) {
	var rows *sql.Rows
	var err error

	if parentID != nil && cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, parent_id, name, created_at, updated_at FROM test_folders
			 WHERE workspace_id = $1 AND parent_id = $2 AND id < $3
			 ORDER BY id DESC LIMIT $4`,
			workspaceID, *parentID, cursor, limit,
		)
	} else if parentID != nil {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, parent_id, name, created_at, updated_at FROM test_folders
			 WHERE workspace_id = $1 AND parent_id = $2
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, *parentID, limit,
		)
	} else if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, parent_id, name, created_at, updated_at FROM test_folders
			 WHERE workspace_id = $1 AND parent_id IS NULL AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, parent_id, name, created_at, updated_at FROM test_folders
			 WHERE workspace_id = $1 AND parent_id IS NULL
			 ORDER BY id DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []TestFolder
	for rows.Next() {
		var f TestFolder
		var pid sql.NullString
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &pid, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		if pid.Valid {
			parsed, err := uuid.Parse(pid.String)
			if err != nil {
				return nil, fmt.Errorf("invalid stored parent_id: %w", err)
			}
			f.ParentID = &parsed
		}
		folders = append(folders, f)
	}
	return folders, rows.Err()
}

func (r *SQLRepository) UpdateFolder(ctx context.Context, folder *TestFolder) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE test_folders SET name = $2, updated_at = $3 WHERE id = $1`,
		folder.ID, folder.Name, folder.UpdatedAt,
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

func (r *SQLRepository) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM test_folders WHERE id = $1`, id)
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

func (r *SQLRepository) CreateSuite(ctx context.Context, suite *TestSuite) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO test_suites (id, workspace_id, folder_id, name, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		suite.ID, suite.WorkspaceID, suite.FolderID, suite.Name, suite.Description, suite.CreatedAt, suite.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetSuiteByID(ctx context.Context, id uuid.UUID) (*TestSuite, error) {
	var s TestSuite
	var folderID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, folder_id, name, description, created_at, updated_at FROM test_suites WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.WorkspaceID, &folderID, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if folderID.Valid {
		fid, err := uuid.Parse(folderID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid stored folder_id: %w", err)
		}
		s.FolderID = &fid
	}
	return &s, nil
}

func (r *SQLRepository) ListSuites(ctx context.Context, workspaceID uuid.UUID, folderID *uuid.UUID, cursor string, limit int) ([]TestSuite, error) {
	var rows *sql.Rows
	var err error

	if folderID != nil && cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, folder_id, name, description, created_at, updated_at FROM test_suites
			 WHERE workspace_id = $1 AND folder_id = $2 AND id < $3
			 ORDER BY id DESC LIMIT $4`,
			workspaceID, *folderID, cursor, limit,
		)
	} else if folderID != nil {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, folder_id, name, description, created_at, updated_at FROM test_suites
			 WHERE workspace_id = $1 AND folder_id = $2
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, *folderID, limit,
		)
	} else if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, folder_id, name, description, created_at, updated_at FROM test_suites
			 WHERE workspace_id = $1 AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, folder_id, name, description, created_at, updated_at FROM test_suites
			 WHERE workspace_id = $1
			 ORDER BY id DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suites []TestSuite
	for rows.Next() {
		var s TestSuite
		var fid sql.NullString
		if err := rows.Scan(&s.ID, &s.WorkspaceID, &fid, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if fid.Valid {
			parsed, err := uuid.Parse(fid.String)
			if err != nil {
				return nil, fmt.Errorf("invalid stored folder_id: %w", err)
			}
			s.FolderID = &parsed
		}
		suites = append(suites, s)
	}
	return suites, rows.Err()
}

func (r *SQLRepository) UpdateSuite(ctx context.Context, suite *TestSuite) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE test_suites SET name = $2, description = $3, updated_at = $4 WHERE id = $1`,
		suite.ID, suite.Name, suite.Description, suite.UpdatedAt,
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

func (r *SQLRepository) DeleteSuite(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM test_suites WHERE id = $1`, id)
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

func (r *SQLRepository) CreateCase(ctx context.Context, tc *TestCase) error {
	stepsJSON, err := json.Marshal(tc.Steps)
	if err != nil {
		return fmt.Errorf("marshal steps: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO test_cases (id, workspace_id, project_id, suite_id, title, description, preconditions, steps, status, priority, tags, version, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		tc.ID, tc.WorkspaceID, tc.ProjectID, tc.SuiteID, tc.Title, tc.Description, tc.Preconditions,
		stepsJSON, string(tc.Status), string(tc.Priority), pq.Array(tc.Tags), tc.Version,
		tc.CreatedBy, tc.CreatedAt, tc.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetCaseByID(ctx context.Context, id uuid.UUID) (*TestCase, error) {
	var tc TestCase
	var suiteID sql.NullString
	var stepsJSON string
	var tags []string
	var status, priority string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, project_id, suite_id, title, description, preconditions, steps::text, status, priority, tags, version, created_by, created_at, updated_at
		 FROM test_cases WHERE id = $1`,
		id,
	).Scan(&tc.ID, &tc.WorkspaceID, &tc.ProjectID, &suiteID, &tc.Title, &tc.Description, &tc.Preconditions,
		&stepsJSON, &status, &priority, pq.Array(&tags), &tc.Version, &tc.CreatedBy, &tc.CreatedAt, &tc.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	tc.Status = TestCaseStatus(status)
	tc.Priority = TestCasePriority(priority)
	tc.Tags = tags
	if suiteID.Valid {
		sid, err := uuid.Parse(suiteID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid stored suite_id: %w", err)
		}
		tc.SuiteID = &sid
	}
	if err := json.Unmarshal([]byte(stepsJSON), &tc.Steps); err != nil {
		return nil, fmt.Errorf("unmarshal steps: %w", err)
	}
	return &tc, nil
}

func (r *SQLRepository) ListCases(ctx context.Context, projectID uuid.UUID, suiteID *uuid.UUID, cursor string, limit int) ([]TestCase, error) {
	var rows *sql.Rows
	var err error

	if suiteID != nil && cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, suite_id, title, description, preconditions, steps::text, status, priority, tags, version, created_by, created_at, updated_at
			 FROM test_cases WHERE project_id = $1 AND suite_id = $2 AND id < $3
			 ORDER BY id DESC LIMIT $4`,
			projectID, *suiteID, cursor, limit,
		)
	} else if suiteID != nil {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, suite_id, title, description, preconditions, steps::text, status, priority, tags, version, created_by, created_at, updated_at
			 FROM test_cases WHERE project_id = $1 AND suite_id = $2
			 ORDER BY id DESC LIMIT $3`,
			projectID, *suiteID, limit,
		)
	} else if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, suite_id, title, description, preconditions, steps::text, status, priority, tags, version, created_by, created_at, updated_at
			 FROM test_cases WHERE project_id = $1 AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			projectID, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, suite_id, title, description, preconditions, steps::text, status, priority, tags, version, created_by, created_at, updated_at
			 FROM test_cases WHERE project_id = $1
			 ORDER BY id DESC LIMIT $2`,
			projectID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCases(rows)
}

func (r *SQLRepository) SearchCases(ctx context.Context, workspaceID uuid.UUID, query string, cursor string, limit int) ([]TestCase, string, error) {
	tsQuery := toTSQuery(query)
	if tsQuery == "" {
		return nil, "", nil
	}

	var cursorRank float64
	var cursorID string
	hasCursor := false
	if cursor != "" {
		r, id, err := decodeSearchCursor(cursor)
		if err != nil {
			return nil, "", sharederrors.ErrInvalidInput
		}
		cursorRank = r
		cursorID = id
		hasCursor = true
	}

	var rows *sql.Rows
	var err error

	if hasCursor {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, suite_id, title, description, preconditions, steps::text, status, priority, tags, version, created_by, created_at, updated_at,
			        ts_rank(search_tsv, to_tsquery('pg_catalog.english', $2)) AS rank
			 FROM test_cases
			 WHERE workspace_id = $1 AND search_tsv @@ to_tsquery('pg_catalog.english', $2)
			   AND (ts_rank(search_tsv, to_tsquery('pg_catalog.english', $2)) < $3
			        OR (ts_rank(search_tsv, to_tsquery('pg_catalog.english', $2)) = $3 AND id < $4::uuid))
			 ORDER BY rank DESC, id DESC LIMIT $5`,
			workspaceID, tsQuery, cursorRank, cursorID, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, project_id, suite_id, title, description, preconditions, steps::text, status, priority, tags, version, created_by, created_at, updated_at,
			        ts_rank(search_tsv, to_tsquery('pg_catalog.english', $2)) AS rank
			 FROM test_cases
			 WHERE workspace_id = $1 AND search_tsv @@ to_tsquery('pg_catalog.english', $2)
			 ORDER BY rank DESC, id DESC LIMIT $3`,
			workspaceID, tsQuery, limit,
		)
	}
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	cases, lastRank, err := scanCasesWithRank(rows)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(cases) == limit {
		lastID := cases[len(cases)-1].ID.String()
		nextCursor, err = encodeSearchCursor(lastRank, lastID)
		if err != nil {
			return nil, "", err
		}
	}

	return cases, nextCursor, nil
}

func (r *SQLRepository) UpdateCase(ctx context.Context, tc *TestCase) error {
	stepsJSON, err := json.Marshal(tc.Steps)
	if err != nil {
		return fmt.Errorf("marshal steps: %w", err)
	}
	result, err := r.db.ExecContext(ctx,
		`UPDATE test_cases SET title = $2, description = $3, preconditions = $4, steps = $5, status = $6, priority = $7, tags = $8, suite_id = $9, version = $10, updated_at = $11
		 WHERE id = $1`,
		tc.ID, tc.Title, tc.Description, tc.Preconditions, stepsJSON,
		string(tc.Status), string(tc.Priority), pq.Array(tc.Tags), tc.SuiteID, tc.Version, tc.UpdatedAt,
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

func (r *SQLRepository) DeleteCase(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM test_cases WHERE id = $1`, id)
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

func (r *SQLRepository) CreateVersion(ctx context.Context, version *TestCaseVersion) error {
	stepsJSON, err := json.Marshal(version.Steps)
	if err != nil {
		return fmt.Errorf("marshal steps: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO test_case_versions (id, test_case_id, version, title, description, preconditions, steps, changed_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		version.ID, version.TestCaseID, version.Version, version.Title, version.Description,
		version.Preconditions, stepsJSON, version.ChangedBy, version.CreatedAt,
	)
	return err
}

func (r *SQLRepository) ListVersions(ctx context.Context, caseID uuid.UUID, cursor string, limit int) ([]TestCaseVersion, error) {
	var rows *sql.Rows
	var err error

	if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, test_case_id, version, title, description, preconditions, steps::text, changed_by, created_at
			 FROM test_case_versions WHERE test_case_id = $1 AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			caseID, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, test_case_id, version, title, description, preconditions, steps::text, changed_by, created_at
			 FROM test_case_versions WHERE test_case_id = $1
			 ORDER BY id DESC LIMIT $2`,
			caseID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []TestCaseVersion
	for rows.Next() {
		var v TestCaseVersion
		var stepsJSON string
		if err := rows.Scan(&v.ID, &v.TestCaseID, &v.Version, &v.Title, &v.Description, &v.Preconditions, &stepsJSON, &v.ChangedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(stepsJSON), &v.Steps); err != nil {
			return nil, fmt.Errorf("unmarshal steps: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func scanCases(rows *sql.Rows) ([]TestCase, error) {
	var cases []TestCase
	for rows.Next() {
		var tc TestCase
		var suiteID sql.NullString
		var stepsJSON string
		var tags []string
		var status, priority string
		if err := rows.Scan(&tc.ID, &tc.WorkspaceID, &tc.ProjectID, &suiteID, &tc.Title, &tc.Description, &tc.Preconditions,
			&stepsJSON, &status, &priority, pq.Array(&tags), &tc.Version, &tc.CreatedBy, &tc.CreatedAt, &tc.UpdatedAt); err != nil {
			return nil, err
		}
		tc.Status = TestCaseStatus(status)
		tc.Priority = TestCasePriority(priority)
		tc.Tags = tags
		if suiteID.Valid {
			sid, err := uuid.Parse(suiteID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid stored suite_id: %w", err)
			}
			tc.SuiteID = &sid
		}
		if err := json.Unmarshal([]byte(stepsJSON), &tc.Steps); err != nil {
			return nil, fmt.Errorf("unmarshal steps: %w", err)
		}
		cases = append(cases, tc)
	}
	return cases, rows.Err()
}

func scanCasesWithRank(rows *sql.Rows) ([]TestCase, float64, error) {
	var cases []TestCase
	var lastRank float64
	for rows.Next() {
		var tc TestCase
		var suiteID sql.NullString
		var stepsJSON string
		var tags []string
		var status, priority string
		var rank float64
		if err := rows.Scan(&tc.ID, &tc.WorkspaceID, &tc.ProjectID, &suiteID, &tc.Title, &tc.Description, &tc.Preconditions,
			&stepsJSON, &status, &priority, pq.Array(&tags), &tc.Version, &tc.CreatedBy, &tc.CreatedAt, &tc.UpdatedAt, &rank); err != nil {
			return nil, 0, err
		}
		tc.Status = TestCaseStatus(status)
		tc.Priority = TestCasePriority(priority)
		tc.Tags = tags
		if suiteID.Valid {
			sid, err := uuid.Parse(suiteID.String)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid stored suite_id: %w", err)
			}
			tc.SuiteID = &sid
		}
		if err := json.Unmarshal([]byte(stepsJSON), &tc.Steps); err != nil {
			return nil, 0, fmt.Errorf("unmarshal steps: %w", err)
		}
		cases = append(cases, tc)
		lastRank = rank
	}
	return cases, lastRank, rows.Err()
}

func encodeSearchCursor(rank float64, id string) (string, error) {
	b, err := json.Marshal(map[string]string{"rank": strconv.FormatFloat(rank, 'f', -1, 64), "id": id})
	if err != nil {
		return "", fmt.Errorf("failed to encode search cursor: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func decodeSearchCursor(cursor string) (float64, string, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, "", fmt.Errorf("invalid search cursor: %w", err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return 0, "", fmt.Errorf("invalid search cursor: %w", err)
	}
	rank, err := strconv.ParseFloat(m["rank"], 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid search cursor rank: %w", err)
	}
	return rank, m["id"], nil
}

func toTSQuery(query string) string {
	words := splitWords(query)
	if len(words) == 0 {
		return ""
	}
	result := ""
	for i, w := range words {
		if i > 0 {
			result += " & "
		}
		result += w
	}
	return result
}

func splitWords(s string) []string {
	var words []string
	current := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			current += string(c)
		} else if current != "" {
			words = append(words, current)
			current = ""
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
