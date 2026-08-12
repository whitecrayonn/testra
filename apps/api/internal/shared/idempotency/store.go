package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

const DefaultTTL = 24 * time.Hour

// Store persists idempotency records for replayable side-effecting operations.
type Store interface {
	Get(ctx context.Context, orgID, workspaceID uuid.UUID, operation, key string) (*Record, error)
	Save(ctx context.Context, record *Record) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

// Record captures a processed request so that identical retries return the same response.
type Record struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	WorkspaceID    uuid.UUID
	Operation      string
	Key            string
	Fingerprint    string
	ResponseBody   json.RawMessage
	StatusCode     int
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

// PostgresStore is a PostgreSQL implementation of Store.
type PostgresStore struct {
	db db.DBTX
}

func NewPostgresStore(db db.DBTX) *PostgresStore {
	return &PostgresStore{db: db}
}

// nullUUID returns a UUID value that can be passed to the database driver.
// uuid.Nil is passed as nil so Postgres stores NULL rather than the nil UUID.
func nullUUID(id uuid.UUID) interface{} {
	if id == uuid.Nil {
		return nil
	}
	return id.String()
}

func (s *PostgresStore) Get(ctx context.Context, orgID, workspaceID uuid.UUID, operation, key string) (*Record, error) {
	var record Record
	var expiresAt time.Time

	workspaceArg := nullUUID(workspaceID)
	err := s.db.QueryRowContext(ctx,
		`SELECT request_fingerprint, response_body, status_code, expires_at
		 FROM idempotency_records
		 WHERE organization_id = $1 AND workspace_id IS NOT DISTINCT FROM $2 AND operation = $3 AND key = $4 AND expires_at > NOW()`,
		orgID, workspaceArg, operation, key,
	).Scan(&record.Fingerprint, &record.ResponseBody, &record.StatusCode, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	record.OrganizationID = orgID
	record.WorkspaceID = workspaceID
	record.Operation = operation
	record.Key = key
	record.ExpiresAt = expiresAt
	return &record, nil
}

func (s *PostgresStore) Save(ctx context.Context, record *Record) error {
	workspaceArg := nullUUID(record.WorkspaceID)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO idempotency_records (id, organization_id, workspace_id, operation, key, request_fingerprint, response_body, status_code, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (organization_id, workspace_id, operation, key) DO NOTHING`,
		record.ID, record.OrganizationID, workspaceArg, record.Operation, record.Key,
		record.Fingerprint, record.ResponseBody, record.StatusCode,
		record.CreatedAt, record.ExpiresAt,
	)
	return err
}

func (s *PostgresStore) DeleteExpired(ctx context.Context, before time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM idempotency_records WHERE expires_at <= $1`,
		before,
	)
	return err
}

// HashKey returns a SHA-256 hex digest of a caller-provided idempotency key.
func HashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// Fingerprint returns a SHA-256 hex digest of a normalized request body.
// JSON payloads are compacted to remove insignificant whitespace; other bodies are hashed as-is.
func Fingerprint(body []byte) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, body); err == nil {
		body = buf.Bytes()
	}
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}
