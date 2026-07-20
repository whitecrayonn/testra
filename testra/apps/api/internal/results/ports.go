package results

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateRun(ctx context.Context, run *TestRun) error
	GetRunByID(ctx context.Context, id uuid.UUID) (*TestRun, error)
	ListRuns(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]TestRun, error)
	UpdateRun(ctx context.Context, run *TestRun) error
	DeleteRun(ctx context.Context, id uuid.UUID) error

	CreateItem(ctx context.Context, item *TestRunItem) error
	GetItemByID(ctx context.Context, id uuid.UUID) (*TestRunItem, error)
	ListItems(ctx context.Context, runID uuid.UUID) ([]TestRunItem, error)
	UpdateItem(ctx context.Context, item *TestRunItem) error
	DeleteItemsByRunID(ctx context.Context, runID uuid.UUID) error

	// Manual execution support.
	CreateItemExecution(ctx context.Context, item *TestRunItem) error
	ListItemsByRunPaged(ctx context.Context, runID uuid.UUID, status, search, cursor string, limit int) ([]TestRunItem, error)

	CreateItemHistory(ctx context.Context, history *RunItemHistory) error
	ListItemHistory(ctx context.Context, itemID uuid.UUID) ([]RunItemHistory, error)

	CreateEvidence(ctx context.Context, evidence *EvidenceRef) error
	ListEvidenceByItem(ctx context.Context, itemID uuid.UUID) ([]EvidenceRef, error)
	DeleteEvidence(ctx context.Context, id uuid.UUID) error

	CreateRunItemDefect(ctx context.Context, itemID, defectID uuid.UUID) error
	ListRunItemDefects(ctx context.Context, itemID uuid.UUID) ([]uuid.UUID, error)
	DeleteRunItemDefect(ctx context.Context, itemID, defectID uuid.UUID) error

	CreatePlan(ctx context.Context, plan *TestPlan) error
	GetPlanByID(ctx context.Context, id uuid.UUID) (*TestPlan, error)
	ListPlans(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]TestPlan, error)
	UpdatePlan(ctx context.Context, plan *TestPlan) error
	DeletePlan(ctx context.Context, id uuid.UUID) error
	CreatePlanItem(ctx context.Context, item *TestPlanItem) error
	ListPlanItems(ctx context.Context, planID uuid.UUID) ([]TestPlanItem, error)
	DeletePlanItemsByPlanID(ctx context.Context, planID uuid.UUID) error

	RunInTx(ctx context.Context, fn func(Repository) error) error
}

type CreateRunInput struct {
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID
	SuiteID     *uuid.UUID
	Name        string
	Source      RunSource
	CreatedBy   uuid.UUID
	TestCaseIDs []uuid.UUID
	PlanID      *uuid.UUID
}

type CreatePlanInput struct {
	WorkspaceID   uuid.UUID
	ProjectID     uuid.UUID
	SuiteID       *uuid.UUID
	Name          string
	Description   string
	Configuration map[string]interface{}
	CreatedBy     uuid.UUID
	TestCaseIDs   []uuid.UUID
}

type UpdatePlanInput struct {
	Name          string
	Description   string
	Status        TestPlanStatus
	Configuration map[string]interface{}
	TestCaseIDs   []uuid.UUID
}

type RunProgressEvent struct {
	RunID    uuid.UUID
	ItemID   uuid.UUID
	Status   string
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Blocked  int
	Progress float64
}

type RunItemHistory struct {
	ID          uuid.UUID
	RunItemID   uuid.UUID
	Status      RunItemStatus
	StepResults []StepResult
	Comment     string
	DurationMs  int64
	ExecutedBy  *uuid.UUID
	CreatedAt   time.Time
}

type ExecuteItemInput struct {
	Status       RunItemStatus
	StepResults  []StepResult
	Comment      string
	DurationMs   int64
	ErrorMessage string
	StackTrace   string
	ExecutedBy   uuid.UUID
}

type EvidenceInput struct {
	RunItemID   uuid.UUID
	StepOrder   int
	FileName    string
	ContentType string
	StoragePath string
	UploadedBy  uuid.UUID
}
