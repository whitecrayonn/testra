package results

import (
	"time"

	"github.com/google/uuid"
)

type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusPassed    RunStatus = "passed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusSkipped   RunStatus = "skipped"
	RunStatusCancelled RunStatus = "cancelled"
)

type RunItemStatus string

const (
	RunItemStatusPending RunItemStatus = "pending"
	RunItemStatusRunning RunItemStatus = "running"
	RunItemStatusPassed  RunItemStatus = "passed"
	RunItemStatusFailed  RunItemStatus = "failed"
	RunItemStatusSkipped RunItemStatus = "skipped"
	RunItemStatusBlocked RunItemStatus = "blocked"
)

type RunSource string

const (
	RunSourceManual RunSource = "manual"
	RunSourceCI     RunSource = "ci"
	RunSourceAPI    RunSource = "api"
)

type TestRun struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID
	SuiteID     *uuid.UUID
	Name        string
	Status      RunStatus
	Total       int
	Passed      int
	Failed      int
	Skipped     int
	Blocked     int
	DurationMs  int64
	Source      RunSource
	Metadata    map[string]interface{}
	CreatedBy   uuid.UUID
	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TestRunItem struct {
	ID           uuid.UUID
	RunID        uuid.UUID
	TestCaseID   *uuid.UUID
	Title        string
	Status       RunItemStatus
	DurationMs   int64
	ErrorMessage string
	StackTrace   string
	Artifacts    []string
	SortOrder    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func IsValidRunStatus(s string) bool {
	switch RunStatus(s) {
	case RunStatusPending, RunStatusRunning, RunStatusPassed, RunStatusFailed, RunStatusSkipped, RunStatusCancelled:
		return true
	}
	return false
}

func IsValidRunItemStatus(s string) bool {
	switch RunItemStatus(s) {
	case RunItemStatusPending, RunItemStatusRunning, RunItemStatusPassed, RunItemStatusFailed, RunItemStatusSkipped, RunItemStatusBlocked:
		return true
	}
	return false
}

func IsTerminalRunStatus(s RunStatus) bool {
	switch s {
	case RunStatusPassed, RunStatusFailed, RunStatusSkipped, RunStatusCancelled:
		return true
	}
	return false
}
