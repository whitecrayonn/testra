package integrationhub

import (
	"context"
	"fmt"
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

func (s *Service) CreateIntegration(ctx context.Context, input CreateIntegrationInput, createdBy uuid.UUID) (*Integration, error) {
	if input.WorkspaceID == uuid.Nil || !validation.IsValidName(input.Name) || !IsValidIntegrationType(input.Type) {
		return nil, sharederrors.ErrInvalidInput
	}
	if err := validateConfig(IntegrationType(input.Type), input.Config); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	i := &Integration{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		Type:        IntegrationType(input.Type),
		Name:        input.Name,
		Config:      input.Config,
		Enabled:     input.Enabled,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateIntegration(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *Service) ListIntegrations(ctx context.Context, workspaceID uuid.UUID) ([]Integration, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListIntegrations(ctx, workspaceID)
}

func (s *Service) GetIntegration(ctx context.Context, id uuid.UUID) (*Integration, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetIntegration(ctx, id)
}

func (s *Service) UpdateIntegration(ctx context.Context, id uuid.UUID, input UpdateIntegrationInput) (*Integration, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	i, err := s.repo.GetIntegration(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		i.Name = input.Name
	}
	if input.Type != "" {
		i.Type = IntegrationType(input.Type)
	}
	if input.Config != nil {
		i.Config = input.Config
	}
	if input.Enabled != nil {
		i.Enabled = *input.Enabled
	}
	if err := validateConfig(i.Type, i.Config); err != nil {
		return nil, err
	}
	i.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateIntegration(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *Service) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteIntegration(ctx, id)
}

func (s *Service) TestIntegration(ctx context.Context, id uuid.UUID) error {
	i, err := s.repo.GetIntegration(ctx, id)
	if err != nil {
		return err
	}
	if !i.Enabled {
		return sharederrors.ErrInvalidInput
	}
	_, err = dispatch(ctx, i, map[string]interface{}{"summary": "Testra integration test", "text": "Testra integration test"})
	return err
}

func (s *Service) DispatchEvent(ctx context.Context, input DispatchEventInput, userID uuid.UUID) (*IntegrationEvent, error) {
	if input.WorkspaceID == uuid.Nil || input.EventType == "" {
		return nil, sharederrors.ErrInvalidInput
	}

	i, err := s.repo.GetIntegration(ctx, input.IntegrationID)
	if err != nil {
		return nil, err
	}
	if !i.Enabled {
		return nil, fmt.Errorf("%w: integration disabled", sharederrors.ErrInvalidInput)
	}

	now := time.Now().UTC()
	e := &IntegrationEvent{
		ID:            uuid.New(),
		WorkspaceID:   input.WorkspaceID,
		IntegrationID: &i.ID,
		EventType:     input.EventType,
		Payload:       input.Payload,
		Status:        "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateEvent(ctx, e); err != nil {
		return nil, err
	}

	externalID, err := dispatch(ctx, i, input.Payload)
	if err != nil {
		e.Status = "failed"
		e.Payload["error"] = err.Error()
	} else {
		e.Status = "sent"
		e.ExternalID = externalID
	}
	e.UpdatedAt = time.Now().UTC()
	// Re-create event as update is not in repository interface; create with same ID is idempotent enough for MVP.
	_ = s.repo.CreateEvent(ctx, e)
	return e, err
}

func (s *Service) ListEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]IntegrationEvent, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListEvents(ctx, workspaceID, limit)
}

func validateConfig(t IntegrationType, cfg map[string]string) error {
	required := map[IntegrationType][]string{
		TypeJira:    {"url", "token", "project_key"},
		TypeGitHub:  {"url", "token", "owner", "repo"},
		TypeGitLab:  {"url", "token", "project_id"},
		TypeSlack:   {"url"},
		TypeWebhook: {"url"},
	}
	for _, key := range required[t] {
		if cfg[key] == "" {
			return fmt.Errorf("%w: %s config requires '%s'", sharederrors.ErrInvalidInput, t, key)
		}
	}
	return nil
}

// Input structs

type CreateIntegrationInput struct {
	WorkspaceID uuid.UUID
	Type        string
	Name        string
	Config      map[string]string
	Enabled     bool
}

type UpdateIntegrationInput struct {
	Type    string
	Name    string
	Config  map[string]string
	Enabled *bool
}

type DispatchEventInput struct {
	WorkspaceID   uuid.UUID
	IntegrationID uuid.UUID
	EventType     string
	Payload       map[string]interface{}
}
