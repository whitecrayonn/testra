package workspace

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	apidb "github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type Service struct {
	repo Repository
	db   apidb.BeginTxer
}

func NewService(repo Repository, db apidb.BeginTxer) *Service {
	return &Service{repo: repo, db: db}
}

type CreateInput struct {
	OrganizationID uuid.UUID
	Name           string
	Slug           string
	Description    string
	OwnerID        uuid.UUID
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Workspace, error) {
	if input.Name == "" {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.OrganizationID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if slug == "" {
		slug = validation.Slugify(input.Name)
	}

	existing, err := s.repo.GetBySlug(ctx, input.OrganizationID, slug)
	if err != nil && !errors.Is(err, sharederrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, sharederrors.ErrConflict
	}

	now := time.Now().UTC()
	workspace := &Workspace{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		Name:           input.Name,
		Slug:           slug,
		Description:    strings.TrimSpace(input.Description),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	member := &Member{
		WorkspaceID: workspace.ID,
		UserID:      input.OwnerID,
		Role:        "owner",
		CreatedAt:   now,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	workCtx := apidb.WithTx(ctx, tx)
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	if err := s.repo.Create(workCtx, workspace); err != nil {
		return nil, err
	}
	if err := s.repo.AddMember(workCtx, member); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	eventbus.Default().Publish(ctx, eventbus.Event{
		Type:     "workspace.created",
		TenantID: workspace.OrganizationID.String(),
		Payload: map[string]interface{}{
			"workspace_id":    workspace.ID.String(),
			"organization_id": workspace.OrganizationID.String(),
			"name":            workspace.Name,
			"slug":            workspace.Slug,
		},
	})

	return workspace, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListForOrganizationPaginated(ctx context.Context, orgID uuid.UUID, cursor string, limit int) ([]Workspace, error) {
	return s.repo.ListForOrganizationPaginated(ctx, orgID, cursor, limit)
}
