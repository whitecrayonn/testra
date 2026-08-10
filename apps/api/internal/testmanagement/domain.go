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

// TestCaseSource distinguishes manually authored test cases from ones
// produced by generation. TestCaseSourceGeneratedSpec is deterministic,
// rule-based generation from an OpenAPI spec (no AI/ML — see
// docs/BIBLICAL_TESTRA.md's "No External LLM" principle).
// TestCaseSourceGeneratedFile is the one intentional, opt-in exception to
// that principle: it comes from the testgen.GenerateFromFile path, which
// calls an external LLM (Gemini) to turn an uploaded spreadsheet's rows into
// draft test cases. Both sources are always created as pending_review and
// require human approval before they count toward coverage.
type TestCaseSource string

const (
	TestCaseSourceManual        TestCaseSource = "manual"
	TestCaseSourceGeneratedSpec TestCaseSource = "generated_spec"
	TestCaseSourceGeneratedFile TestCaseSource = "generated_file"
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
