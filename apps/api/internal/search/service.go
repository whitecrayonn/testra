package search

import (
	"context"
	"strings"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type Repository interface {
	Search(ctx context.Context, workspaceID uuid.UUID, query string, limit int) (*Result, error)
}

func (s *Service) Search(ctx context.Context, workspaceID uuid.UUID, query string, limit int) (*Result, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if strings.TrimSpace(query) == "" {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return s.repo.Search(ctx, workspaceID, query, limit)
}
