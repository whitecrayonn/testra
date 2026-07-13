package identity

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateMFA(ctx context.Context, userID uuid.UUID, secret string, enabled bool) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error

	CreateResetToken(ctx context.Context, token *PasswordResetToken) error
	GetResetTokenByHash(ctx context.Context, hash string) (*PasswordResetToken, error)
	MarkResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error
}
