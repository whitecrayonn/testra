package project

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
)

var keyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Key         string
	Description string
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Project, error) {
	if input.Name == "" || input.WorkspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	key := strings.ToUpper(strings.TrimSpace(input.Key))
	if !keyPattern.MatchString(key) {
		return nil, sharederrors.ErrInvalidInput
	}

	existing, err := s.repo.GetByKey(ctx, input.WorkspaceID, key)
	if err != nil && !errors.Is(err, sharederrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, sharederrors.ErrConflict
	}

	now := time.Now().UTC()
	project := &Project{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Key:         key,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, err
	}

	eventbus.Default().Publish(ctx, eventbus.Event{
		Type:     "project.created",
		TenantID: project.WorkspaceID.String(),
		Payload: map[string]interface{}{
			"project_id":   project.ID.String(),
			"workspace_id": project.WorkspaceID.String(),
			"key":          project.Key,
			"name":         project.Name,
		},
	})

	return project, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Project, error) {
	return s.repo.ListForWorkspace(ctx, workspaceID)
}

func (s *Service) ListForWorkspacePaginated(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Project, error) {
	return s.repo.ListForWorkspacePaginated(ctx, workspaceID, cursor, limit)
}
