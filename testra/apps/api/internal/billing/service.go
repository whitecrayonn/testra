package billing

import (
	"context"
	"errors"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type Service struct {
	repo     Repository
	provider PaymentProvider
}

func NewService(repo Repository, provider PaymentProvider) *Service {
	return &Service{repo: repo, provider: provider}
}

func (s *Service) GetSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	if orgID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	sub, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		if errors.Is(err, sharederrors.ErrNotFound) {
			return s.createDefaultSubscription(ctx, orgID)
		}
		return nil, err
	}
	if s.provider != nil && sub.ProviderSubscriptionID != "" {
		if fresh, err := s.provider.GetSubscription(ctx, sub.ProviderSubscriptionID); err == nil {
			fresh.ID = sub.ID
			fresh.OrganizationID = sub.OrganizationID
			fresh.CreatedAt = sub.CreatedAt
			fresh.UpdatedAt = nowUTC()
			if err := s.repo.UpsertSubscription(ctx, fresh); err != nil {
				return nil, err
			}
			sub = fresh
		}
	}
	return sub, nil
}

func (s *Service) UpdateSubscription(ctx context.Context, input UpdateSubscriptionInput) (*Subscription, error) {
	if input.OrganizationID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	sub, err := s.GetSubscription(ctx, input.OrganizationID)
	if err != nil && !errors.Is(err, sharederrors.ErrNotFound) {
		return nil, err
	}
	if errors.Is(err, sharederrors.ErrNotFound) {
		sub, err = s.createDefaultSubscription(ctx, input.OrganizationID)
		if err != nil {
			return nil, err
		}
	}

	if input.Plan != "" {
		sub.Plan = input.Plan
	}
	if input.Seats > 0 {
		sub.Seats = input.Seats
	}
	if input.CancelAtPeriodEnd != nil {
		sub.CancelAtPeriodEnd = *input.CancelAtPeriodEnd
	}
	sub.UpdatedAt = nowUTC()

	if s.provider != nil && sub.ProviderSubscriptionID != "" {
		if fresh, err := s.provider.UpdateSubscription(ctx, sub); err == nil {
			fresh.ID = sub.ID
			fresh.OrganizationID = sub.OrganizationID
			fresh.CreatedAt = sub.CreatedAt
			fresh.UpdatedAt = sub.UpdatedAt
			sub = fresh
		}
	}

	if err := s.repo.UpsertSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Service) ListInvoices(ctx context.Context, orgID uuid.UUID, limit int) ([]Invoice, error) {
	if orgID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	localInvoices, err := s.repo.ListInvoices(ctx, orgID, limit)
	if err != nil {
		return nil, err
	}

	if s.provider != nil {
		sub, err := s.repo.GetSubscription(ctx, orgID)
		if err == nil && sub.ProviderSubscriptionID != "" {
			providerInvoices, err := s.provider.ListInvoices(ctx, sub.ProviderSubscriptionID)
			if err == nil {
				for _, inv := range providerInvoices {
					inv.ID = uuid.New()
					inv.OrganizationID = orgID
					inv.CreatedAt = nowUTC()
					inv.UpdatedAt = inv.CreatedAt
					if err := s.repo.CreateInvoice(ctx, &inv); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return localInvoices, nil
}

func (s *Service) createDefaultSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	now := nowUTC()
	end := now.AddDate(0, 1, 0)
	sub := &Subscription{
		ID:                 uuid.New(),
		OrganizationID:     orgID,
		Plan:               "free",
		Status:             "active",
		Seats:              1,
		CurrentPeriodStart: &now,
		CurrentPeriodEnd:   &end,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.repo.UpsertSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// Input structs

type UpdateSubscriptionInput struct {
	OrganizationID    uuid.UUID
	Plan              string
	Seats             int
	CancelAtPeriodEnd *bool
}
