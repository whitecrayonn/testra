package audit

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type LogInput struct {
	UserID     uuid.UUID
	Action     string
	Resource   string
	ResourceID string
	IPAddress  string
	Metadata   map[string]string
}

func (s *Service) Log(ctx context.Context, input LogInput) {
	event := &Event{
		UserID:     input.UserID,
		Action:     input.Action,
		Resource:   input.Resource,
		ResourceID: input.ResourceID,
		IPAddress:  input.IPAddress,
		Metadata:   input.Metadata,
	}
	// Audit events are compliance evidence: never drop failures silently.
	// Persistence errors are logged so lost events are observable (C6).
	if err := s.repo.Insert(ctx, event); err != nil {
		log.Printf("audit: failed to persist event action=%q resource=%q resource_id=%q user_id=%s: %v",
			input.Action, input.Resource, input.ResourceID, input.UserID, err)
	}
}

// ListByUser returns a page of the given user's own audit trail, most recent
// first.
func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID, beforeCreatedAt *time.Time, beforeID uuid.UUID, limit int) ([]Event, error) {
	return s.repo.ListByUser(ctx, userID, beforeCreatedAt, beforeID, limit)
}
