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
	UpdateEvent(ctx context.Context, e *IntegrationEvent) error
	GetEvent(ctx context.Context, id uuid.UUID) (*IntegrationEvent, error)
	ListEvents(ctx context.Context, workspaceID uuid.UUID, status string, limit int) ([]IntegrationEvent, error)
	ListEventsByStatus(ctx context.Context, workspaceID uuid.UUID, status string, limit int) ([]IntegrationEvent, error)
}

// WebhookVerifier is implemented by providers that can verify incoming webhook signatures.
type WebhookVerifier interface {
	VerifyWebhook(i *Integration, body []byte, signature string) error
}

type Adapter interface {
	Send(ctx context.Context, i *Integration, payload map[string]interface{}) (externalID string, err error)
}
