package project

import (
	"context"
	"testing"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type fakeRepository struct {
	projects map[uuid.UUID]*Project
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{projects: make(map[uuid.UUID]*Project)}
}

func (f *fakeRepository) Create(_ context.Context, project *Project) error {
	f.projects[project.ID] = project
	return nil
}

func (f *fakeRepository) GetByID(_ context.Context, id uuid.UUID) (*Project, error) {
	if p, ok := f.projects[id]; ok {
		return p, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) GetByKey(_ context.Context, workspaceID uuid.UUID, key string) (*Project, error) {
	for _, p := range f.projects {
		if p.WorkspaceID == workspaceID && p.Key == key {
			return p, nil
		}
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListForWorkspace(_ context.Context, workspaceID uuid.UUID) ([]Project, error) {
	var result []Project
	for _, p := range f.projects {
		if p.WorkspaceID == workspaceID {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (f *fakeRepository) ListForWorkspacePaginated(_ context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Project, error) {
	var result []Project
	for _, p := range f.projects {
		if p.WorkspaceID == workspaceID {
			result = append(result, *p)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func TestServiceCreate(t *testing.T) {
	workspaceID := uuid.New()

	tests := []struct {
		name    string
		input   CreateInput
		wantErr error
	}{
		{
			name:    "valid project",
			input:   CreateInput{WorkspaceID: workspaceID, Name: "Core Platform", Key: "CORE"},
			wantErr: nil,
		},
		{
			name:    "lowercase key is normalized",
			input:   CreateInput{WorkspaceID: workspaceID, Name: "Billing", Key: "bill"},
			wantErr: nil,
		},
		{
			name:    "missing name",
			input:   CreateInput{WorkspaceID: workspaceID, Name: "", Key: "NONAME"},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name:    "missing workspace",
			input:   CreateInput{WorkspaceID: uuid.Nil, Name: "Orphan", Key: "ORPH"},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name:    "invalid key characters",
			input:   CreateInput{WorkspaceID: workspaceID, Name: "Bad Key", Key: "bad-key!"},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name:    "key too short",
			input:   CreateInput{WorkspaceID: workspaceID, Name: "Short", Key: "A"},
			wantErr: sharederrors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(newFakeRepository())
			project, err := service.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if project.ID == uuid.Nil {
				t.Error("expected project ID to be set")
			}
		})
	}
}

func TestServiceCreateNormalizesKey(t *testing.T) {
	service := NewService(newFakeRepository())

	project, err := service.Create(context.Background(), CreateInput{
		WorkspaceID: uuid.New(),
		Name:        "Billing",
		Key:         " bill ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.Key != "BILL" {
		t.Errorf("expected key BILL, got %s", project.Key)
	}
}

func TestServiceCreateDuplicateKey(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(repo)
	workspaceID := uuid.New()

	if _, err := service.Create(context.Background(), CreateInput{WorkspaceID: workspaceID, Name: "First", Key: "CORE"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := service.Create(context.Background(), CreateInput{WorkspaceID: workspaceID, Name: "Second", Key: "CORE"})
	if err != sharederrors.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	// Same key in a different workspace is allowed.
	if _, err := service.Create(context.Background(), CreateInput{WorkspaceID: uuid.New(), Name: "Other", Key: "CORE"}); err != nil {
		t.Fatalf("expected no error for different workspace, got %v", err)
	}
}

func TestServiceGetAndList(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(repo)
	workspaceID := uuid.New()

	created, err := service.Create(context.Background(), CreateInput{WorkspaceID: workspaceID, Name: "Core", Key: "CORE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := service.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected project %s, got %s", created.ID, got.ID)
	}

	list, err := service.ListForWorkspace(context.Background(), workspaceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 project, got %d", len(list))
	}

	if _, err := service.Get(context.Background(), uuid.New()); err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
