package intelligence

import (
	"context"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type Service struct {
	repo Repository
	ml   MLClient
}

func NewService(repo Repository, ml MLClient) *Service {
	if ml == nil {
		ml = NewMLClient("")
	}
	return &Service{repo: repo, ml: ml}
}

func (s *Service) PredictFlaky(ctx context.Context, input PredictFlakyInput, workspaceID uuid.UUID) (*FlakyPrediction, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	history := input.History
	var tcID *uuid.UUID
	if input.TestCaseID != "" {
		id, err := uuid.Parse(input.TestCaseID)
		if err != nil {
			return nil, sharederrors.ErrInvalidInput
		}
		tcID = &id
		if len(history) == 0 {
			h, err := s.repo.GetTestCaseHistory(ctx, id, 50)
			if err == nil {
				history = h
			}
		}
	}

	predInput := PredictionInput{
		TestCaseID:    input.TestCaseID,
		TestCaseTitle: input.TestCaseTitle,
		History:       history,
	}

	result, err := s.ml.PredictFlaky(ctx, predInput)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	p := &FlakyPrediction{
		ID:             uuid.New(),
		WorkspaceID:    workspaceID,
		TestCaseID:     tcID,
		TestCaseTitle:  input.TestCaseTitle,
		FlakinessScore: result.FlakinessScore,
		Confidence:     result.Confidence,
		Features: map[string]interface{}{
			"history_count": len(history),
			"explanation":   result.Explanation,
		},
		PredictedAt: now,
		LastSeenAt:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreatePrediction(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListPredictions(ctx context.Context, workspaceID uuid.UUID, minScore float64, limit int) ([]FlakyPrediction, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListPredictions(ctx, workspaceID, minScore, limit)
}

func (s *Service) GetPrediction(ctx context.Context, id uuid.UUID) (*FlakyPrediction, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetPrediction(ctx, id)
}

func (s *Service) ClassifyFailure(ctx context.Context, input ClassifyFailureInput, workspaceID uuid.UUID) (*ClassificationResult, error) {
	if workspaceID == uuid.Nil || !validation.IsValidName(input.ErrorMessage) {
		return nil, sharederrors.ErrInvalidInput
	}

	result, err := s.ml.ClassifyFailure(ctx, input.ErrorMessage, input.StackTrace)
	if err != nil {
		return nil, err
	}

	if err := s.repo.IncrementCluster(ctx, workspaceID, result.Label, input.ErrorMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *Service) ListClusters(ctx context.Context, workspaceID uuid.UUID, limit int) ([]FailureCluster, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListClusters(ctx, workspaceID, limit)
}

// Input structs

type PredictFlakyInput struct {
	TestCaseID    string
	TestCaseTitle string
	History       []RunHistoryPoint
}

type ClassifyFailureInput struct {
	ErrorMessage string
	StackTrace   string
}
