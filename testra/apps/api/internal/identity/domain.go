package identity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	Email      string
	Password   string // hashed
	Name       string
	MFASecret  string // TOTP secret, empty if not enrolled
	MFAEnabled bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type RefreshToken struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	TokenHash         string
	FamilyID          uuid.UUID
	ExpiresAt         time.Time
	AbsoluteExpiresAt time.Time
	RevokedAt         *time.Time
	ReplacedBy        *uuid.UUID
	CreatedAt         time.Time
}
