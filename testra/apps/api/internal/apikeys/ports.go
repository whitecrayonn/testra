package apikeys

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, key *APIKey) error
	GetByHash(ctx context.Context, hash string) (*APIKey, error)
	ListForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]APIKey, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}
