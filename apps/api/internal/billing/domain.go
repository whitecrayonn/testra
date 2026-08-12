package billing

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID                     uuid.UUID
	OrganizationID         uuid.UUID
	ProviderSubscriptionID string
	Plan                   string
	Status                 string
	Seats                  int
	CurrentPeriodStart     *time.Time
	CurrentPeriodEnd       *time.Time
	CancelAtPeriodEnd      bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type Invoice struct {
	ID                uuid.UUID
	OrganizationID    uuid.UUID
	ProviderInvoiceID string
	AmountCents       int
	Currency          string
	Status            string
	PeriodStart       *time.Time
	PeriodEnd         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
