package billing

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error)
	UpsertSubscription(ctx context.Context, s *Subscription) error
	ListInvoices(ctx context.Context, orgID uuid.UUID, limit int) ([]Invoice, error)
	CreateInvoice(ctx context.Context, inv *Invoice) error
}

// PaymentProvider abstracts an external billing service such as Stripe.
// The local repository is always the source of truth; the provider is used
// to sync with the external system when configured.
type PaymentProvider interface {
	GetSubscription(ctx context.Context, providerID string) (*Subscription, error)
	UpdateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	ListInvoices(ctx context.Context, providerSubscriptionID string) ([]Invoice, error)
}
