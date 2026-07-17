package integrationhub

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateIntegration(ctx context.Context, i *Integration) error
	ListIntegrations(ctx context.Context, workspaceID uuid.UUID) ([]Integration, error)
	GetIntegration(ctx context.Context, id uuid.UUID) (*Integration, error)
	UpdateIntegration(ctx context.Context, i *Integration) error
	DeleteIntegration(ctx context.Context, id uuid.UUID) error

	CreateEvent(ctx context.Context, e *IntegrationEvent) error
	ListEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]IntegrationEvent, error)
}

type Adapter interface {
	Send(ctx context.Context, i *Integration, payload map[string]interface{}) (externalID string, err error)
}
