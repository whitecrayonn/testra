package automationhub

import (
	"context"

	"github.com/google/uuid"
)

// Repository persists automation hub entities.
type Repository interface {
	CreateProject(ctx context.Context, p *AutomationProject) error
	GetProject(ctx context.Context, id uuid.UUID) (*AutomationProject, error)
	ListProjects(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]AutomationProject, error)
	UpdateProject(ctx context.Context, p *AutomationProject) error
	DeleteProject(ctx context.Context, id uuid.UUID) error

	CreateExecution(ctx context.Context, e *AutomationExecution) error
	GetExecution(ctx context.Context, id uuid.UUID) (*AutomationExecution, error)
	ListExecutions(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]AutomationExecution, error)
	ListExecutionsByWorkspace(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]AutomationExecution, error)
	UpdateExecution(ctx context.Context, e *AutomationExecution) error
	DeleteExecution(ctx context.Context, id uuid.UUID) error

	CreateArtifact(ctx context.Context, a *AutomationArtifact) error
	GetArtifact(ctx context.Context, id uuid.UUID) (*AutomationArtifact, error)
	ListArtifacts(ctx context.Context, executionID uuid.UUID, cursor string, limit int) ([]AutomationArtifact, error)
	DeleteArtifact(ctx context.Context, id uuid.UUID) error

	CreateLog(ctx context.Context, l *AutomationLog) error
	ListLogs(ctx context.Context, executionID uuid.UUID, cursor string, limit int) ([]AutomationLog, error)

	RunInTx(ctx context.Context, fn func(Repository) error) error
}
