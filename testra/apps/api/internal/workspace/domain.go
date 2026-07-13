package workspace

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Slug           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Member struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Role        string
	CreatedAt   time.Time
}
