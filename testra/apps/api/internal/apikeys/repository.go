package apikeys

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

func (r *SQLRepository) Create(ctx context.Context, key *APIKey) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO api_keys (id, workspace_id, name, key_hash, key_prefix, scopes, last_used_at, expires_at, revoked_at, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		key.ID, key.WorkspaceID, key.Name, key.KeyHash, key.KeyPrefix,
		pqArray(key.Scopes), key.LastUsedAt, key.ExpiresAt, key.RevokedAt, key.CreatedBy, key.CreatedAt,
	)
	return err
}

func (r *SQLRepository) GetByHash(ctx context.Context, hash string) (*APIKey, error) {
	var k APIKey
	var scopes string
	var lastUsed, expires, revoked sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, name, key_hash, key_prefix, scopes::text, last_used_at, expires_at, revoked_at, created_by, created_at
		 FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL`,
		hash,
	).Scan(&k.ID, &k.WorkspaceID, &k.Name, &k.KeyHash, &k.KeyPrefix, &scopes, &lastUsed, &expires, &revoked, &k.CreatedBy, &k.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	k.Scopes = parseArray(scopes)
	if lastUsed.Valid {
		k.LastUsedAt = &lastUsed.Time
	}
	if expires.Valid {
		k.ExpiresAt = &expires.Time
	}
	if revoked.Valid {
		k.RevokedAt = &revoked.Time
	}
	return &k, nil
}

func (r *SQLRepository) ListForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]APIKey, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, name, key_hash, key_prefix, scopes::text, last_used_at, expires_at, revoked_at, created_by, created_at
		 FROM api_keys WHERE workspace_id = $1 AND revoked_at IS NULL
		 ORDER BY created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		var scopes string
		var lastUsed, expires, revoked sql.NullTime
		if err := rows.Scan(&k.ID, &k.WorkspaceID, &k.Name, &k.KeyHash, &k.KeyPrefix, &scopes, &lastUsed, &expires, &revoked, &k.CreatedBy, &k.CreatedAt); err != nil {
			return nil, err
		}
		k.Scopes = parseArray(scopes)
		if lastUsed.Valid {
			k.LastUsedAt = &lastUsed.Time
		}
		if expires.Valid {
			k.ExpiresAt = &expires.Time
		}
		if revoked.Valid {
			k.RevokedAt = &revoked.Time
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (r *SQLRepository) ListForWorkspacePaginated(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]APIKey, error) {
	if cursor != "" {
		rows, err := r.db.QueryContext(ctx,
			`SELECT id, workspace_id, name, key_hash, key_prefix, scopes::text, last_used_at, expires_at, revoked_at, created_by, created_at
			 FROM api_keys WHERE workspace_id = $1 AND revoked_at IS NULL AND id < $2
			 ORDER BY id DESC
			 LIMIT $3`,
			workspaceID, cursor, limit,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var keys []APIKey
		for rows.Next() {
			var k APIKey
			var scopes string
			var lastUsed, expires, revoked sql.NullTime
			if err := rows.Scan(&k.ID, &k.WorkspaceID, &k.Name, &k.KeyHash, &k.KeyPrefix, &scopes, &lastUsed, &expires, &revoked, &k.CreatedBy, &k.CreatedAt); err != nil {
				return nil, err
			}
			k.Scopes = parseArray(scopes)
			if lastUsed.Valid {
				k.LastUsedAt = &lastUsed.Time
			}
			if expires.Valid {
				k.ExpiresAt = &expires.Time
			}
			if revoked.Valid {
				k.RevokedAt = &revoked.Time
			}
			keys = append(keys, k)
		}
		return keys, rows.Err()
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, name, key_hash, key_prefix, scopes::text, last_used_at, expires_at, revoked_at, created_by, created_at
		 FROM api_keys WHERE workspace_id = $1 AND revoked_at IS NULL
		 ORDER BY id DESC
		 LIMIT $2`,
		workspaceID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		var scopes string
		var lastUsed, expires, revoked sql.NullTime
		if err := rows.Scan(&k.ID, &k.WorkspaceID, &k.Name, &k.KeyHash, &k.KeyPrefix, &scopes, &lastUsed, &expires, &revoked, &k.CreatedBy, &k.CreatedAt); err != nil {
			return nil, err
		}
		k.Scopes = parseArray(scopes)
		if lastUsed.Valid {
			k.LastUsedAt = &lastUsed.Time
		}
		if expires.Valid {
			k.ExpiresAt = &expires.Time
		}
		if revoked.Valid {
			k.RevokedAt = &revoked.Time
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (r *SQLRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE api_keys SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sharederrors.ErrNotFound
	}
	return nil
}

func (r *SQLRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

func pqArray(scopes []string) string {
	if len(scopes) == 0 {
		return "{}"
	}
	result := "{"
	for i, s := range scopes {
		if i > 0 {
			result += ","
		}
		result += s
	}
	result += "}"
	return result
}

func parseArray(s string) []string {
	if len(s) < 2 {
		return nil
	}
	inner := s[1 : len(s)-1]
	if inner == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i <= len(inner); i++ {
		if i == len(inner) || inner[i] == ',' {
			result = append(result, inner[start:i])
			start = i + 1
		}
	}
	return result
}
