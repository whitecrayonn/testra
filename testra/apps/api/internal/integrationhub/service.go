package integrationhub

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/testra/testra/apps/api/internal/audit"
	"github.com/testra/testra/apps/api/internal/queue"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

// Service orchestrates integrations, dispatch, health checks and audit logging.
type Service struct {
	repo  Repository
	audit *audit.Service
	bus   *eventbus.Bus
	db    *sql.DB
}

// NewService creates an integration hub service.
func NewService(repo Repository, auditSvc *audit.Service, bus *eventbus.Bus, db *sql.DB) *Service {
	return &Service{repo: repo, audit: auditSvc, bus: bus, db: db}
}

func (s *Service) CreateIntegration(ctx context.Context, input CreateIntegrationInput, createdBy uuid.UUID) (*Integration, error) {
	if input.WorkspaceID == uuid.Nil || !validation.IsValidName(input.Name) || !IsValidIntegrationType(input.Type) {
		return nil, sharederrors.ErrInvalidInput
	}
	if err := validateProviderConfig(IntegrationType(input.Type), input.Config); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	i := &Integration{
		ID:           uuid.New(),
		WorkspaceID:  input.WorkspaceID,
		Type:         IntegrationType(input.Type),
		Name:         input.Name,
		Config:       input.Config,
		Enabled:      input.Enabled,
		HealthStatus: "unknown",
		SyncStatus:   "pending",
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateIntegration(ctx, i); err != nil {
		return nil, err
	}
	s.logAudit(ctx, createdBy, "integration.create", i.ID.String(), map[string]string{"type": string(i.Type), "name": i.Name})
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

func (s *Service) UpdateIntegration(ctx context.Context, id uuid.UUID, input UpdateIntegrationInput, userID uuid.UUID) (*Integration, error) {
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
	if err := validateProviderConfig(i.Type, i.Config); err != nil {
		return nil, err
	}
	i.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateIntegration(ctx, i); err != nil {
		return nil, err
	}
	s.logAudit(ctx, userID, "integration.update", i.ID.String(), map[string]string{"type": string(i.Type), "enabled": fmt.Sprintf("%v", i.Enabled)})
	return i, nil
}

func (s *Service) DeleteIntegration(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	s.logAudit(ctx, userID, "integration.delete", id.String(), nil)
	return s.repo.DeleteIntegration(ctx, id)
}

func (s *Service) TestIntegration(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	i, err := s.repo.GetIntegration(ctx, id)
	if err != nil {
		return err
	}
	if !i.Enabled {
		return fmt.Errorf("%w: integration disabled", sharederrors.ErrInvalidInput)
	}
	now := time.Now().UTC()
	i.LastTestedAt = &now
	externalID, err := testIntegration(ctx, i)
	if err != nil {
		i.HealthStatus = "error"
		i.LastError = err.Error()
		i.SyncStatus = "error"
		_ = s.repo.UpdateIntegration(ctx, i)
		s.logAudit(ctx, userID, "integration.test_failed", i.ID.String(), map[string]string{"error": err.Error()})
		return err
	}
	i.HealthStatus = "healthy"
	i.LastError = ""
	i.SyncStatus = "synced"
	i.RetryCount = 0
	_ = s.repo.UpdateIntegration(ctx, i)
	s.logAudit(ctx, userID, "integration.test_succeeded", i.ID.String(), map[string]string{"external_id": externalID})
	return nil
}

func (s *Service) HealthStatus(ctx context.Context, id uuid.UUID) (string, error) {
	i, err := s.repo.GetIntegration(ctx, id)
	if err != nil {
		return "", err
	}
	if !i.Enabled {
		return "disabled", nil
	}
	status, err := healthIntegration(ctx, i)
	if err != nil {
		return "error", err
	}
	return status, nil
}

func (s *Service) EnableIntegration(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	enabled := true
	_, err := s.UpdateIntegration(ctx, id, UpdateIntegrationInput{Enabled: &enabled}, userID)
	return err
}

func (s *Service) DisableIntegration(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	enabled := false
	_, err := s.UpdateIntegration(ctx, id, UpdateIntegrationInput{Enabled: &enabled}, userID)
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

	externalID, sendErr := dispatch(ctx, i, input.Payload)
	e.UpdatedAt = time.Now().UTC()
	if sendErr != nil {
		e.Status = "failed"
		e.Payload["error"] = sendErr.Error()
		_ = s.repo.UpdateEvent(ctx, e)
		s.enqueueRetry(ctx, e, i)
		s.logAudit(ctx, userID, "integration.dispatch_failed", i.ID.String(), map[string]string{"event_id": e.ID.String(), "error": sendErr.Error()})
	} else {
		e.Status = "sent"
		e.ExternalID = externalID
		_ = s.repo.UpdateEvent(ctx, e)
		s.logAudit(ctx, userID, "integration.dispatch_succeeded", i.ID.String(), map[string]string{"event_id": e.ID.String(), "external_id": externalID})
	}

	if s.bus != nil {
		s.bus.Publish(ctx, eventbus.Event{
			Type:     "integration.dispatched",
			Payload:  map[string]interface{}{"integration_id": i.ID.String(), "event_id": e.ID.String(), "status": e.Status},
			TenantID: input.WorkspaceID.String(),
		})
	}
	return e, sendErr
}

func (s *Service) RetryEvent(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) (*IntegrationEvent, error) {
	e, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	e.RetryCount++
	e.Status = "pending"
	e.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateEvent(ctx, e); err != nil {
		return nil, err
	}

	i, err := s.repo.GetIntegration(ctx, *e.IntegrationID)
	if err != nil {
		return nil, err
	}
	if !i.Enabled {
		e.Status = "failed"
		e.Payload["error"] = "integration disabled"
		_ = s.repo.UpdateEvent(ctx, e)
		return e, fmt.Errorf("%w: integration disabled", sharederrors.ErrInvalidInput)
	}

	externalID, sendErr := dispatch(ctx, i, e.Payload)
	e.UpdatedAt = time.Now().UTC()
	if sendErr != nil {
		e.Status = "failed"
		e.Payload["error"] = sendErr.Error()
		_ = s.repo.UpdateEvent(ctx, e)
		s.enqueueRetry(ctx, e, i)
		s.logAudit(ctx, userID, "integration.retry_failed", i.ID.String(), map[string]string{"event_id": e.ID.String(), "error": sendErr.Error()})
	} else {
		e.Status = "sent"
		e.ExternalID = externalID
		e.Payload["error"] = nil
		_ = s.repo.UpdateEvent(ctx, e)
		s.logAudit(ctx, userID, "integration.retry_succeeded", i.ID.String(), map[string]string{"event_id": e.ID.String(), "external_id": externalID})
	}
	return e, sendErr
}

func (s *Service) ListEvents(ctx context.Context, workspaceID uuid.UUID, status string, limit int) ([]IntegrationEvent, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListEvents(ctx, workspaceID, status, limit)
}

func (s *Service) GetEvent(ctx context.Context, eventID uuid.UUID) (*IntegrationEvent, error) {
	if eventID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetEvent(ctx, eventID)
}

func (s *Service) ListDeadLetterEvents(ctx context.Context, workspaceID uuid.UUID, limit int) ([]IntegrationEvent, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListEventsByStatus(ctx, workspaceID, "dead_letter", limit)
}

func (s *Service) ReplayDeadLetter(ctx context.Context, eventID uuid.UUID, userID uuid.UUID) (*IntegrationEvent, error) {
	e, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if e.Status != "dead_letter" {
		return nil, fmt.Errorf("%w: event is not in dead-letter state", sharederrors.ErrInvalidInput)
	}
	return s.RetryEvent(ctx, eventID, userID)
}

func (s *Service) ProcessIncomingWebhook(ctx context.Context, provider IntegrationType, integrationID uuid.UUID, signature string, body []byte) (*IntegrationEvent, error) {
	p, err := ProviderFor(provider)
	if err != nil {
		return nil, err
	}

	var i *Integration
	if integrationID != uuid.Nil {
		i, err = s.repo.GetIntegration(ctx, integrationID)
		if err != nil {
			return nil, err
		}
	}

	if vp, ok := p.(WebhookVerifier); ok {
		if err := vp.VerifyWebhook(i, body, signature); err != nil {
			return nil, err
		}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("%w: invalid webhook payload: %v", sharederrors.ErrInvalidInput, err)
	}

	now := time.Now().UTC()
	e := &IntegrationEvent{
		ID:          uuid.New(),
		WorkspaceID: uuid.Nil, // filled by route if available; otherwise anonymous
		EventType:   string(provider) + ".incoming",
		Payload:     payload,
		Status:      "received",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if i != nil {
		e.IntegrationID = &i.ID
		e.WorkspaceID = i.WorkspaceID
	}
	if err := s.repo.CreateEvent(ctx, e); err != nil {
		return nil, err
	}
	if s.bus != nil {
		s.bus.Publish(ctx, eventbus.Event{
			Type:    "integration.incoming_webhook",
			Payload: map[string]interface{}{"provider": string(provider), "event_id": e.ID.String()},
		})
	}
	return e, nil
}

func (s *Service) enqueueRetry(ctx context.Context, e *IntegrationEvent, i *Integration) {
	if s.db == nil {
		return
	}
	if e.RetryCount >= 3 {
		e.Status = "dead_letter"
		e.UpdatedAt = time.Now().UTC()
		_ = s.repo.UpdateEvent(ctx, e)
		return
	}
	payload := map[string]interface{}{
		"workspace_id":   e.WorkspaceID.String(),
		"integration_id": i.ID.String(),
		"event_id":       e.ID.String(),
		"event_type":     e.EventType,
		"payload":        e.Payload,
	}
	if err := queue.Enqueue(ctx, s.db, e.WorkspaceID, "default", "integration:retry", payload); err != nil {
		// Log and continue; the event is already marked failed.
		if s.audit != nil {
			s.audit.Log(ctx, audit.LogInput{Action: "integration.retry_enqueue_failed", Resource: "integration_event", ResourceID: e.ID.String()})
		}
	}
}

func (s *Service) logAudit(ctx context.Context, userID uuid.UUID, action, resourceID string, metadata map[string]string) {
	if s.audit == nil {
		return
	}
	s.audit.Log(ctx, audit.LogInput{
		UserID:     userID,
		Action:     action,
		Resource:   "integration",
		ResourceID: resourceID,
		Metadata:   metadata,
	})
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
