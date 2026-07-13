package identity

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

func (r *SQLRepository) Create(ctx context.Context, user *User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.Password, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password, name, mfa_secret, mfa_enabled, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.MFASecret, &user.MFAEnabled, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password, name, mfa_secret, mfa_enabled, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.MFASecret, &user.MFAEnabled, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLRepository) UpdateMFA(ctx context.Context, userID uuid.UUID, secret string, enabled bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET mfa_secret = $1, mfa_enabled = $2, updated_at = NOW() WHERE id = $3`,
		secret, enabled, userID,
	)
	return err
}

func (r *SQLRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`,
		passwordHash, userID,
	)
	return err
}

func (r *SQLRepository) CreateResetToken(ctx context.Context, token *PasswordResetToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt,
	)
	return err
}

func (r *SQLRepository) GetResetTokenByHash(ctx context.Context, hash string) (*PasswordResetToken, error) {
	var t PasswordResetToken
	var usedAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at FROM password_reset_tokens WHERE token_hash = $1`,
		hash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &usedAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if usedAt.Valid {
		t.UsedAt = &usedAt.Time
	}
	return &t, nil
}

func (r *SQLRepository) MarkResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`,
		tokenID,
	)
	return err
}
