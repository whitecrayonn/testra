package apitesting

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/pagination"
)

// SQLRepository implements Repository using PostgreSQL.
type SQLRepository struct {
	db db.DBTX
}

// NewSQLRepository creates a new SQL-backed repository.
func NewSQLRepository(sqlDB *sql.DB) *SQLRepository {
	return &SQLRepository{db: db.Wrap(sqlDB)}
}

func nullUUIDPtr(ns sql.NullString) (*uuid.UUID, error) {
	if !ns.Valid {
		return nil, nil
	}
	id, err := uuid.Parse(ns.String)
	if err != nil {
		return nil, fmt.Errorf("invalid stored uuid: %w", err)
	}
	return &id, nil
}

func marshalKeyValuePairs(pairs []KeyValuePair) ([]byte, error) {
	if pairs == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(pairs)
}

func unmarshalKeyValuePairs(data string) ([]KeyValuePair, error) {
	var pairs []KeyValuePair
	if data == "" {
		return pairs, nil
	}
	if err := json.Unmarshal([]byte(data), &pairs); err != nil {
		return nil, fmt.Errorf("unmarshal key/value pairs: %w", err)
	}
	return pairs, nil
}

func marshalAuthConfig(cfg AuthConfig) ([]byte, error) {
	return json.Marshal(cfg)
}

func unmarshalAuthConfig(data string) (AuthConfig, error) {
	var cfg AuthConfig
	if data == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal auth config: %w", err)
	}
	return cfg, nil
}

// RunInTx executes fn within a transaction. The transactional Repository is
// passed to fn and the app.tenant_id session variable is propagated.
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

// Collections

func (r *SQLRepository) CreateCollection(ctx context.Context, c *Collection) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO api_collections (id, workspace_id, name, description, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		c.ID, c.WorkspaceID, c.Name, c.Description, c.CreatedBy, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetCollectionByID(ctx context.Context, id uuid.UUID) (*Collection, error) {
	var c Collection
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, name, description, created_by, created_at, updated_at
		 FROM api_collections WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *SQLRepository) ListCollections(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Collection, error) {
	var rows *sql.Rows
	var err error
	if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, name, description, created_by, created_at, updated_at
			 FROM api_collections WHERE workspace_id = $1 AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, name, description, created_by, created_at, updated_at
			 FROM api_collections WHERE workspace_id = $1
			 ORDER BY id DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []Collection
	for rows.Next() {
		var c Collection
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		collections = append(collections, c)
	}
	return collections, rows.Err()
}

func (r *SQLRepository) UpdateCollection(ctx context.Context, c *Collection) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE api_collections SET name = $2, description = $3, updated_at = $4 WHERE id = $1`,
		c.ID, c.Name, c.Description, c.UpdatedAt,
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

func (r *SQLRepository) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM api_collections WHERE id = $1`, id)
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

// Folders

func (r *SQLRepository) CreateFolder(ctx context.Context, f *Folder) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO api_folders (id, workspace_id, collection_id, parent_id, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		f.ID, f.WorkspaceID, f.CollectionID, f.ParentID, f.Name, f.CreatedAt, f.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetFolderByID(ctx context.Context, id uuid.UUID) (*Folder, error) {
	var f Folder
	var parentID sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, collection_id, parent_id, name, created_at, updated_at
		 FROM api_folders WHERE id = $1`,
		id,
	).Scan(&f.ID, &f.WorkspaceID, &f.CollectionID, &parentID, &f.Name, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	f.ParentID, err = nullUUIDPtr(parentID)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *SQLRepository) ListFolders(ctx context.Context, collectionID uuid.UUID, parentID *uuid.UUID, cursor string, limit int) ([]Folder, error) {
	var rows *sql.Rows
	var err error
	switch {
	case parentID != nil && cursor != "":
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, parent_id, name, created_at, updated_at
			 FROM api_folders WHERE collection_id = $1 AND parent_id = $2 AND id < $3
			 ORDER BY id DESC LIMIT $4`,
			collectionID, *parentID, cursor, limit,
		)
	case parentID != nil:
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, parent_id, name, created_at, updated_at
			 FROM api_folders WHERE collection_id = $1 AND parent_id = $2
			 ORDER BY id DESC LIMIT $3`,
			collectionID, *parentID, limit,
		)
	case cursor != "":
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, parent_id, name, created_at, updated_at
			 FROM api_folders WHERE collection_id = $1 AND parent_id IS NULL AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			collectionID, cursor, limit,
		)
	default:
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, parent_id, name, created_at, updated_at
			 FROM api_folders WHERE collection_id = $1 AND parent_id IS NULL
			 ORDER BY id DESC LIMIT $2`,
			collectionID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []Folder
	for rows.Next() {
		var f Folder
		var pid sql.NullString
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.CollectionID, &pid, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		f.ParentID, err = nullUUIDPtr(pid)
		if err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	return folders, rows.Err()
}

func (r *SQLRepository) UpdateFolder(ctx context.Context, f *Folder) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE api_folders SET name = $2, parent_id = $3, updated_at = $4 WHERE id = $1`,
		f.ID, f.Name, f.ParentID, f.UpdatedAt,
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
	result, err := r.db.ExecContext(ctx, `DELETE FROM api_folders WHERE id = $1`, id)
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

// Environments

func (r *SQLRepository) CreateEnvironment(ctx context.Context, env *Environment) error {
	variablesJSON, err := marshalKeyValuePairs(env.Variables)
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO api_environments (id, workspace_id, name, variables, created_at, updated_at)
		 VALUES ($1, $2, $3, $4::jsonb, $5, $6)`,
		env.ID, env.WorkspaceID, env.Name, variablesJSON, env.CreatedAt, env.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetEnvironmentByID(ctx context.Context, id uuid.UUID) (*Environment, error) {
	var env Environment
	var variablesJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, name, variables::text, created_at, updated_at
		 FROM api_environments WHERE id = $1`,
		id,
	).Scan(&env.ID, &env.WorkspaceID, &env.Name, &variablesJSON, &env.CreatedAt, &env.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	env.Variables, err = unmarshalKeyValuePairs(variablesJSON)
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *SQLRepository) ListEnvironments(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Environment, error) {
	var rows *sql.Rows
	var err error
	if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, name, variables::text, created_at, updated_at
			 FROM api_environments WHERE workspace_id = $1 AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, name, variables::text, created_at, updated_at
			 FROM api_environments WHERE workspace_id = $1
			 ORDER BY id DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var environments []Environment
	for rows.Next() {
		var env Environment
		var variablesJSON string
		if err := rows.Scan(&env.ID, &env.WorkspaceID, &env.Name, &variablesJSON, &env.CreatedAt, &env.UpdatedAt); err != nil {
			return nil, err
		}
		env.Variables, err = unmarshalKeyValuePairs(variablesJSON)
		if err != nil {
			return nil, err
		}
		environments = append(environments, env)
	}
	return environments, rows.Err()
}

func (r *SQLRepository) UpdateEnvironment(ctx context.Context, env *Environment) error {
	variablesJSON, err := marshalKeyValuePairs(env.Variables)
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}
	result, err := r.db.ExecContext(ctx,
		`UPDATE api_environments SET name = $2, variables = $3::jsonb, updated_at = $4 WHERE id = $1`,
		env.ID, env.Name, variablesJSON, env.UpdatedAt,
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

func (r *SQLRepository) DeleteEnvironment(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM api_environments WHERE id = $1`, id)
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

// Requests

func scanRequest(rows *sql.Rows) (Request, error) {
	var req Request
	var folderID, envID, createdBy sql.NullString
	var headersJSON, queryJSON, authJSON, variablesJSON string
	var method, authType, bodyType string
	err := rows.Scan(
		&req.ID, &req.WorkspaceID, &req.CollectionID, &folderID, &envID,
		&req.Name, &method, &req.URL, &headersJSON, &queryJSON, &authType, &authJSON,
		&bodyType, &req.BodyContent, &variablesJSON, &createdBy, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return req, err
	}
	req.Method = HTTPMethod(method)
	req.AuthType = AuthType(authType)
	req.BodyType = BodyType(bodyType)
	req.FolderID, err = nullUUIDPtr(folderID)
	if err != nil {
		return req, err
	}
	req.EnvironmentID, err = nullUUIDPtr(envID)
	if err != nil {
		return req, err
	}
	req.Headers, err = unmarshalKeyValuePairs(headersJSON)
	if err != nil {
		return req, err
	}
	req.QueryParams, err = unmarshalKeyValuePairs(queryJSON)
	if err != nil {
		return req, err
	}
	req.AuthConfig, err = unmarshalAuthConfig(authJSON)
	if err != nil {
		return req, err
	}
	req.Variables, err = unmarshalKeyValuePairs(variablesJSON)
	if err != nil {
		return req, err
	}
	if createdBy.Valid {
		req.CreatedBy, err = uuid.Parse(createdBy.String)
		if err != nil {
			return req, fmt.Errorf("invalid stored created_by: %w", err)
		}
	}
	return req, nil
}

func (r *SQLRepository) CreateRequest(ctx context.Context, req *Request) error {
	headersJSON, err := marshalKeyValuePairs(req.Headers)
	if err != nil {
		return fmt.Errorf("marshal headers: %w", err)
	}
	queryJSON, err := marshalKeyValuePairs(req.QueryParams)
	if err != nil {
		return fmt.Errorf("marshal query params: %w", err)
	}
	authJSON, err := marshalAuthConfig(req.AuthConfig)
	if err != nil {
		return fmt.Errorf("marshal auth config: %w", err)
	}
	variablesJSON, err := marshalKeyValuePairs(req.Variables)
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO api_requests (
			id, workspace_id, collection_id, folder_id, environment_id, name, method, url,
			headers, query_params, auth_type, auth_config, body_type, body_content, variables,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb, $11, $12::jsonb, $13, $14, $15::jsonb, $16, $17, $18)`,
		req.ID, req.WorkspaceID, req.CollectionID, req.FolderID, req.EnvironmentID,
		req.Name, string(req.Method), req.URL, headersJSON, queryJSON, string(req.AuthType), authJSON,
		string(req.BodyType), req.BodyContent, variablesJSON, req.CreatedBy, req.CreatedAt, req.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*Request, error) {
	row, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, collection_id, folder_id, environment_id, name, method, url,
			   headers::text, query_params::text, auth_type, auth_config::text, body_type, body_content,
			   variables::text, created_by, created_at, updated_at
		 FROM api_requests WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		if err := row.Err(); err != nil {
			return nil, err
		}
		return nil, sharederrors.ErrNotFound
	}
	req, err := scanRequest(row)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *SQLRepository) ListRequests(ctx context.Context, collectionID uuid.UUID, folderID *uuid.UUID, cursor string, limit int) ([]Request, error) {
	var rows *sql.Rows
	var err error
	switch {
	case folderID != nil && cursor != "":
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, folder_id, environment_id, name, method, url,
				   headers::text, query_params::text, auth_type, auth_config::text, body_type, body_content,
				   variables::text, created_by, created_at, updated_at
			 FROM api_requests WHERE collection_id = $1 AND folder_id = $2 AND id < $3
			 ORDER BY id DESC LIMIT $4`,
			collectionID, *folderID, cursor, limit,
		)
	case folderID != nil:
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, folder_id, environment_id, name, method, url,
				   headers::text, query_params::text, auth_type, auth_config::text, body_type, body_content,
				   variables::text, created_by, created_at, updated_at
			 FROM api_requests WHERE collection_id = $1 AND folder_id = $2
			 ORDER BY id DESC LIMIT $3`,
			collectionID, *folderID, limit,
		)
	case cursor != "":
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, folder_id, environment_id, name, method, url,
				   headers::text, query_params::text, auth_type, auth_config::text, body_type, body_content,
				   variables::text, created_by, created_at, updated_at
			 FROM api_requests WHERE collection_id = $1 AND folder_id IS NULL AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			collectionID, cursor, limit,
		)
	default:
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, folder_id, environment_id, name, method, url,
				   headers::text, query_params::text, auth_type, auth_config::text, body_type, body_content,
				   variables::text, created_by, created_at, updated_at
			 FROM api_requests WHERE collection_id = $1 AND folder_id IS NULL
			 ORDER BY id DESC LIMIT $2`,
			collectionID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		req, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, rows.Err()
}

func (r *SQLRepository) SearchRequests(ctx context.Context, workspaceID uuid.UUID, query string, cursor string, limit int) ([]Request, string, error) {
	pattern := "%" + strings.TrimSpace(query) + "%"
	if query == "" {
		return nil, "", nil
	}

	var rows *sql.Rows
	var err error
	if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, folder_id, environment_id, name, method, url,
				   headers::text, query_params::text, auth_type, auth_config::text, body_type, body_content,
				   variables::text, created_by, created_at, updated_at
			 FROM api_requests
			 WHERE workspace_id = $1 AND (name ILIKE $2 OR url ILIKE $2) AND id < $3::uuid
			 ORDER BY id DESC LIMIT $4`,
			workspaceID, pattern, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, collection_id, folder_id, environment_id, name, method, url,
				   headers::text, query_params::text, auth_type, auth_config::text, body_type, body_content,
				   variables::text, created_by, created_at, updated_at
			 FROM api_requests
			 WHERE workspace_id = $1 AND (name ILIKE $2 OR url ILIKE $2)
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, pattern, limit,
		)
	}
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		req, err := scanRequest(rows)
		if err != nil {
			return nil, "", err
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(requests) == limit {
		nextCursor, err = pagination.EncodeCursor(requests[len(requests)-1].ID.String())
		if err != nil {
			return nil, "", err
		}
	}

	return requests, nextCursor, nil
}

func (r *SQLRepository) UpdateRequest(ctx context.Context, req *Request) error {
	headersJSON, err := marshalKeyValuePairs(req.Headers)
	if err != nil {
		return fmt.Errorf("marshal headers: %w", err)
	}
	queryJSON, err := marshalKeyValuePairs(req.QueryParams)
	if err != nil {
		return fmt.Errorf("marshal query params: %w", err)
	}
	authJSON, err := marshalAuthConfig(req.AuthConfig)
	if err != nil {
		return fmt.Errorf("marshal auth config: %w", err)
	}
	variablesJSON, err := marshalKeyValuePairs(req.Variables)
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE api_requests SET
			collection_id = $2, folder_id = $3, environment_id = $4, name = $5, method = $6, url = $7,
			headers = $8::jsonb, query_params = $9::jsonb, auth_type = $10, auth_config = $11::jsonb,
			body_type = $12, body_content = $13, variables = $14::jsonb, updated_at = $15
		 WHERE id = $1`,
		req.ID, req.CollectionID, req.FolderID, req.EnvironmentID, req.Name, string(req.Method), req.URL,
		headersJSON, queryJSON, string(req.AuthType), authJSON,
		string(req.BodyType), req.BodyContent, variablesJSON, req.UpdatedAt,
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

func (r *SQLRepository) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM api_requests WHERE id = $1`, id)
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

// Execution history

func (r *SQLRepository) CreateRequestHistory(ctx context.Context, h *RequestHistory) error {
	reqHeadersJSON, err := marshalKeyValuePairs(h.RequestHeaders)
	if err != nil {
		return fmt.Errorf("marshal request headers: %w", err)
	}
	respHeadersJSON, err := marshalKeyValuePairs(h.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("marshal response headers: %w", err)
	}

	var responseStatus *int
	if h.ResponseStatus != 0 {
		v := h.ResponseStatus
		responseStatus = &v
	}
	var responseTimeMs *int
	if h.ResponseTimeMs != 0 {
		v := h.ResponseTimeMs
		responseTimeMs = &v
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO api_request_history (
			id, workspace_id, request_id, environment_id, name, method, url,
			request_headers, request_body, response_status, response_status_text,
			response_headers, response_body, response_time_ms, error, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, $11, $12::jsonb, $13, $14, $15, $16, $17)`,
		h.ID, h.WorkspaceID, h.RequestID, h.EnvironmentID, h.Name, h.Method, h.URL,
		reqHeadersJSON, h.RequestBody, responseStatus, h.ResponseStatusText,
		respHeadersJSON, h.ResponseBody, responseTimeMs, h.Error, h.CreatedBy, h.CreatedAt,
	)
	return err
}

func (r *SQLRepository) GetRequestHistoryByID(ctx context.Context, id uuid.UUID) (*RequestHistory, error) {
	row, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, request_id, environment_id, name, method, url,
			   request_headers::text, request_body, response_status, response_status_text,
			   response_headers::text, response_body, response_time_ms, error, created_by, created_at
		 FROM api_request_history WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		if err := row.Err(); err != nil {
			return nil, err
		}
		return nil, sharederrors.ErrNotFound
	}

	var h RequestHistory
	var requestID, envID, createdBy sql.NullString
	var reqHeadersJSON, respHeadersJSON string
	var responseStatus, responseTimeMs sql.NullInt32
	err = row.Scan(
		&h.ID, &h.WorkspaceID, &requestID, &envID, &h.Name, &h.Method, &h.URL,
		&reqHeadersJSON, &h.RequestBody, &responseStatus, &h.ResponseStatusText,
		&respHeadersJSON, &h.ResponseBody, &responseTimeMs, &h.Error, &createdBy, &h.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	h.RequestID, err = nullUUIDPtr(requestID)
	if err != nil {
		return nil, err
	}
	h.EnvironmentID, err = nullUUIDPtr(envID)
	if err != nil {
		return nil, err
	}
	if createdBy.Valid {
		h.CreatedBy, err = uuid.Parse(createdBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid stored created_by: %w", err)
		}
	}
	h.ResponseStatus = int(responseStatus.Int32)
	h.ResponseTimeMs = int(responseTimeMs.Int32)
	h.RequestHeaders, err = unmarshalKeyValuePairs(reqHeadersJSON)
	if err != nil {
		return nil, err
	}
	h.ResponseHeaders, err = unmarshalKeyValuePairs(respHeadersJSON)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *SQLRepository) ListRequestHistory(ctx context.Context, requestID *uuid.UUID, workspaceID uuid.UUID, cursor string, limit int) ([]RequestHistory, error) {
	var rows *sql.Rows
	var err error
	if requestID != nil && cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, request_id, environment_id, name, method, url,
				   request_headers::text, request_body, response_status, response_status_text,
				   response_headers::text, response_body, response_time_ms, error, created_by, created_at
			 FROM api_request_history WHERE request_id = $1 AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			*requestID, cursor, limit,
		)
	} else if requestID != nil {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, request_id, environment_id, name, method, url,
				   request_headers::text, request_body, response_status, response_status_text,
				   response_headers::text, response_body, response_time_ms, error, created_by, created_at
			 FROM api_request_history WHERE request_id = $1
			 ORDER BY id DESC LIMIT $2`,
			*requestID, limit,
		)
	} else if cursor != "" {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, request_id, environment_id, name, method, url,
				   request_headers::text, request_body, response_status, response_status_text,
				   response_headers::text, response_body, response_time_ms, error, created_by, created_at
			 FROM api_request_history WHERE workspace_id = $1 AND id < $2
			 ORDER BY id DESC LIMIT $3`,
			workspaceID, cursor, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, workspace_id, request_id, environment_id, name, method, url,
				   request_headers::text, request_body, response_status, response_status_text,
				   response_headers::text, response_body, response_time_ms, error, created_by, created_at
			 FROM api_request_history WHERE workspace_id = $1
			 ORDER BY id DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []RequestHistory
	for rows.Next() {
		var h RequestHistory
		var requestIDStr, envID, createdBy sql.NullString
		var reqHeadersJSON, respHeadersJSON string
		var responseStatus, responseTimeMs sql.NullInt32
		if err := rows.Scan(
			&h.ID, &h.WorkspaceID, &requestIDStr, &envID, &h.Name, &h.Method, &h.URL,
			&reqHeadersJSON, &h.RequestBody, &responseStatus, &h.ResponseStatusText,
			&respHeadersJSON, &h.ResponseBody, &responseTimeMs, &h.Error, &createdBy, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		h.RequestID, err = nullUUIDPtr(requestIDStr)
		if err != nil {
			return nil, err
		}
		h.EnvironmentID, err = nullUUIDPtr(envID)
		if err != nil {
			return nil, err
		}
		h.CreatedBy, err = uuid.Parse(createdBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid stored created_by: %w", err)
		}
		h.ResponseStatus = int(responseStatus.Int32)
		h.ResponseTimeMs = int(responseTimeMs.Int32)
		h.RequestHeaders, err = unmarshalKeyValuePairs(reqHeadersJSON)
		if err != nil {
			return nil, err
		}
		h.ResponseHeaders, err = unmarshalKeyValuePairs(respHeadersJSON)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}

// search helpers (mirror testmanagement search)

func toTSQuery(query string) string {
	terms := []string{}
	for _, t := range splitTerms(query) {
		if t == "" {
			continue
		}
		// Prefix matching for partial word search.
		terms = append(terms, t+":*")
	}
	return strings.Join(terms, " & ")
}

func splitTerms(query string) []string {
	fields := strings.FieldsFunc(query, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == '.' || r == '_'
	})
	for i := range fields {
		fields[i] = strings.ToLower(strings.TrimSpace(fields[i]))
	}
	return fields
}
