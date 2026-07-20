package automationhub

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/defects"
	"github.com/testra/testra/apps/api/internal/results"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
	"github.com/testra/testra/apps/api/internal/shared/validation"
	"github.com/testra/testra/apps/api/internal/testmanagement"
)

type Service struct {
	repo         Repository
	resultsRepo  results.Repository
	defectsRepo  defects.Repository
	testMgmtRepo testmanagement.Repository
	storage      *ArtifactStorage
}

func NewService(repo Repository, resultsRepo results.Repository, defectsRepo defects.Repository, testMgmtRepo testmanagement.Repository, storage *ArtifactStorage) *Service {
	return &Service{
		repo:         repo,
		resultsRepo:  resultsRepo,
		defectsRepo:  defectsRepo,
		testMgmtRepo: testMgmtRepo,
		storage:      storage,
	}
}

// Ingest is the legacy CI endpoint that creates a TestRun directly.
type IngestInput struct {
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID
	SuiteID     *uuid.UUID
	Name        string
	Format      IngestionFormat
	Body        []byte
	CreatedBy   uuid.UUID
}

func (s *Service) Ingest(ctx context.Context, input IngestInput) (*IngestResult, error) {
	if input.WorkspaceID == uuid.Nil || input.ProjectID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}
	if !IsValidFormat(string(input.Format)) {
		return nil, sharederrors.ErrInvalidInput
	}
	if len(input.Body) == 0 {
		return nil, sharederrors.ErrInvalidInput
	}

	report, err := ParseReport(input.Format, input.Body)
	if err != nil {
		return nil, fmt.Errorf("parse report: %w", err)
	}

	run, err := s.createResultsRun(ctx, input.WorkspaceID, input.ProjectID, input.SuiteID, input.Name, results.RunSourceCI, input.CreatedBy, report)
	if err != nil {
		return nil, err
	}

	return &IngestResult{
		RunID:      run.ID,
		Total:      run.Total,
		Passed:     run.Passed,
		Failed:     run.Failed,
		Skipped:    run.Skipped,
		DurationMs: run.DurationMs,
	}, nil
}

// ----------------- Automation Projects -----------------

type CreateProjectInput struct {
	WorkspaceID   uuid.UUID
	ProjectID     *uuid.UUID
	Name          string
	Framework     string
	RepositoryURL string
	Branch        string
	Command       string
	CreatedBy     uuid.UUID
}

func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (*AutomationProject, error) {
	if input.WorkspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	now := nowUTC()
	p := &AutomationProject{
		ID:            uuid.New(),
		WorkspaceID:   input.WorkspaceID,
		ProjectID:     input.ProjectID,
		Name:          strings.TrimSpace(input.Name),
		Framework:     strings.TrimSpace(input.Framework),
		RepositoryURL: strings.TrimSpace(input.RepositoryURL),
		Branch:        strings.TrimSpace(input.Branch),
		Command:       strings.TrimSpace(input.Command),
		CreatedBy:     input.CreatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateProject(ctx, p); err != nil {
		return nil, err
	}

	eventbus.Default().Publish(ctx, eventbus.Event{
		Type:     "automation.project_created",
		TenantID: p.WorkspaceID.String(),
		Payload: map[string]interface{}{
			"automation_project_id": p.ID.String(),
			"workspace_id":          p.WorkspaceID.String(),
			"name":                  p.Name,
		},
	})

	return p, nil
}

func (s *Service) GetProject(ctx context.Context, id uuid.UUID) (*AutomationProject, error) {
	return s.repo.GetProject(ctx, id)
}

func (s *Service) ListProjects(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]AutomationProject, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 20
	}
	return s.repo.ListProjects(ctx, workspaceID, cursor, limit)
}

type UpdateProjectInput struct {
	Name          string
	Framework     string
	RepositoryURL string
	Branch        string
	Command       string
}

func (s *Service) UpdateProject(ctx context.Context, id uuid.UUID, input UpdateProjectInput) (*AutomationProject, error) {
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}
	p, err := s.repo.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Name = strings.TrimSpace(input.Name)
	p.Framework = strings.TrimSpace(input.Framework)
	p.RepositoryURL = strings.TrimSpace(input.RepositoryURL)
	p.Branch = strings.TrimSpace(input.Branch)
	p.Command = strings.TrimSpace(input.Command)
	p.UpdatedAt = nowUTC()
	if err := s.repo.UpdateProject(ctx, p); err != nil {
		return nil, err
	}

	eventbus.Default().Publish(ctx, eventbus.Event{
		Type:     "automation.project_updated",
		TenantID: p.WorkspaceID.String(),
		Payload: map[string]interface{}{
			"automation_project_id": p.ID.String(),
			"workspace_id":          p.WorkspaceID.String(),
			"name":                  p.Name,
		},
	})

	return p, nil
}

func (s *Service) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteProject(ctx, id)
}

// ----------------- Executions & Import -----------------

type ImportExecutionInput struct {
	ProjectID         uuid.UUID
	Name              string
	Format            IngestionFormat
	Report            []byte
	CreatedBy         uuid.UUID
	AutoCreateDefects bool
	MapTestCases      bool
}

func (s *Service) ImportExecution(ctx context.Context, input ImportExecutionInput) (*IngestResult, error) {
	if input.ProjectID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}
	if !IsValidFormat(string(input.Format)) {
		return nil, sharederrors.ErrInvalidInput
	}
	if len(input.Report) == 0 {
		return nil, sharederrors.ErrInvalidInput
	}

	project, err := s.repo.GetProject(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	if project.ProjectID == nil {
		return nil, fmt.Errorf("automation project must be linked to a project")
	}

	report, err := ParseReport(input.Format, input.Report)
	if err != nil {
		return nil, fmt.Errorf("parse report: %w", err)
	}

	execution, run, err := s.createExecutionFromReport(ctx, project, input, report)
	if err != nil {
		return nil, err
	}

	if input.AutoCreateDefects {
		if err := s.createDefectsForFailedItems(ctx, project, run); err != nil {
			return nil, err
		}
	}

	return &IngestResult{
		ExecutionID: execution.ID,
		RunID:       run.ID,
		Total:       run.Total,
		Passed:      run.Passed,
		Failed:      run.Failed,
		Skipped:     run.Skipped,
		DurationMs:  run.DurationMs,
	}, nil
}

func (s *Service) GetExecution(ctx context.Context, id uuid.UUID) (*AutomationExecution, error) {
	return s.repo.GetExecution(ctx, id)
}

func (s *Service) ListExecutions(ctx context.Context, projectID uuid.UUID, cursor string, limit int) ([]AutomationExecution, error) {
	if projectID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 20
	}
	return s.repo.ListExecutions(ctx, projectID, cursor, limit)
}

func (s *Service) ListExecutionsByWorkspace(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]AutomationExecution, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 20
	}
	return s.repo.ListExecutionsByWorkspace(ctx, workspaceID, cursor, limit)
}

func (s *Service) RerunExecution(ctx context.Context, id uuid.UUID, triggeredBy uuid.UUID) (*AutomationExecution, error) {
	orig, err := s.repo.GetExecution(ctx, id)
	if err != nil {
		return nil, err
	}
	project, err := s.repo.GetProject(ctx, orig.ProjectID)
	if err != nil {
		return nil, err
	}

	now := nowUTC()
	retry := &AutomationExecution{
		ID:           uuid.New(),
		ProjectID:    orig.ProjectID,
		WorkspaceID:  project.WorkspaceID,
		Name:         fmt.Sprintf("%s (retry)", orig.Name),
		Status:       "pending",
		ReportFormat: orig.ReportFormat,
		RetryOf:      &orig.ID,
		CreatedBy:    triggeredBy,
		TriggeredBy:  triggeredBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateExecution(ctx, retry); err != nil {
		return nil, err
	}
	return retry, nil
}

func (s *Service) DeleteExecution(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteExecution(ctx, id)
}

// ----------------- Artifacts -----------------

type UploadArtifactInput struct {
	ExecutionID   uuid.UUID
	WorkspaceID   uuid.UUID
	TestRunItemID *uuid.UUID
	Kind          ArtifactKind
	Name          string
	MimeType      string
	Data          []byte
}

func (s *Service) UploadArtifact(ctx context.Context, input UploadArtifactInput) (*AutomationArtifact, error) {
	if input.ExecutionID == uuid.Nil || input.WorkspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if !IsValidArtifactKind(string(input.Kind)) {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.Name == "" || len(input.Data) == 0 {
		return nil, sharederrors.ErrInvalidInput
	}

	path, size, err := s.storage.SaveArtifact(input.ExecutionID, input.Kind, input.Name, input.Data)
	if err != nil {
		return nil, err
	}

	a := &AutomationArtifact{
		ID:            uuid.New(),
		ExecutionID:   input.ExecutionID,
		WorkspaceID:   input.WorkspaceID,
		TestRunItemID: input.TestRunItemID,
		Kind:          input.Kind,
		Name:          input.Name,
		FilePath:      s.storage.RelativePath(path),
		MimeType:      input.MimeType,
		FileSize:      size,
		Metadata:      map[string]interface{}{},
		CreatedAt:     nowUTC(),
	}
	if err := s.repo.CreateArtifact(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) GetArtifact(ctx context.Context, id uuid.UUID) (*AutomationArtifact, []byte, error) {
	a, err := s.repo.GetArtifact(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	data, err := s.storage.ReadArtifact(s.storage.FullPath(a.FilePath))
	if err != nil {
		return nil, nil, err
	}
	return a, data, nil
}

func (s *Service) ListArtifacts(ctx context.Context, executionID uuid.UUID, cursor string, limit int) ([]AutomationArtifact, error) {
	if executionID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 20
	}
	return s.repo.ListArtifacts(ctx, executionID, cursor, limit)
}

func (s *Service) DeleteArtifact(ctx context.Context, id uuid.UUID) error {
	a, err := s.repo.GetArtifact(ctx, id)
	if err != nil {
		return err
	}
	if err := s.storage.DeleteArtifact(s.storage.FullPath(a.FilePath)); err != nil {
		// Continue and remove the database record even if the file is missing.
	}
	return s.repo.DeleteArtifact(ctx, id)
}

// ----------------- Logs -----------------

type AddLogInput struct {
	ExecutionID uuid.UUID
	WorkspaceID uuid.UUID
	Level       string
	Message     string
	LoggedAt    time.Time
}

func (s *Service) AddLog(ctx context.Context, input AddLogInput) (*AutomationLog, error) {
	if input.ExecutionID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.Message == "" {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.Level == "" {
		input.Level = "info"
	}
	if input.LoggedAt.IsZero() {
		input.LoggedAt = nowUTC()
	}
	l := &AutomationLog{
		ID:          uuid.New(),
		ExecutionID: input.ExecutionID,
		WorkspaceID: input.WorkspaceID,
		Level:       input.Level,
		Message:     input.Message,
		LoggedAt:    input.LoggedAt,
		CreatedAt:   nowUTC(),
	}
	if err := s.repo.CreateLog(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Service) ListLogs(ctx context.Context, executionID uuid.UUID, cursor string, limit int) ([]AutomationLog, error) {
	if executionID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 100
	}
	return s.repo.ListLogs(ctx, executionID, cursor, limit)
}

// ----------------- internal helpers -----------------

func (s *Service) createExecutionFromReport(ctx context.Context, project *AutomationProject, input ImportExecutionInput, report *ParsedReport) (*AutomationExecution, *results.TestRun, error) {
	now := nowUTC()
	execution := &AutomationExecution{
		ID:           uuid.New(),
		ProjectID:    project.ID,
		WorkspaceID:  project.WorkspaceID,
		Name:         strings.TrimSpace(input.Name),
		Status:       string(results.RunStatusRunning),
		ReportFormat: string(input.Format),
		CreatedBy:    input.CreatedBy,
		TriggeredBy:  input.CreatedBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	var run *results.TestRun
	err := s.repo.RunInTx(ctx, func(txRepo Repository) error {
		if err := txRepo.CreateExecution(ctx, execution); err != nil {
			return err
		}

		path, size, err := s.storage.SaveArtifact(execution.ID, ArtifactKindReport, fmt.Sprintf("report.%s", input.Format), input.Report)
		if err != nil {
			return err
		}
		execution.ReportPath = s.storage.RelativePath(path)

		reportArtifact := &AutomationArtifact{
			ID:          uuid.New(),
			ExecutionID: execution.ID,
			WorkspaceID: project.WorkspaceID,
			Kind:        ArtifactKindReport,
			Name:        fmt.Sprintf("report.%s", input.Format),
			FilePath:    execution.ReportPath,
			MimeType:    mimeTypeFor(input.Format),
			FileSize:    size,
			Metadata:    map[string]interface{}{"format": string(input.Format)},
			CreatedAt:   now,
		}
		if err := txRepo.CreateArtifact(ctx, reportArtifact); err != nil {
			return err
		}

		run, err = s.createResultsRunTx(ctx, txRepo, project, input, report)
		if err != nil {
			return err
		}
		execution.TestRunID = &run.ID
		execution.Status = string(run.Status)
		execution.Total = run.Total
		execution.Passed = run.Passed
		execution.Failed = run.Failed
		execution.Skipped = run.Skipped
		execution.Blocked = run.Blocked
		execution.DurationMs = run.DurationMs

		if err := txRepo.UpdateExecution(ctx, execution); err != nil {
			return err
		}

		// Persist parsed logs as timeline entries.
		for _, suite := range report.Suites {
			for _, c := range suite.Cases {
				for _, line := range c.Logs {
					l := &AutomationLog{
						ID:          uuid.New(),
						ExecutionID: execution.ID,
						WorkspaceID: project.WorkspaceID,
						Level:       "info",
						Message:     line,
						LoggedAt:    now,
						CreatedAt:   now,
					}
					if strings.Contains(line, "FAIL") || strings.Contains(line, "ERROR") {
						l.Level = "error"
					}
					if err := txRepo.CreateLog(ctx, l); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if execution != nil {
		eventbus.Default().Publish(ctx, eventbus.Event{
			Type:     "automation.execution_completed",
			TenantID: project.WorkspaceID.String(),
			Payload: map[string]interface{}{
				"execution_id": execution.ID.String(),
				"run_id":       run.ID.String(),
				"project_id":   project.ID.String(),
				"workspace_id": project.WorkspaceID.String(),
				"status":       string(execution.Status),
			},
		})
	}

	return execution, run, nil
}

func (s *Service) createResultsRun(ctx context.Context, workspaceID, projectID uuid.UUID, suiteID *uuid.UUID, name string, source results.RunSource, createdBy uuid.UUID, report *ParsedReport) (*results.TestRun, error) {
	now := nowUTC()
	runID := uuid.New()
	run := &results.TestRun{
		ID:          runID,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		SuiteID:     suiteID,
		Name:        name,
		Status:      results.RunStatusRunning,
		Source:      source,
		CreatedBy:   createdBy,
		StartedAt:   &now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    map[string]interface{}{},
	}

	if err := s.resultsRepo.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	if err := s.createRunItems(ctx, run, report, nil); err != nil {
		return nil, err
	}

	run.Total, run.Passed, run.Failed, run.Skipped, run.Blocked, run.DurationMs = summarizeReport(report)
	run.Status = runStatusFromCounts(run.Total, run.Failed)
	run.CompletedAt = &now
	run.UpdatedAt = now

	if err := s.resultsRepo.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *Service) createResultsRunTx(ctx context.Context, txRepo Repository, project *AutomationProject, input ImportExecutionInput, report *ParsedReport) (*results.TestRun, error) {
	now := nowUTC()
	runID := uuid.New()
	run := &results.TestRun{
		ID:          runID,
		WorkspaceID: project.WorkspaceID,
		ProjectID:   *project.ProjectID,
		SuiteID:     nil,
		Name:        input.Name,
		Status:      results.RunStatusRunning,
		Source:      results.RunSourceCI,
		CreatedBy:   input.CreatedBy,
		StartedAt:   &now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    map[string]interface{}{"automation_project_id": project.ID.String(), "format": string(input.Format)},
	}

	if err := s.resultsRepo.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	tcMap := map[string]uuid.UUID{}
	if input.MapTestCases && project.ProjectID != nil {
		tcMap = s.buildTestCaseMap(ctx, *project.ProjectID)
	}

	if err := s.createRunItems(ctx, run, report, tcMap); err != nil {
		return nil, err
	}

	run.Total, run.Passed, run.Failed, run.Skipped, run.Blocked, run.DurationMs = summarizeReport(report)
	run.Status = runStatusFromCounts(run.Total, run.Failed)
	run.CompletedAt = &now
	run.UpdatedAt = now

	if err := s.resultsRepo.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *Service) createRunItems(ctx context.Context, run *results.TestRun, report *ParsedReport, tcMap map[string]uuid.UUID) error {
	now := nowUTC()
	sortOrder := 0
	for _, suite := range report.Suites {
		for _, c := range suite.Cases {
			item := parsedCaseToRunItem(run.ID, c, sortOrder, tcMap, now)
			if err := s.resultsRepo.CreateItem(ctx, item); err != nil {
				return err
			}
			sortOrder++
		}
	}
	return nil
}

func parsedCaseToRunItem(runID uuid.UUID, c ParsedCase, sortOrder int, tcMap map[string]uuid.UUID, now time.Time) *results.TestRunItem {
	status := toRunItemStatus(c.Status)
	tcID := tcMap[strings.ToLower(c.Name)]
	var tcIDPtr *uuid.UUID
	if tcID != uuid.Nil {
		tcIDPtr = &tcID
	}
	artifacts := []string{}
	for _, s := range c.Screenshots {
		artifacts = append(artifacts, s)
	}
	return &results.TestRunItem{
		ID:           uuid.New(),
		RunID:        runID,
		TestCaseID:   tcIDPtr,
		Title:        c.Name,
		Status:       status,
		DurationMs:   c.DurationMs,
		ErrorMessage: c.ErrorMessage,
		StackTrace:   c.StackTrace,
		Artifacts:    artifacts,
		SortOrder:    sortOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (s *Service) buildTestCaseMap(ctx context.Context, projectID uuid.UUID) map[string]uuid.UUID {
	m := map[string]uuid.UUID{}
	cases, err := s.testMgmtRepo.ListCases(ctx, projectID, nil, "", 1000)
	if err != nil {
		return m
	}
	for _, tc := range cases {
		m[strings.ToLower(tc.Title)] = tc.ID
	}
	return m
}

func (s *Service) createDefectsForFailedItems(ctx context.Context, project *AutomationProject, run *results.TestRun) error {
	if project.ProjectID == nil {
		return nil
	}
	items, err := s.resultsRepo.ListItems(ctx, run.ID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Status != results.RunItemStatusFailed {
			continue
		}
		d := &defects.Defect{
			ID:            uuid.New(),
			WorkspaceID:   run.WorkspaceID,
			ProjectID:     *project.ProjectID,
			TestRunItemID: &item.ID,
			Title:         fmt.Sprintf("Automation failure: %s", item.Title),
			Description:   item.ErrorMessage,
			Severity:      defects.DefectSeverityHigh,
			Priority:      defects.DefectPriorityMedium,
			Status:        defects.DefectStatusOpen,
			CreatedBy:     run.CreatedBy,
			CreatedAt:     nowUTC(),
			UpdatedAt:     nowUTC(),
		}
		if err := s.defectsRepo.Create(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

func toRunItemStatus(status string) results.RunItemStatus {
	switch strings.ToLower(status) {
	case "passed", "pass":
		return results.RunItemStatusPassed
	case "failed", "fail":
		return results.RunItemStatusFailed
	case "skipped", "skip", "pending":
		return results.RunItemStatusSkipped
	case "blocked":
		return results.RunItemStatusBlocked
	default:
		return results.RunItemStatusPassed
	}
}

func toRunStatus(status string) results.RunStatus {
	switch strings.ToLower(status) {
	case "passed":
		return results.RunStatusPassed
	case "failed":
		return results.RunStatusFailed
	case "skipped":
		return results.RunStatusSkipped
	case "cancelled":
		return results.RunStatusCancelled
	default:
		return results.RunStatusPending
	}
}

func runStatusFromCounts(total, failed int) results.RunStatus {
	if total == 0 {
		return results.RunStatusPassed
	}
	if failed > 0 {
		return results.RunStatusFailed
	}
	return results.RunStatusPassed
}

func summarizeReport(report *ParsedReport) (int, int, int, int, int, int64) {
	return report.Total, report.Passed, report.Failed, report.Skipped, report.Blocked, report.DurationMs
}

func mimeTypeFor(format IngestionFormat) string {
	switch format {
	case FormatJUnit, FormatPytestJUnit, FormatRobot:
		return "application/xml"
	case FormatPlaywright, FormatCypress, FormatNewman:
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
