package analytics

import (
	"time"

	"github.com/google/uuid"
)

type Dashboard struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Type        string
	Config      map[string]interface{}
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DailyMetric struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ProjectID   *uuid.UUID
	MetricDate  time.Time
	TotalRuns   int
	Passed      int
	Failed      int
	Skipped     int
	Blocked     int
	DurationMs  int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Summary struct {
	TotalRuns  int   `json:"total_runs"`
	Passed     int   `json:"passed"`
	Failed     int   `json:"failed"`
	Skipped    int   `json:"skipped"`
	Blocked    int   `json:"blocked"`
	DurationMs int64 `json:"duration_ms"`
}

type TrendPoint struct {
	Date       string `json:"date"`
	TotalRuns  int    `json:"total_runs"`
	Passed     int    `json:"passed"`
	Failed     int    `json:"failed"`
	Skipped    int    `json:"skipped"`
	Blocked    int    `json:"blocked"`
	DurationMs int64  `json:"duration_ms"`
}
