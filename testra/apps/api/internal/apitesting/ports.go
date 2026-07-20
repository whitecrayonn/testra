package apitesting

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the data access contract for the API testing domain.
type Repository interface {
	// Collections
	CreateCollection(ctx context.Context, c *Collection) error
	GetCollectionByID(ctx context.Context, id uuid.UUID) (*Collection, error)
	ListCollections(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Collection, error)
	UpdateCollection(ctx context.Context, c *Collection) error
	DeleteCollection(ctx context.Context, id uuid.UUID) error

	// Folders
	CreateFolder(ctx context.Context, f *Folder) error
	GetFolderByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	ListFolders(ctx context.Context, collectionID uuid.UUID, parentID *uuid.UUID, cursor string, limit int) ([]Folder, error)
	UpdateFolder(ctx context.Context, f *Folder) error
	DeleteFolder(ctx context.Context, id uuid.UUID) error

	// Environments
	CreateEnvironment(ctx context.Context, env *Environment) error
	GetEnvironmentByID(ctx context.Context, id uuid.UUID) (*Environment, error)
	ListEnvironments(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Environment, error)
	UpdateEnvironment(ctx context.Context, env *Environment) error
	DeleteEnvironment(ctx context.Context, id uuid.UUID) error

	// Requests
	CreateRequest(ctx context.Context, req *Request) error
	GetRequestByID(ctx context.Context, id uuid.UUID) (*Request, error)
	ListRequests(ctx context.Context, collectionID uuid.UUID, folderID *uuid.UUID, cursor string, limit int) ([]Request, error)
	SearchRequests(ctx context.Context, workspaceID uuid.UUID, query string, cursor string, limit int) ([]Request, string, error)
	UpdateRequest(ctx context.Context, req *Request) error
	DeleteRequest(ctx context.Context, id uuid.UUID) error

	// Execution history
	CreateRequestHistory(ctx context.Context, h *RequestHistory) error
	GetRequestHistoryByID(ctx context.Context, id uuid.UUID) (*RequestHistory, error)
	ListRequestHistory(ctx context.Context, requestID *uuid.UUID, workspaceID uuid.UUID, cursor string, limit int) ([]RequestHistory, error)

	// Transactions
	RunInTx(ctx context.Context, fn func(Repository) error) error
}
