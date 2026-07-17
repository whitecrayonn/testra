package billing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type SQLRepository struct {
	db DBTX
}

func NewSQLRepository(sqlDB *sql.DB) *SQLRepository {
	return &SQLRepository{db: db.Wrap(sqlDB)}
}

func (r *SQLRepository) GetSubscription(ctx context.Context, orgID uuid.UUID) (*Subscription, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, organization_id, provider_subscription_id, plan, status, seats, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at
		 FROM subscriptions WHERE organization_id = $1`, orgID)
	return scanSubscription(row)
}

func (r *SQLRepository) UpsertSubscription(ctx context.Context, s *Subscription) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions (id, organization_id, provider_subscription_id, plan, status, seats, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (organization_id) DO UPDATE SET
		   provider_subscription_id = EXCLUDED.provider_subscription_id,
		   plan = EXCLUDED.plan,
		   status = EXCLUDED.status,
		   seats = EXCLUDED.seats,
		   current_period_start = EXCLUDED.current_period_start,
		   current_period_end = EXCLUDED.current_period_end,
		   cancel_at_period_end = EXCLUDED.cancel_at_period_end,
		   updated_at = EXCLUDED.updated_at`,
		s.ID, s.OrganizationID, s.ProviderSubscriptionID, s.Plan, s.Status, s.Seats, s.CurrentPeriodStart, s.CurrentPeriodEnd, s.CancelAtPeriodEnd, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *SQLRepository) ListInvoices(ctx context.Context, orgID uuid.UUID, limit int) ([]Invoice, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, organization_id, provider_invoice_id, amount_cents, currency, status, period_start, period_end, created_at, updated_at
		 FROM invoices WHERE organization_id = $1 ORDER BY created_at DESC LIMIT $2`,
		orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInvoices(rows)
}

func (r *SQLRepository) CreateInvoice(ctx context.Context, inv *Invoice) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO invoices (id, organization_id, provider_invoice_id, amount_cents, currency, status, period_start, period_end, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		inv.ID, inv.OrganizationID, inv.ProviderInvoiceID, inv.AmountCents, inv.Currency, inv.Status, inv.PeriodStart, inv.PeriodEnd, inv.CreatedAt, inv.UpdatedAt)
	return err
}

func (r *SQLRepository) RunInTx(ctx context.Context, fn func(Repository) error) error {
	beginner, ok := r.db.(db.BeginTxer)
	if !ok {
		return fmt.Errorf("database handle does not support transactions")
	}
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if tenantID, ok := db.TenantIDFromContext(ctx); ok {
		_, _ = tx.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", tenantID.String())
	}
	txRepo := &SQLRepository{db: tx}
	if err := fn(txRepo); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanSubscription(row rowScanner) (*Subscription, error) {
	var s Subscription
	var providerID sql.NullString
	var start, end sql.NullTime
	if err := row.Scan(&s.ID, &s.OrganizationID, &providerID, &s.Plan, &s.Status, &s.Seats, &start, &end, &s.CancelAtPeriodEnd, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	if providerID.Valid {
		s.ProviderSubscriptionID = providerID.String
	}
	if start.Valid {
		s.CurrentPeriodStart = &start.Time
	}
	if end.Valid {
		s.CurrentPeriodEnd = &end.Time
	}
	return &s, nil
}

func scanInvoice(row rowScanner) (*Invoice, error) {
	var inv Invoice
	var providerID sql.NullString
	var start, end sql.NullTime
	if err := row.Scan(&inv.ID, &inv.OrganizationID, &providerID, &inv.AmountCents, &inv.Currency, &inv.Status, &start, &end, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	if providerID.Valid {
		inv.ProviderInvoiceID = providerID.String
	}
	if start.Valid {
		inv.PeriodStart = &start.Time
	}
	if end.Valid {
		inv.PeriodEnd = &end.Time
	}
	return &inv, nil
}

func scanInvoices(rows *sql.Rows) ([]Invoice, error) {
	var list []Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *inv)
	}
	return list, rows.Err()
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
