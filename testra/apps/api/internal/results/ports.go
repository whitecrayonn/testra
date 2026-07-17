package results

import (
	"context"

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
