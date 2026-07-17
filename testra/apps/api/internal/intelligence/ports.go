package intelligence

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreatePrediction(ctx context.Context, p *FlakyPrediction) error
	ListPredictions(ctx context.Context, workspaceID uuid.UUID, minScore float64, limit int) ([]FlakyPrediction, error)
	GetPrediction(ctx context.Context, id uuid.UUID) (*FlakyPrediction, error)

	CreateCluster(ctx context.Context, c *FailureCluster) error
	ListClusters(ctx context.Context, workspaceID uuid.UUID, limit int) ([]FailureCluster, error)
	IncrementCluster(ctx context.Context, workspaceID uuid.UUID, label string, sampleError string) error

	GetTestCaseHistory(ctx context.Context, testCaseID uuid.UUID, limit int) ([]RunHistoryPoint, error)
}

type MLClient interface {
	PredictFlaky(ctx context.Context, input PredictionInput) (PredictionResult, error)
	ClassifyFailure(ctx context.Context, errorMessage string, stackTrace string) (ClassificationResult, error)
}
