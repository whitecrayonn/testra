package organization

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]Organization, error)
	AddMember(ctx context.Context, member *Member) error
	GetMember(ctx context.Context, orgID, userID uuid.UUID) (*Member, error)
}
