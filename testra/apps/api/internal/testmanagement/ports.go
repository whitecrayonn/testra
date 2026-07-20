package testmanagement

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Folders
	CreateFolder(ctx context.Context, folder *TestFolder) error
	GetFolderByID(ctx context.Context, id uuid.UUID) (*TestFolder, error)
	ListFolders(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID, cursor string, limit int) ([]TestFolder, error)
	UpdateFolder(ctx context.Context, folder *TestFolder) error
	DeleteFolder(ctx context.Context, id uuid.UUID) error

	// Suites
	CreateSuite(ctx context.Context, suite *TestSuite) error
	GetSuiteByID(ctx context.Context, id uuid.UUID) (*TestSuite, error)
	ListSuites(ctx context.Context, workspaceID uuid.UUID, folderID *uuid.UUID, cursor string, limit int) ([]TestSuite, error)
	UpdateSuite(ctx context.Context, suite *TestSuite) error
	DeleteSuite(ctx context.Context, id uuid.UUID) error

	// Test Cases
	CreateCase(ctx context.Context, tc *TestCase) error
	GetCaseByID(ctx context.Context, id uuid.UUID) (*TestCase, error)
	ListCases(ctx context.Context, projectID uuid.UUID, suiteID *uuid.UUID, cursor string, limit int) ([]TestCase, error)
	SearchCases(ctx context.Context, workspaceID uuid.UUID, query string, cursor string, limit int) ([]TestCase, string, error)
	UpdateCase(ctx context.Context, tc *TestCase) error
	DeleteCase(ctx context.Context, id uuid.UUID) error

	// Versions
	CreateVersion(ctx context.Context, version *TestCaseVersion) error
	ListVersions(ctx context.Context, caseID uuid.UUID, cursor string, limit int) ([]TestCaseVersion, error)

	// Transaction support
	RunInTx(ctx context.Context, fn func(Repository) error) error
}
