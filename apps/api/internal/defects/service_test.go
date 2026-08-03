package defects

import (
	"context"
	"testing"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type fakeRepository struct {
	defects map[uuid.UUID]*Defect
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{defects: make(map[uuid.UUID]*Defect)}
}

func (f *fakeRepository) Create(_ context.Context, d *Defect) error {
	f.defects[d.ID] = d
	return nil
}

func (f *fakeRepository) GetByID(_ context.Context, id uuid.UUID) (*Defect, error) {
	if d, ok := f.defects[id]; ok {
		return d, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListByProject(_ context.Context, projectID uuid.UUID, cursor string, limit int) ([]Defect, error) {
	var result []Defect
	for _, d := range f.defects {
		if d.ProjectID == projectID {
			result = append(result, *d)
		}
	}
	return result, nil
}

func (f *fakeRepository) Update(_ context.Context, d *Defect) error {
	if _, ok := f.defects[d.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.defects[d.ID] = d
	return nil
}

func (f *fakeRepository) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := f.defects[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.defects, id)
	return nil
}

func TestServiceCreate(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	userID := uuid.New()

	svc := NewService(newFakeRepository())

	tests := []struct {
		name    string
		input   CreateInput
		wantErr error
	}{
		{
			name: "valid defect",
			input: CreateInput{
				WorkspaceID: wsID,
				ProjectID:   projID,
				Title:       "Login fails",
				Description: "User cannot login",
				CreatedBy:   userID,
			},
			wantErr: nil,
		},
		{
			name:    "missing workspace",
			input:   CreateInput{ProjectID: projID, Title: "Bug", CreatedBy: userID},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name:    "missing project",
			input:   CreateInput{WorkspaceID: wsID, Title: "Bug", CreatedBy: userID},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name:    "missing title",
			input:   CreateInput{WorkspaceID: wsID, ProjectID: projID, CreatedBy: userID},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name:    "missing created by",
			input:   CreateInput{WorkspaceID: wsID, ProjectID: projID, Title: "Bug"},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name: "invalid status",
			input: CreateInput{
				WorkspaceID: wsID,
				ProjectID:   projID,
				Title:       "Bug",
				CreatedBy:   userID,
				Status:      DefectStatus("invalid"),
			},
			wantErr: sharederrors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.ID == uuid.Nil {
				t.Error("expected defect ID")
			}
			if d.Status != DefectStatusOpen {
				t.Errorf("expected default status open, got %s", d.Status)
			}
		})
	}
}

func TestServiceUpdateAndDelete(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)
	wsID := uuid.New()
	projID := uuid.New()
	userID := uuid.New()

	d, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Title:       "Title",
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resolved := DefectStatusResolved
	updated, err := svc.Update(context.Background(), d.ID, UpdateInput{
		Title:  strPtr("Updated title"),
		Status: &resolved,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "Updated title" {
		t.Errorf("expected title 'Updated title', got %s", updated.Title)
	}
	if updated.Status != DefectStatusResolved {
		t.Errorf("expected status resolved, got %s", updated.Status)
	}

	if err := svc.Delete(context.Background(), d.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := svc.Delete(context.Background(), d.ID); err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestServiceListRequiresProject(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.List(context.Background(), uuid.Nil, "", 10)
	if err != sharederrors.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestServiceGetNotFound(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.Get(context.Background(), uuid.New())
	if err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceUpdateClearsOptionalRelations(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)
	wsID := uuid.New()
	projID := uuid.New()
	userID := uuid.New()
	assigned := uuid.New()
	itemID := uuid.New()

	d, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID:   wsID,
		ProjectID:     projID,
		Title:         "Bug",
		CreatedBy:     userID,
		AssignedTo:    &assigned,
		TestRunItemID: &itemID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.AssignedTo == nil || *d.AssignedTo != assigned {
		t.Fatal("expected assigned user to be set")
	}
	if d.TestRunItemID == nil || *d.TestRunItemID != itemID {
		t.Fatal("expected test run item to be set")
	}

	nilUUID := uuid.Nil
	updated, err := svc.Update(context.Background(), d.ID, UpdateInput{
		Title:         strPtr("Bug updated"),
		AssignedTo:    &nilUUID,
		TestRunItemID: &nilUUID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.AssignedTo != nil {
		t.Errorf("expected assigned_to to be cleared, got %v", updated.AssignedTo)
	}
	if updated.TestRunItemID != nil {
		t.Errorf("expected test_run_item_id to be cleared, got %v", updated.TestRunItemID)
	}
}

func strPtr(s string) *string { return &s }
