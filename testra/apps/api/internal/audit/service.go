package audit

import (
	"context"

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
	_ = s.repo.Insert(ctx, event)
}
