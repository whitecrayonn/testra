package workspace

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	OrganizationID uuid.UUID
	Name           string
	Slug           string
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
	if err != nil && err != sharederrors.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, sharederrors.ErrConflict
	}

	workspace := &Workspace{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		Name:           input.Name,
		Slug:           slug,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	member := &Member{
		WorkspaceID: workspace.ID,
		UserID:      input.OwnerID,
		Role:        "owner",
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListForOrganization(ctx context.Context, orgID uuid.UUID) ([]Workspace, error) {
	return s.repo.ListForOrganization(ctx, orgID)
}

func (s *Service) ListForOrganizationPaginated(ctx context.Context, orgID uuid.UUID, cursor string, limit int) ([]Workspace, error) {
	return s.repo.ListForOrganizationPaginated(ctx, orgID, cursor, limit)
}

func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID) ([]Workspace, error) {
	return s.repo.ListForUser(ctx, userID)
}
