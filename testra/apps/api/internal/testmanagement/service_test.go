package testmanagement

import (
	"context"
	"testing"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type fakeRepository struct {
	folders  map[uuid.UUID]*TestFolder
	suites   map[uuid.UUID]*TestSuite
	cases    map[uuid.UUID]*TestCase
	versions []TestCaseVersion
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		folders: make(map[uuid.UUID]*TestFolder),
		suites:  make(map[uuid.UUID]*TestSuite),
		cases:   make(map[uuid.UUID]*TestCase),
	}
}

func (f *fakeRepository) CreateFolder(_ context.Context, folder *TestFolder) error {
	f.folders[folder.ID] = folder
	return nil
}

func (f *fakeRepository) GetFolderByID(_ context.Context, id uuid.UUID) (*TestFolder, error) {
	if folder, ok := f.folders[id]; ok {
		return folder, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListFolders(_ context.Context, workspaceID uuid.UUID, parentID *uuid.UUID) ([]TestFolder, error) {
	var result []TestFolder
	for _, folder := range f.folders {
		if folder.WorkspaceID == workspaceID {
			if parentID == nil && folder.ParentID == nil {
				result = append(result, *folder)
			} else if parentID != nil && folder.ParentID != nil && *folder.ParentID == *parentID {
				result = append(result, *folder)
			}
		}
	}
	return result, nil
}

func (f *fakeRepository) UpdateFolder(_ context.Context, folder *TestFolder) error {
	if _, ok := f.folders[folder.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.folders[folder.ID] = folder
	return nil
}

func (f *fakeRepository) DeleteFolder(_ context.Context, id uuid.UUID) error {
	if _, ok := f.folders[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.folders, id)
	return nil
}

func (f *fakeRepository) CreateSuite(_ context.Context, suite *TestSuite) error {
	f.suites[suite.ID] = suite
	return nil
}

func (f *fakeRepository) GetSuiteByID(_ context.Context, id uuid.UUID) (*TestSuite, error) {
	if suite, ok := f.suites[id]; ok {
		return suite, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListSuites(_ context.Context, workspaceID uuid.UUID, folderID *uuid.UUID) ([]TestSuite, error) {
	var result []TestSuite
	for _, suite := range f.suites {
		if suite.WorkspaceID == workspaceID {
			result = append(result, *suite)
		}
	}
	return result, nil
}

func (f *fakeRepository) UpdateSuite(_ context.Context, suite *TestSuite) error {
	if _, ok := f.suites[suite.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.suites[suite.ID] = suite
	return nil
}

func (f *fakeRepository) DeleteSuite(_ context.Context, id uuid.UUID) error {
	if _, ok := f.suites[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.suites, id)
	return nil
}

func (f *fakeRepository) CreateCase(_ context.Context, tc *TestCase) error {
	f.cases[tc.ID] = tc
	return nil
}

func (f *fakeRepository) GetCaseByID(_ context.Context, id uuid.UUID) (*TestCase, error) {
	if tc, ok := f.cases[id]; ok {
		return tc, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListCases(_ context.Context, projectID uuid.UUID, suiteID *uuid.UUID, cursor string, limit int) ([]TestCase, error) {
	var result []TestCase
	for _, tc := range f.cases {
		if tc.ProjectID == projectID {
			result = append(result, *tc)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (f *fakeRepository) SearchCases(_ context.Context, workspaceID uuid.UUID, query string, cursor string, limit int) ([]TestCase, string, error) {
	var result []TestCase
	for _, tc := range f.cases {
		if tc.WorkspaceID == workspaceID {
			result = append(result, *tc)
		}
	}
	return result, "", nil
}

func (f *fakeRepository) RunInTx(_ context.Context, fn func(Repository) error) error {
	return fn(f)
}

func (f *fakeRepository) UpdateCase(_ context.Context, tc *TestCase) error {
	if _, ok := f.cases[tc.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.cases[tc.ID] = tc
	return nil
}

func (f *fakeRepository) DeleteCase(_ context.Context, id uuid.UUID) error {
	if _, ok := f.cases[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.cases, id)
	return nil
}

func (f *fakeRepository) CreateVersion(_ context.Context, version *TestCaseVersion) error {
	f.versions = append(f.versions, *version)
	return nil
}

func (f *fakeRepository) ListVersions(_ context.Context, caseID uuid.UUID) ([]TestCaseVersion, error) {
	var result []TestCaseVersion
	for _, v := range f.versions {
		if v.TestCaseID == caseID {
			result = append(result, v)
		}
	}
	return result, nil
}

func TestServiceCreateFolder(t *testing.T) {
	wsID := uuid.New()

	tests := []struct {
		name    string
		input   CreateFolderInput
		wantErr error
	}{
		{
			name:    "valid folder",
			input:   CreateFolderInput{WorkspaceID: wsID, Name: "Root Folder"},
			wantErr: nil,
		},
		{
			name:    "empty name",
			input:   CreateFolderInput{WorkspaceID: wsID, Name: ""},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name:    "missing workspace",
			input:   CreateFolderInput{WorkspaceID: uuid.Nil, Name: "Folder"},
			wantErr: sharederrors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(newFakeRepository())
			folder, err := service.CreateFolder(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if folder.ID == uuid.Nil {
				t.Error("expected folder ID to be set")
			}
		})
	}
}

func TestServiceCreateSuite(t *testing.T) {
	wsID := uuid.New()

	tests := []struct {
		name    string
		input   CreateSuiteInput
		wantErr error
	}{
		{
			name:    "valid suite",
			input:   CreateSuiteInput{WorkspaceID: wsID, Name: "Regression Suite", Description: "All regression tests"},
			wantErr: nil,
		},
		{
			name:    "empty name",
			input:   CreateSuiteInput{WorkspaceID: wsID, Name: ""},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name:    "missing workspace",
			input:   CreateSuiteInput{WorkspaceID: uuid.Nil, Name: "Suite"},
			wantErr: sharederrors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(newFakeRepository())
			suite, err := service.CreateSuite(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if suite.ID == uuid.Nil {
				t.Error("expected suite ID to be set")
			}
		})
	}
}

func TestServiceCreateCase(t *testing.T) {
	wsID := uuid.New()
	projID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		input   CreateCaseInput
		wantErr error
	}{
		{
			name: "valid test case",
			input: CreateCaseInput{
				WorkspaceID: wsID,
				ProjectID:   projID,
				Title:       "Login with valid credentials",
				CreatedBy:   userID,
				Steps: []TestStep{
					{Order: 1, Action: "Navigate to login page", Expected: "Login form displayed"},
					{Order: 2, Action: "Enter valid credentials", Expected: "User logged in"},
				},
			},
			wantErr: nil,
		},
		{
			name: "empty title",
			input: CreateCaseInput{
				WorkspaceID: wsID,
				ProjectID:   projID,
				Title:       "",
				CreatedBy:   userID,
			},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name: "missing workspace",
			input: CreateCaseInput{
				WorkspaceID: uuid.Nil,
				ProjectID:   projID,
				Title:       "Test",
				CreatedBy:   userID,
			},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name: "missing project",
			input: CreateCaseInput{
				WorkspaceID: wsID,
				ProjectID:   uuid.Nil,
				Title:       "Test",
				CreatedBy:   userID,
			},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name: "missing created by",
			input: CreateCaseInput{
				WorkspaceID: wsID,
				ProjectID:   projID,
				Title:       "Test",
				CreatedBy:   uuid.Nil,
			},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name: "invalid status",
			input: CreateCaseInput{
				WorkspaceID: wsID,
				ProjectID:   projID,
				Title:       "Test",
				Status:      TestCaseStatus("invalid"),
				CreatedBy:   userID,
			},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name: "invalid priority",
			input: CreateCaseInput{
				WorkspaceID: wsID,
				ProjectID:   projID,
				Title:       "Test",
				Priority:    TestCasePriority("invalid"),
				CreatedBy:   userID,
			},
			wantErr: sharederrors.ErrInvalidInput,
		},
		{
			name: "default status and priority",
			input: CreateCaseInput{
				WorkspaceID: wsID,
				ProjectID:   projID,
				Title:       "Test with defaults",
				CreatedBy:   userID,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(newFakeRepository())
			tc, err := service.CreateCase(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.ID == uuid.Nil {
				t.Error("expected test case ID to be set")
			}
			if tc.Version != 1 {
				t.Errorf("expected version 1, got %d", tc.Version)
			}
			if tc.Status != TestCaseStatusDraft && tt.input.Status == "" {
				t.Errorf("expected default status draft, got %s", tc.Status)
			}
			if tc.Priority != TestCasePriorityMedium && tt.input.Priority == "" {
				t.Errorf("expected default priority medium, got %s", tc.Priority)
			}
		})
	}
}

func TestServiceUpdateCaseCreatesVersion(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(repo)
	wsID := uuid.New()
	projID := uuid.New()
	userID := uuid.New()

	tc, err := service.CreateCase(context.Background(), CreateCaseInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Title:       "Original Title",
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := service.UpdateCase(context.Background(), tc.ID, UpdateCaseInput{
		Title:     "Updated Title",
		Status:    TestCaseStatusActive,
		ChangedBy: userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", updated.Title)
	}
	if updated.Version != 2 {
		t.Errorf("expected version 2, got %d", updated.Version)
	}
	if updated.Status != TestCaseStatusActive {
		t.Errorf("expected status active, got %s", updated.Status)
	}

	versions, err := service.ListVersions(context.Background(), tc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].Title != "Original Title" {
		t.Errorf("expected version title 'Original Title', got %s", versions[0].Title)
	}
	if versions[0].Version != 1 {
		t.Errorf("expected version 1, got %d", versions[0].Version)
	}
}

func TestServiceGetAndDeleteCase(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(repo)
	wsID := uuid.New()
	projID := uuid.New()
	userID := uuid.New()

	created, err := service.CreateCase(context.Background(), CreateCaseInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Title:       "Test Case",
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := service.GetCase(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected test case %s, got %s", created.ID, got.ID)
	}

	if _, err := service.GetCase(context.Background(), uuid.New()); err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	if err := service.DeleteCase(context.Background(), created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := service.DeleteCase(context.Background(), created.ID); err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestServiceSearchCasesEmptyQuery(t *testing.T) {
	service := NewService(newFakeRepository())
	cases, _, err := service.SearchCases(context.Background(), uuid.New(), "", "", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cases != nil {
		t.Errorf("expected nil for empty query, got %v", cases)
	}
}

func TestServiceUpdateFolderNotFound(t *testing.T) {
	service := NewService(newFakeRepository())
	_, err := service.UpdateFolder(context.Background(), uuid.New(), UpdateFolderInput{Name: "Updated"})
	if err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceUpdateSuiteNotFound(t *testing.T) {
	service := NewService(newFakeRepository())
	_, err := service.UpdateSuite(context.Background(), uuid.New(), UpdateSuiteInput{Name: "Updated"})
	if err != sharederrors.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
