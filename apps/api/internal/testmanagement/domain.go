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
	TestCaseStatusDraft         TestCaseStatus = "draft"
	TestCaseStatusActive        TestCaseStatus = "active"
	TestCaseStatusDeprecated    TestCaseStatus = "deprecated"
	TestCaseStatusPendingReview TestCaseStatus = "pending_review"
)

// TestCaseSource distinguishes manually authored test cases from ones produced
// by deterministic generation. Per docs/AI_MEMORY.md ("No external LLM
// processing"), generation is rule-based only — there is no "generated_text"
// source, since free-text requirement parsing is not possible without a model.
type TestCaseSource string

const (
	TestCaseSourceManual        TestCaseSource = "manual"
	TestCaseSourceGeneratedSpec TestCaseSource = "generated_spec"
)

type TestCasePriority string

const (
	TestCasePriorityLow      TestCasePriority = "low"
	TestCasePriorityMedium   TestCasePriority = "medium"
	TestCasePriorityHigh     TestCasePriority = "high"
	TestCasePriorityCritical TestCasePriority = "critical"
)

type TestCase struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	ProjectID       uuid.UUID
	SuiteID         *uuid.UUID
	Title           string
	Description     string
	Preconditions   string
	Steps           []TestStep
	Status          TestCaseStatus
	Priority        TestCasePriority
	Tags            []string
	Version         int
	CreatedBy       uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Source          TestCaseSource
	GenerationRunID *uuid.UUID
	ReviewedBy      *uuid.UUID
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
