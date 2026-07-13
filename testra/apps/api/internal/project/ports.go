package project

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	GetByKey(ctx context.Context, workspaceID uuid.UUID, key string) (*Project, error)
	ListForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Project, error)
}
