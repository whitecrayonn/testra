package integrationhub

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (r *SQLRepository) CreateIntegration(ctx context.Context, i *Integration) error {
	cfgJSON, _ := json.Marshal(i.Config)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO integrations (id, workspace_id, type, name, config, enabled, health_status, last_tested_at, last_error, sync_status, retry_count, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		i.ID, i.WorkspaceID, string(i.Type), i.Name, cfgJSON, i.Enabled, i.HealthStatus, i.LastTestedAt, i.LastError, i.SyncStatus, i.RetryCount, i.CreatedBy, i.CreatedAt, i.UpdatedAt)
	return err
}

func (r *SQLRepository) ListIntegrations(ctx context.Context, workspaceID uuid.UUID) ([]Integration, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, type, name, config::text, enabled, health_status, last_tested_at, last_error, sync_status, retry_count, created_by, created_at, updated_at
		 FROM integrations WHERE workspace_id = $1 ORDER BY created_at DESC`,
		workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIntegrations(rows)
}

func (r *SQLRepository) GetIntegration(ctx context.Context, id uuid.UUID) (*Integration, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, type, name, config::text, enabled, health_status, last_tested_at, last_error, sync_status, retry_count, created_by, created_at, updated_at
		 FROM integrations WHERE id = $1`, id)
	return scanIntegration(row)
}

func (r *SQLRepository) UpdateIntegration(ctx context.Context, i *Integration) error {
	cfgJSON, _ := json.Marshal(i.Config)
	result, err := r.db.ExecContext(ctx,
		`UPDATE integrations SET name = $2, type = $3, config = $4, enabled = $5, health_status = $6, last_tested_at = $7, last_error = $8, sync_status = $9, retry_count = $10, updated_at = $11 WHERE id = $1`,
		i.ID, i.Name, string(i.Type), cfgJSON, i.Enabled, i.HealthStatus, i.LastTestedAt, i.LastError, i.SyncStatus, i.RetryCount, i.UpdatedAt)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM integrations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) CreateEvent(ctx context.Context, e *IntegrationEvent) error {
	payloadJSON, err := json.Marshal(e.Payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	intID := uuid.NullUUID{}
	if e.IntegrationID != nil {
		intID = uuid.NullUUID{UUID: *e.IntegrationID, Valid: true}
	}
	workspaceID := uuid.NullUUID{UUID: e.WorkspaceID, Valid: e.WorkspaceID != uuid.Nil}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO integration_events (id, workspace_id, integration_id, event_type, payload, status, external_id, retry_count, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		e.ID, workspaceID, intID, e.EventType, payloadJSON, e.Status, e.ExternalID, e.RetryCount, e.CreatedAt, e.UpdatedAt)
	return err
}

func (r *SQLRepository) UpdateEvent(ctx context.Context, e *IntegrationEvent) error {
	payloadJSON, err := json.Marshal(e.Payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	intID := uuid.NullUUID{}
	if e.IntegrationID != nil {
		intID = uuid.NullUUID{UUID: *e.IntegrationID, Valid: true}
	}
	workspaceID := uuid.NullUUID{UUID: e.WorkspaceID, Valid: e.WorkspaceID != uuid.Nil}
	result, err := r.db.ExecContext(ctx,
		`UPDATE integration_events SET workspace_id = $2, integration_id = $3, event_type = $4, payload = $5, status = $6, external_id = $7, retry_count = $8, updated_at = $9 WHERE id = $1`,
		e.ID, workspaceID, intID, e.EventType, payloadJSON, e.Status, e.ExternalID, e.RetryCount, e.UpdatedAt)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) GetEvent(ctx context.Context, id uuid.UUID) (*IntegrationEvent, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, integration_id, event_type, payload::text, status, external_id, retry_count, created_at, updated_at
		 FROM integration_events WHERE id = $1`, id)
	return scanEvent(row)
}

func (r *SQLRepository) ListEvents(ctx context.Context, workspaceID uuid.UUID, status string, limit int) ([]IntegrationEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `SELECT id, workspace_id, integration_id, event_type, payload::text, status, external_id, retry_count, created_at, updated_at
		 FROM integration_events WHERE workspace_id = $1 `
	args := []interface{}{workspaceID}
	if status != "" {
		query += ` AND status = $2 `
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (r *SQLRepository) ListEventsByStatus(ctx context.Context, workspaceID uuid.UUID, status string, limit int) ([]IntegrationEvent, error) {
	return r.ListEvents(ctx, workspaceID, status, limit)
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

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanIntegration(row rowScanner) (*Integration, error) {
	var i Integration
	var createdBy sql.NullString
	var cfgStr string
	var lastTested sql.NullTime
	var lastError sql.NullString
	var t string
	if err := row.Scan(&i.ID, &i.WorkspaceID, &t, &i.Name, &cfgStr, &i.Enabled, &i.HealthStatus, &lastTested, &lastError, &i.SyncStatus, &i.RetryCount, &createdBy, &i.CreatedAt, &i.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	i.Type = IntegrationType(t)
	if cfgStr != "" {
		_ = json.Unmarshal([]byte(cfgStr), &i.Config)
	}
	if i.Config == nil {
		i.Config = make(map[string]string)
	}
	if lastTested.Valid {
		i.LastTestedAt = &lastTested.Time
	}
	if lastError.Valid {
		i.LastError = lastError.String
	}
	if createdBy.Valid {
		i.CreatedBy, _ = uuid.Parse(createdBy.String)
	}
	return &i, nil
}

func scanIntegrations(rows *sql.Rows) ([]Integration, error) {
	var list []Integration
	for rows.Next() {
		i, err := scanIntegration(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *i)
	}
	return list, rows.Err()
}

func scanEvent(row rowScanner) (*IntegrationEvent, error) {
	var e IntegrationEvent
	var intID sql.NullString
	var wsID sql.NullString
	var payloadStr string
	if err := row.Scan(&e.ID, &wsID, &intID, &e.EventType, &payloadStr, &e.Status, &e.ExternalID, &e.RetryCount, &e.CreatedAt, &e.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	if wsID.Valid {
		id, _ := uuid.Parse(wsID.String)
		e.WorkspaceID = id
	}
	if intID.Valid {
		id, _ := uuid.Parse(intID.String)
		e.IntegrationID = &id
	}
	if payloadStr != "" {
		_ = json.Unmarshal([]byte(payloadStr), &e.Payload)
	}
	if e.Payload == nil {
		e.Payload = make(map[string]interface{})
	}
	return &e, nil
}

func scanEvents(rows *sql.Rows) ([]IntegrationEvent, error) {
	var list []IntegrationEvent
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *e)
	}
	return list, rows.Err()
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
