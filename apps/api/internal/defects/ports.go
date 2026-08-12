package defects

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, defect *Defect) error
	GetByID(ctx context.Context, id uuid.UUID) (*Defect, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]Defect, error)
	Update(ctx context.Context, defect *Defect) error
	Delete(ctx context.Context, id uuid.UUID) error
}
