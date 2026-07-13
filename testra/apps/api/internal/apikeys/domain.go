package apikeys

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	KeyHash     string
	KeyPrefix   string
	Scopes      []string
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time
	RevokedAt   *time.Time
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
}
