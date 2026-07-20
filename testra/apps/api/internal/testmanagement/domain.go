package testmanagement

import (
	"time"

	"github.com/google/uuid"
)

type TestFolder struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TestSuite struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	FolderID    *uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TestCaseStatus string

const (
	TestCaseStatusDraft      TestCaseStatus = "draft"
	TestCaseStatusActive     TestCaseStatus = "active"
	TestCaseStatusDeprecated TestCaseStatus = "deprecated"
)

type TestCasePriority string

const (
	TestCasePriorityLow      TestCasePriority = "low"
	TestCasePriorityMedium   TestCasePriority = "medium"
	TestCasePriorityHigh     TestCasePriority = "high"
	TestCasePriorityCritical TestCasePriority = "critical"
)

type TestCase struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	ProjectID     uuid.UUID
	SuiteID       *uuid.UUID
	Title         string
	Description   string
	Preconditions string
	Steps         []TestStep
	Status        TestCaseStatus
	Priority      TestCasePriority
	Tags          []string
	Version       int
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TestStep struct {
	Order    int
	Action   string
	Expected string
	TestData string
}

type TestCaseVersion struct {
	ID            uuid.UUID
	TestCaseID    uuid.UUID
	Version       int
	Title         string
	Description   string
	Preconditions string
	Steps         []TestStep
	ChangedBy     uuid.UUID
	CreatedAt     time.Time
}
