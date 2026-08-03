package testgen

import (
	"context"

	"github.com/google/uuid"
)

// Repository persists generation run records only. Test cases themselves are
// created through testmanagement.Repository (injected into Service), the
// same cross-module composition pattern automationhub uses for
// testmanagement — see apps/api/internal/automationhub/service.go.
type Repository interface {
	CreateRun(ctx context.Context, run *GenerationRun) error
	GetRunByID(ctx context.Context, id uuid.UUID) (*GenerationRun, error)
}
