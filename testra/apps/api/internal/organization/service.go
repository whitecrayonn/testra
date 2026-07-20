package organization

import (
	"context"
	"errors"
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
	Name  string
	Slug  string
	Owner uuid.UUID
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Organization, error) {
	if input.Name == "" {
		return nil, sharederrors.ErrInvalidInput
	}

	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if slug == "" {
		slug = validation.Slugify(input.Name)
	}

	existing, err := s.repo.GetBySlug(ctx, slug)
	if err != nil && !errors.Is(err, sharederrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, sharederrors.ErrConflict
	}

	org := &Organization{
		ID:        uuid.New(),
		Name:      input.Name,
		Slug:      slug,
		OwnerID:   input.Owner,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, org); err != nil {
		return nil, err
	}

	member := &Member{
		OrganizationID: org.ID,
		UserID:         input.Owner,
		Role:           "owner",
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Organization, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID) ([]Organization, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *Service) ListForUserPaginated(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]Organization, error) {
	return s.repo.ListForUserPaginated(ctx, userID, cursor, limit)
}

func (s *Service) AddMember(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	if role == "" {
		role = "member"
	}
	member := &Member{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		CreatedAt:      time.Now().UTC(),
	}
	return s.repo.AddMember(ctx, member)
}
