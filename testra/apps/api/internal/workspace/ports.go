package workspace

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	GetBySlug(ctx context.Context, orgID uuid.UUID, slug string) (*Workspace, error)
	ListForOrganization(ctx context.Context, orgID uuid.UUID) ([]Workspace, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]Workspace, error)
	AddMember(ctx context.Context, member *Member) error
}
