package testmanagement

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateFolderInput struct {
	WorkspaceID uuid.UUID
	ParentID    *uuid.UUID
	Name        string
}

func (s *Service) CreateFolder(ctx context.Context, input CreateFolderInput) (*TestFolder, error) {
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.WorkspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	folder := &TestFolder{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		ParentID:    input.ParentID,
		Name:        strings.TrimSpace(input.Name),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := s.repo.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *Service) GetFolder(ctx context.Context, id uuid.UUID) (*TestFolder, error) {
	return s.repo.GetFolderByID(ctx, id)
}

func (s *Service) ListFolders(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID, cursor string, limit int) ([]TestFolder, error) {
	return s.repo.ListFolders(ctx, workspaceID, parentID, cursor, limit)
}

type UpdateFolderInput struct {
	Name string
}

func (s *Service) UpdateFolder(ctx context.Context, id uuid.UUID, input UpdateFolderInput) (*TestFolder, error) {
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	folder, err := s.repo.GetFolderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	folder.Name = strings.TrimSpace(input.Name)
	folder.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateFolder(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *Service) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteFolder(ctx, id)
}

type CreateSuiteInput struct {
	WorkspaceID uuid.UUID
	FolderID    *uuid.UUID
	Name        string
	Description string
}

func (s *Service) CreateSuite(ctx context.Context, input CreateSuiteInput) (*TestSuite, error) {
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.WorkspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	suite := &TestSuite{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		FolderID:    input.FolderID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := s.repo.CreateSuite(ctx, suite); err != nil {
		return nil, err
	}
	return suite, nil
}

func (s *Service) GetSuite(ctx context.Context, id uuid.UUID) (*TestSuite, error) {
	return s.repo.GetSuiteByID(ctx, id)
}

func (s *Service) ListSuites(ctx context.Context, workspaceID uuid.UUID, folderID *uuid.UUID, cursor string, limit int) ([]TestSuite, error) {
	return s.repo.ListSuites(ctx, workspaceID, folderID, cursor, limit)
}

type UpdateSuiteInput struct {
	Name        string
	Description string
}

func (s *Service) UpdateSuite(ctx context.Context, id uuid.UUID, input UpdateSuiteInput) (*TestSuite, error) {
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	suite, err := s.repo.GetSuiteByID(ctx, id)
	if err != nil {
		return nil, err
	}
	suite.Name = strings.TrimSpace(input.Name)
	suite.Description = strings.TrimSpace(input.Description)
	suite.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateSuite(ctx, suite); err != nil {
		return nil, err
	}
	return suite, nil
}

func (s *Service) DeleteSuite(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteSuite(ctx, id)
}

type CreateCaseInput struct {
	WorkspaceID   uuid.UUID
	ProjectID     uuid.UUID
	SuiteID       *uuid.UUID
	Title         string
	Description   string
	Preconditions string
	Steps         []TestStep
	Status        TestCaseStatus
	Priority      TestCasePriority
	Tags          []string
	CreatedBy     uuid.UUID
}

func (s *Service) CreateCase(ctx context.Context, input CreateCaseInput) (*TestCase, error) {
	if !validation.IsValidName(input.Title) {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.WorkspaceID == uuid.Nil || input.ProjectID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.CreatedBy == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if err := validateSteps(input.Steps); err != nil {
		return nil, err
	}

	status := input.Status
	if status == "" {
		status = TestCaseStatusDraft
	}
	if !isValidStatus(status) {
		return nil, sharederrors.ErrInvalidInput
	}

	priority := input.Priority
	if priority == "" {
		priority = TestCasePriorityMedium
	}
	if !isValidPriority(priority) {
		return nil, sharederrors.ErrInvalidInput
	}

	tc := &TestCase{
		ID:            uuid.New(),
		WorkspaceID:   input.WorkspaceID,
		ProjectID:     input.ProjectID,
		SuiteID:       input.SuiteID,
		Title:         strings.TrimSpace(input.Title),
		Description:   strings.TrimSpace(input.Description),
		Preconditions: strings.TrimSpace(input.Preconditions),
		Steps:         input.Steps,
		Status:        status,
		Priority:      priority,
		Tags:          input.Tags,
		Version:       1,
		CreatedBy:     input.CreatedBy,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := s.repo.CreateCase(ctx, tc); err != nil {
		return nil, err
	}
	return tc, nil
}

func (s *Service) GetCase(ctx context.Context, id uuid.UUID) (*TestCase, error) {
	return s.repo.GetCaseByID(ctx, id)
}

func (s *Service) ListCases(ctx context.Context, projectID uuid.UUID, suiteID *uuid.UUID, cursor string, limit int) ([]TestCase, error) {
	return s.repo.ListCases(ctx, projectID, suiteID, cursor, limit)
}

func (s *Service) SearchCases(ctx context.Context, workspaceID uuid.UUID, query string, cursor string, limit int) ([]TestCase, string, error) {
	if strings.TrimSpace(query) == "" {
		return nil, "", nil
	}
	return s.repo.SearchCases(ctx, workspaceID, query, cursor, limit)
}

type UpdateCaseInput struct {
	SuiteID       *uuid.UUID
	Title         string
	Description   string
	Preconditions string
	Steps         []TestStep
	Status        TestCaseStatus
	Priority      TestCasePriority
	Tags          []string
	ChangedBy     uuid.UUID
}

func (s *Service) UpdateCase(ctx context.Context, id uuid.UUID, input UpdateCaseInput) (*TestCase, error) {
	if !validation.IsValidName(input.Title) {
		return nil, sharederrors.ErrInvalidInput
	}

	if input.Status != "" && !isValidStatus(input.Status) {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.Priority != "" && !isValidPriority(input.Priority) {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.ChangedBy == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if err := validateSteps(input.Steps); err != nil {
		return nil, err
	}

	var result *TestCase

	err := s.repo.RunInTx(ctx, func(txRepo Repository) error {
		tc, err := txRepo.GetCaseByID(ctx, id)
		if err != nil {
			return err
		}

		oldVersion := &TestCaseVersion{
			ID:            uuid.New(),
			TestCaseID:    tc.ID,
			Version:       tc.Version,
			Title:         tc.Title,
			Description:   tc.Description,
			Preconditions: tc.Preconditions,
			Steps:         tc.Steps,
			ChangedBy:     input.ChangedBy,
			CreatedAt:     time.Now().UTC(),
		}
		if err := txRepo.CreateVersion(ctx, oldVersion); err != nil {
			return err
		}

		tc.SuiteID = input.SuiteID
		tc.Title = strings.TrimSpace(input.Title)
		tc.Description = strings.TrimSpace(input.Description)
		tc.Preconditions = strings.TrimSpace(input.Preconditions)
		tc.Steps = input.Steps
		if input.Status != "" {
			tc.Status = input.Status
		}
		if input.Priority != "" {
			tc.Priority = input.Priority
		}
		tc.Tags = input.Tags
		tc.Version++
		tc.UpdatedAt = time.Now().UTC()

		if err := txRepo.UpdateCase(ctx, tc); err != nil {
			return err
		}

		result = tc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) DeleteCase(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteCase(ctx, id)
}

func (s *Service) ListVersions(ctx context.Context, caseID uuid.UUID, cursor string, limit int) ([]TestCaseVersion, error) {
	return s.repo.ListVersions(ctx, caseID, cursor, limit)
}

func isValidStatus(s TestCaseStatus) bool {
	switch s {
	case TestCaseStatusDraft, TestCaseStatusActive, TestCaseStatusDeprecated:
		return true
	}
	return false
}

func isValidPriority(p TestCasePriority) bool {
	switch p {
	case TestCasePriorityLow, TestCasePriorityMedium, TestCasePriorityHigh, TestCasePriorityCritical:
		return true
	}
	return false
}

func validateSteps(steps []TestStep) error {
	for i, s := range steps {
		if s.Order <= 0 {
			return sharederrors.ErrInvalidInput
		}
		if strings.TrimSpace(s.Action) == "" {
			return sharederrors.ErrInvalidInput
		}
		if strings.TrimSpace(s.Expected) == "" {
			return sharederrors.ErrInvalidInput
		}
		// Consecutive order indices are required for deterministic responses.
		if i > 0 && s.Order != steps[i-1].Order+1 {
			return sharederrors.ErrInvalidInput
		}
	}
	return nil
}
