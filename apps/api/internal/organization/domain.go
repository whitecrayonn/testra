package organization

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	OwnerID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Member struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	Role           string
	CreatedAt      time.Time
}
