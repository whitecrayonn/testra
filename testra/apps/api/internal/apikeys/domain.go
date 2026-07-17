package apikeys

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	KeyHash        string
	KeyPrefix      string
	Scopes         []string
	LastUsedAt     *time.Time
	ExpiresAt      *time.Time
	RevokedAt      *time.Time
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
}

func (k *APIKey) GetWorkspaceID() uuid.UUID    { return k.WorkspaceID }
func (k *APIKey) GetOrganizationID() uuid.UUID { return k.OrganizationID }
func (k *APIKey) GetCreatedBy() uuid.UUID      { return k.CreatedBy }
func (k *APIKey) GetScopes() []string          { return k.Scopes }
