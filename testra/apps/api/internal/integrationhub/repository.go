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
		`INSERT INTO integrations (id, workspace_id, type, name, config, enabled, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		i.ID, i.WorkspaceID, string(i.Type), i.Name, cfgJSON, i.Enabled, i.CreatedBy, i.CreatedAt, i.UpdatedAt)
	return err
}

func (r *SQLRepository) ListIntegrations(ctx context.Context, workspaceID uuid.UUID) ([]Integration, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, type, name, config::text, enabled, created_by, created_at, updated_at
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
		`SELECT id, workspace_id, type, name, config::text, enabled, created_by, created_at, updated_at
		 FROM integrations WHERE id = $1`, id)
	return scanIntegration(row)
}

func (r *SQLRepository) UpdateIntegration(ctx context.Context, i *Integration) error {
	cfgJSON, _ := json.Marshal(i.Config)
	result, err := r.db.ExecContext(ctx,
		`UPDATE integrations SET name = $2, type = $3, config = $4, enabled = $5, updated_at = $6 WHERE id = $1`,
		i.ID, i.Name, string(i.Type), cfgJSON, i.Enabled, i.UpdatedAt)
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
	payloadJSON, _ := json.Marshal(e.Payload)
	intID := uuid.NullUUID{}
	if e.IntegrationID != nil {
		intID = uuid.NullUUID{UUID: *e.IntegrationID, Valid: true}
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO integration_events (id, workspace_id, integration_id, event_type, payload, status, external_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		e.ID, e.WorkspaceID, intID, e.EventType, payloadJSON, e.Status, e.ExternalID, e.CreatedAt, e.UpdatedAt)
	return err
}

func (r *SQLRepository) ListEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]IntegrationEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, integration_id, event_type, payload::text, status, external_id, created_at, updated_at
		 FROM integration_events
		 WHERE workspace_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
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
	var t string
	if err := row.Scan(&i.ID, &i.WorkspaceID, &t, &i.Name, &cfgStr, &i.Enabled, &createdBy, &i.CreatedAt, &i.UpdatedAt); err != nil {
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
	var payloadStr string
	if err := row.Scan(&e.ID, &e.WorkspaceID, &intID, &e.EventType, &payloadStr, &e.Status, &e.ExternalID, &e.CreatedAt, &e.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
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
