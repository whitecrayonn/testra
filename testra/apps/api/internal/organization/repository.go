package organization

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, org *Organization) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		org.ID, org.Name, org.Slug, org.OwnerID, org.CreatedAt, org.UpdatedAt,
	)
	return err
}

func (r *SQLRepository) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	var org Organization
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, owner_id, created_at, updated_at FROM organizations WHERE id = $1`,
		id,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.CreatedAt, &org.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *SQLRepository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, owner_id, created_at, updated_at FROM organizations WHERE slug = $1`,
		slug,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.CreatedAt, &org.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *SQLRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Organization, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT o.id, o.name, o.slug, o.owner_id, o.created_at, o.updated_at
		 FROM organizations o
		 JOIN organization_members om ON o.id = om.organization_id
		 WHERE om.user_id = $1
		 ORDER BY o.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

func (r *SQLRepository) AddMember(ctx context.Context, member *Member) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO organization_members (organization_id, user_id, role, created_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (organization_id, user_id) DO NOTHING`,
		member.OrganizationID, member.UserID, member.Role, member.CreatedAt,
	)
	return err
}

func (r *SQLRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*Member, error) {
	var member Member
	err := r.db.QueryRowContext(ctx,
		`SELECT organization_id, user_id, role, created_at FROM organization_members
		 WHERE organization_id = $1 AND user_id = $2`,
		orgID, userID,
	).Scan(&member.OrganizationID, &member.UserID, &member.Role, &member.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sharederrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &member, nil
}
