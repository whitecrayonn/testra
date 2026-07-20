package intelligence

import (
	"time"

	"github.com/google/uuid"
)

type FlakyPrediction struct {
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	TestCaseID     *uuid.UUID
	TestCaseTitle  string
	FlakinessScore float64
	Confidence     float64
	Features       map[string]interface{}
	PredictedAt    time.Time
	LastSeenAt     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type FailureCluster struct {
	ID           uuid.UUID
	WorkspaceID  uuid.UUID
	ClusterLabel string
	Pattern      string
	SampleError  string
	Count        int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PredictionInput struct {
	TestCaseID    string
	TestCaseTitle string
	History       []RunHistoryPoint
}

type RunHistoryPoint struct {
	Status     string `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	Date       string `json:"date"`
}

type PredictionResult struct {
	FlakinessScore float64 `json:"flakiness_score"`
	Confidence     float64 `json:"confidence"`
	Explanation    string  `json:"explanation"`
}

type ClassificationResult struct {
	Label       string  `json:"label"`
	Confidence  float64 `json:"confidence"`
	Explanation string  `json:"explanation"`
}
