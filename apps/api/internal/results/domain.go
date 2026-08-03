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
	RunItemStatusPending     RunItemStatus = "pending"
	RunItemStatusRunning     RunItemStatus = "running"
	RunItemStatusPassed      RunItemStatus = "passed"
	RunItemStatusFailed      RunItemStatus = "failed"
	RunItemStatusSkipped     RunItemStatus = "skipped"
	RunItemStatusBlocked     RunItemStatus = "blocked"
	RunItemStatusRetest      RunItemStatus = "retest"
	RunItemStatusNotExecuted RunItemStatus = "not_executed"
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

type StepResult struct {
	Order      int        `json:"order"`
	Status     string     `json:"status"`
	Comment    string     `json:"comment"`
	DurationMs int64      `json:"duration_ms"`
	ExecutedBy *uuid.UUID `json:"executed_by,omitempty"`
	ExecutedAt *time.Time `json:"executed_at,omitempty"`
}

type EvidenceRef struct {
	ID          uuid.UUID  `json:"id"`
	RunItemID   uuid.UUID  `json:"run_item_id"`
	StepOrder   int        `json:"step_order"`
	FileName    string     `json:"file_name"`
	ContentType string     `json:"content_type"`
	StoragePath string     `json:"storage_path"`
	UploadedBy  *uuid.UUID `json:"uploaded_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
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
	StepResults  []StepResult
	Comment      string
	ExecutedBy   *uuid.UUID
	ExecutedAt   *time.Time
	SortOrder    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TestPlanStatus string

const (
	TestPlanStatusActive   TestPlanStatus = "active"
	TestPlanStatusArchived TestPlanStatus = "archived"
)

type TestPlan struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	ProjectID     uuid.UUID
	SuiteID       *uuid.UUID
	Name          string
	Description   string
	Status        TestPlanStatus
	Configuration map[string]interface{}
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TestPlanItem struct {
	ID         uuid.UUID
	PlanID     uuid.UUID
	TestCaseID uuid.UUID
	SortOrder  int
	CreatedAt  time.Time
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
	case RunItemStatusPending, RunItemStatusRunning, RunItemStatusPassed, RunItemStatusFailed, RunItemStatusSkipped, RunItemStatusBlocked, RunItemStatusRetest, RunItemStatusNotExecuted:
		return true
	}
	return false
}

func IsValidTestPlanStatus(s string) bool {
	switch TestPlanStatus(s) {
	case TestPlanStatusActive, TestPlanStatusArchived:
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
