package automationhub

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/defects"
	"github.com/testra/testra/apps/api/internal/results"
	"github.com/testra/testra/apps/api/internal/testmanagement"
)

// ---------- Fakes ----------

type fakeAutomationRepo struct {
	projects   map[uuid.UUID]*AutomationProject
	executions map[uuid.UUID]*AutomationExecution
	artifacts  map[uuid.UUID]*AutomationArtifact
	logs       map[uuid.UUID]*AutomationLog
}

func newFakeAutomationRepo() *fakeAutomationRepo {
	return &fakeAutomationRepo{
		projects:   make(map[uuid.UUID]*AutomationProject),
		executions: make(map[uuid.UUID]*AutomationExecution),
		artifacts:  make(map[uuid.UUID]*AutomationArtifact),
		logs:       make(map[uuid.UUID]*AutomationLog),
	}
}

func (f *fakeAutomationRepo) CreateProject(_ context.Context, p *AutomationProject) error {
	f.projects[p.ID] = p
	return nil
}

func (f *fakeAutomationRepo) GetProject(_ context.Context, id uuid.UUID) (*AutomationProject, error) {
	p, ok := f.projects[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (f *fakeAutomationRepo) ListProjects(_ context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]AutomationProject, error) {
	out := make([]AutomationProject, 0)
	for _, p := range f.projects {
		if p.WorkspaceID == workspaceID {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (f *fakeAutomationRepo) UpdateProject(_ context.Context, p *AutomationProject) error {
	f.projects[p.ID] = p
	return nil
}

func (f *fakeAutomationRepo) DeleteProject(_ context.Context, id uuid.UUID) error {
	delete(f.projects, id)
	return nil
}

func (f *fakeAutomationRepo) CreateExecution(_ context.Context, e *AutomationExecution) error {
	f.executions[e.ID] = e
	return nil
}

func (f *fakeAutomationRepo) GetExecution(_ context.Context, id uuid.UUID) (*AutomationExecution, error) {
	e, ok := f.executions[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return e, nil
}

func (f *fakeAutomationRepo) ListExecutions(_ context.Context, projectID uuid.UUID, cursor string, limit int) ([]AutomationExecution, error) {
	out := make([]AutomationExecution, 0)
	for _, e := range f.executions {
		if e.ProjectID == projectID {
			out = append(out, *e)
		}
	}
	return out, nil
}

func (f *fakeAutomationRepo) ListExecutionsByWorkspace(_ context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]AutomationExecution, error) {
	out := make([]AutomationExecution, 0)
	for _, e := range f.executions {
		if e.WorkspaceID == workspaceID {
			out = append(out, *e)
		}
	}
	return out, nil
}

func (f *fakeAutomationRepo) UpdateExecution(_ context.Context, e *AutomationExecution) error {
	f.executions[e.ID] = e
	return nil
}

func (f *fakeAutomationRepo) DeleteExecution(_ context.Context, id uuid.UUID) error {
	delete(f.executions, id)
	return nil
}

func (f *fakeAutomationRepo) CreateArtifact(_ context.Context, a *AutomationArtifact) error {
	f.artifacts[a.ID] = a
	return nil
}

func (f *fakeAutomationRepo) GetArtifact(_ context.Context, id uuid.UUID) (*AutomationArtifact, error) {
	a, ok := f.artifacts[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return a, nil
}

func (f *fakeAutomationRepo) ListArtifacts(_ context.Context, executionID uuid.UUID, cursor string, limit int) ([]AutomationArtifact, error) {
	out := make([]AutomationArtifact, 0)
	for _, a := range f.artifacts {
		if a.ExecutionID == executionID {
			out = append(out, *a)
		}
	}
	return out, nil
}

func (f *fakeAutomationRepo) DeleteArtifact(_ context.Context, id uuid.UUID) error {
	delete(f.artifacts, id)
	return nil
}

func (f *fakeAutomationRepo) CreateLog(_ context.Context, l *AutomationLog) error {
	f.logs[l.ID] = l
	return nil
}

func (f *fakeAutomationRepo) ListLogs(_ context.Context, executionID uuid.UUID, cursor string, limit int) ([]AutomationLog, error) {
	out := make([]AutomationLog, 0)
	for _, l := range f.logs {
		if l.ExecutionID == executionID {
			out = append(out, *l)
		}
	}
	return out, nil
}

func (f *fakeAutomationRepo) RunInTx(_ context.Context, fn func(Repository) error) error {
	return fn(f)
}

type fakeResultsRepo struct {
	runs  map[uuid.UUID]*results.TestRun
	items map[uuid.UUID]*results.TestRunItem
}

func newFakeResultsRepo() *fakeResultsRepo {
	return &fakeResultsRepo{
		runs:  make(map[uuid.UUID]*results.TestRun),
		items: make(map[uuid.UUID]*results.TestRunItem),
	}
}

func (f *fakeResultsRepo) CreateRun(_ context.Context, run *results.TestRun) error {
	f.runs[run.ID] = run
	return nil
}

func (f *fakeResultsRepo) GetRunByID(_ context.Context, id uuid.UUID) (*results.TestRun, error) {
	r, ok := f.runs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (f *fakeResultsRepo) ListRuns(_ context.Context, projectID uuid.UUID, cursor string, limit int) ([]results.TestRun, error) {
	out := make([]results.TestRun, 0)
	for _, r := range f.runs {
		if r.ProjectID == projectID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeResultsRepo) UpdateRun(_ context.Context, run *results.TestRun) error {
	f.runs[run.ID] = run
	return nil
}

func (f *fakeResultsRepo) DeleteRun(_ context.Context, id uuid.UUID) error {
	delete(f.runs, id)
	return nil
}

func (f *fakeResultsRepo) CreateItem(_ context.Context, item *results.TestRunItem) error {
	f.items[item.ID] = item
	return nil
}

func (f *fakeResultsRepo) GetItemByID(_ context.Context, id uuid.UUID) (*results.TestRunItem, error) {
	i, ok := f.items[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return i, nil
}

func (f *fakeResultsRepo) ListItems(_ context.Context, runID uuid.UUID) ([]results.TestRunItem, error) {
	out := make([]results.TestRunItem, 0)
	for _, i := range f.items {
		if i.RunID == runID {
			out = append(out, *i)
		}
	}
	return out, nil
}

func (f *fakeResultsRepo) UpdateItem(_ context.Context, item *results.TestRunItem) error {
	f.items[item.ID] = item
	return nil
}

func (f *fakeResultsRepo) DeleteItemsByRunID(_ context.Context, runID uuid.UUID) error {
	for id, i := range f.items {
		if i.RunID == runID {
			delete(f.items, id)
		}
	}
	return nil
}

func (f *fakeResultsRepo) RunInTx(_ context.Context, fn func(results.Repository) error) error {
	return fn(f)
}

func (f *fakeResultsRepo) CreateItemExecution(_ context.Context, item *results.TestRunItem) error {
	f.items[item.ID] = item
	return nil
}

func (f *fakeResultsRepo) ListItemsByRunPaged(_ context.Context, runID uuid.UUID, status, search, cursor string, limit int) ([]results.TestRunItem, error) {
	out := make([]results.TestRunItem, 0)
	for _, i := range f.items {
		if i.RunID == runID {
			out = append(out, *i)
		}
	}
	return out, nil
}

func (f *fakeResultsRepo) CreateItemHistory(_ context.Context, _ *results.RunItemHistory) error {
	return nil
}
func (f *fakeResultsRepo) ListItemHistory(_ context.Context, _ uuid.UUID) ([]results.RunItemHistory, error) {
	return nil, nil
}

func (f *fakeResultsRepo) CreateEvidence(_ context.Context, _ *results.EvidenceRef) error { return nil }
func (f *fakeResultsRepo) ListEvidenceByItem(_ context.Context, _ uuid.UUID) ([]results.EvidenceRef, error) {
	return nil, nil
}
func (f *fakeResultsRepo) DeleteEvidence(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeResultsRepo) CreateRunItemDefect(_ context.Context, _, _ uuid.UUID) error { return nil }
func (f *fakeResultsRepo) ListRunItemDefects(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (f *fakeResultsRepo) DeleteRunItemDefect(_ context.Context, _, _ uuid.UUID) error { return nil }

func (f *fakeResultsRepo) CreatePlan(_ context.Context, _ *results.TestPlan) error { return nil }
func (f *fakeResultsRepo) GetPlanByID(_ context.Context, _ uuid.UUID) (*results.TestPlan, error) {
	return nil, fmt.Errorf("not found")
}
func (f *fakeResultsRepo) ListPlans(_ context.Context, _ uuid.UUID, _ string, _ int) ([]results.TestPlan, error) {
	return nil, nil
}
func (f *fakeResultsRepo) UpdatePlan(_ context.Context, _ *results.TestPlan) error { return nil }
func (f *fakeResultsRepo) DeletePlan(_ context.Context, _ uuid.UUID) error         { return nil }
func (f *fakeResultsRepo) CreatePlanItem(_ context.Context, _ *results.TestPlanItem) error {
	return nil
}
func (f *fakeResultsRepo) ListPlanItems(_ context.Context, _ uuid.UUID) ([]results.TestPlanItem, error) {
	return nil, nil
}
func (f *fakeResultsRepo) DeletePlanItemsByPlanID(_ context.Context, _ uuid.UUID) error { return nil }

type fakeDefectsRepo struct {
	defects map[uuid.UUID]*defects.Defect
}

func newFakeDefectsRepo() *fakeDefectsRepo {
	return &fakeDefectsRepo{defects: make(map[uuid.UUID]*defects.Defect)}
}

func (f *fakeDefectsRepo) Create(_ context.Context, d *defects.Defect) error {
	f.defects[d.ID] = d
	return nil
}

func (f *fakeDefectsRepo) GetByID(_ context.Context, id uuid.UUID) (*defects.Defect, error) {
	d, ok := f.defects[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return d, nil
}

func (f *fakeDefectsRepo) ListByProject(_ context.Context, projectID uuid.UUID, cursor string, limit int) ([]defects.Defect, error) {
	return nil, nil
}

func (f *fakeDefectsRepo) Update(_ context.Context, d *defects.Defect) error {
	f.defects[d.ID] = d
	return nil
}

func (f *fakeDefectsRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(f.defects, id)
	return nil
}

type fakeTestMgmtRepo struct{}

func (f *fakeTestMgmtRepo) CreateFolder(_ context.Context, _ *testmanagement.TestFolder) error {
	return nil
}
func (f *fakeTestMgmtRepo) GetFolderByID(_ context.Context, _ uuid.UUID) (*testmanagement.TestFolder, error) {
	return nil, fmt.Errorf("not found")
}
func (f *fakeTestMgmtRepo) ListFolders(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ string, _ int) ([]testmanagement.TestFolder, error) {
	return nil, nil
}
func (f *fakeTestMgmtRepo) UpdateFolder(_ context.Context, _ *testmanagement.TestFolder) error {
	return nil
}
func (f *fakeTestMgmtRepo) DeleteFolder(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeTestMgmtRepo) CreateSuite(_ context.Context, _ *testmanagement.TestSuite) error {
	return nil
}
func (f *fakeTestMgmtRepo) GetSuiteByID(_ context.Context, _ uuid.UUID) (*testmanagement.TestSuite, error) {
	return nil, fmt.Errorf("not found")
}
func (f *fakeTestMgmtRepo) ListSuites(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ string, _ int) ([]testmanagement.TestSuite, error) {
	return nil, nil
}
func (f *fakeTestMgmtRepo) UpdateSuite(_ context.Context, _ *testmanagement.TestSuite) error {
	return nil
}
func (f *fakeTestMgmtRepo) DeleteSuite(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeTestMgmtRepo) CreateCase(_ context.Context, _ *testmanagement.TestCase) error {
	return nil
}
func (f *fakeTestMgmtRepo) GetCaseByID(_ context.Context, _ uuid.UUID) (*testmanagement.TestCase, error) {
	return nil, fmt.Errorf("not found")
}
func (f *fakeTestMgmtRepo) ListCases(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ string, _ int) ([]testmanagement.TestCase, error) {
	return nil, nil
}
func (f *fakeTestMgmtRepo) SearchCases(_ context.Context, _ uuid.UUID, _, _ string, _ int) ([]testmanagement.TestCase, string, error) {
	return nil, "", nil
}
func (f *fakeTestMgmtRepo) UpdateCase(_ context.Context, _ *testmanagement.TestCase) error {
	return nil
}
func (f *fakeTestMgmtRepo) DeleteCase(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeTestMgmtRepo) CreateVersion(_ context.Context, _ *testmanagement.TestCaseVersion) error {
	return nil
}
func (f *fakeTestMgmtRepo) ListVersions(_ context.Context, _ uuid.UUID, _ string, _ int) ([]testmanagement.TestCaseVersion, error) {
	return nil, nil
}
func (f *fakeTestMgmtRepo) RunInTx(_ context.Context, fn func(testmanagement.Repository) error) error {
	return fn(f)
}

// ---------- Helpers ----------

func newTestService(t *testing.T) (*Service, *fakeAutomationRepo, *fakeResultsRepo, *fakeDefectsRepo, *ArtifactStorage) {
	aRepo := newFakeAutomationRepo()
	rRepo := newFakeResultsRepo()
	dRepo := newFakeDefectsRepo()
	tmRepo := &fakeTestMgmtRepo{}
	dir, err := os.MkdirTemp("", "automationhub-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	st := NewArtifactStorage(dir)
	return NewService(aRepo, rRepo, dRepo, tmRepo, st), aRepo, rRepo, dRepo, st
}

func newAutomationProject(svc *Service, wsID, linkedProjectID uuid.UUID) *AutomationProject {
	proj, _ := svc.CreateProject(context.Background(), CreateProjectInput{
		WorkspaceID: wsID,
		ProjectID:   &linkedProjectID,
		Name:        "Automation Project",
		Framework:   "junit",
		CreatedBy:   uuid.New(),
	})
	return proj
}

func mustParse(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- Tests ----------

func TestIngestJUnit(t *testing.T) {
	svc, _, rRepo, _, _ := newTestService(t)
	wsID := uuid.New()
	projID := uuid.New()
	uid := uuid.New()

	junitXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="Suite 1" tests="3" failures="1" errors="0" skipped="1" time="2.5">
    <testcase name="Test A" classname="ClassA" time="0.5"/>
    <testcase name="Test B" classname="ClassB" time="1.0">
      <failure message="assertion failed" type="AssertionError">stack trace here</failure>
    </testcase>
    <testcase name="Test C" classname="ClassC" time="1.0" status="skipped"/>
  </testsuite>
</testsuites>`)

	result, err := svc.Ingest(context.Background(), IngestInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		Name:        "CI Build #1",
		Format:      FormatJUnit,
		Body:        junitXML,
		CreatedBy:   uid,
	})
	mustParse(t, err)
	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
	if result.Passed != 1 {
		t.Errorf("expected passed 1, got %d", result.Passed)
	}
	if result.Failed != 1 {
		t.Errorf("expected failed 1, got %d", result.Failed)
	}
	if result.Skipped != 1 {
		t.Errorf("expected skipped 1, got %d", result.Skipped)
	}
	if result.DurationMs != 2500 {
		t.Errorf("expected duration 2500ms, got %d", result.DurationMs)
	}
	if len(rRepo.runs) != 1 || len(rRepo.items) != 3 {
		t.Errorf("expected 1 run and 3 items, got %d runs %d items", len(rRepo.runs), len(rRepo.items))
	}
}

func TestImportExecution(t *testing.T) {
	svc, aRepo, rRepo, dRepo, _ := newTestService(t)
	wsID := uuid.New()
	linkedProjectID := uuid.New()
	uid := uuid.New()
	proj := newAutomationProject(svc, wsID, linkedProjectID)

	junitXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="Suite" tests="2" failures="1" skipped="0" time="1.2">
    <testcase name="Passing" time="0.5"/>
    <testcase name="Failing" time="0.7">
      <failure message="fail">trace</failure>
    </testcase>
  </testsuite>
</testsuites>`)

	res, err := svc.ImportExecution(context.Background(), ImportExecutionInput{
		ProjectID:         proj.ID,
		Name:              "Nightly",
		Format:            FormatJUnit,
		Report:            junitXML,
		CreatedBy:         uid,
		AutoCreateDefects: true,
		MapTestCases:      false,
	})
	mustParse(t, err)
	if res.Total != 2 || res.Failed != 1 || res.Passed != 1 {
		t.Fatalf("unexpected counts: total=%d passed=%d failed=%d", res.Total, res.Passed, res.Failed)
	}

	exec, err := aRepo.GetExecution(context.Background(), res.ExecutionID)
	mustParse(t, err)
	if exec.Status != "passed" && exec.Status != "failed" && exec.Status != "completed" {
		t.Errorf("unexpected execution status %s", exec.Status)
	}
	if exec.ReportPath == "" {
		t.Error("expected report path")
	}
	if len(rRepo.runs) != 1 || len(rRepo.items) != 2 {
		t.Errorf("expected 1 run 2 items, got %d runs %d items", len(rRepo.runs), len(rRepo.items))
	}
	if len(dRepo.defects) != 1 {
		t.Errorf("expected 1 defect, got %d", len(dRepo.defects))
	}
}

func TestImportExecutionWithoutLinkedProject(t *testing.T) {
	svc, aRepo, _, _, _ := newTestService(t)
	wsID := uuid.New()
	uid := uuid.New()

	proj := &AutomationProject{
		ID:          uuid.New(),
		WorkspaceID: wsID,
		ProjectID:   nil,
		Name:        "Orphan",
		Framework:   "junit",
		CreatedBy:   uid,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = aRepo.CreateProject(context.Background(), proj)

	_, err := svc.ImportExecution(context.Background(), ImportExecutionInput{
		ProjectID: proj.ID,
		Name:      "Run",
		Format:    FormatJUnit,
		Report:    []byte(`<testsuites><testsuite name="S" tests="1"><testcase name="A" time="0.1"/></testsuite></testsuites>`),
		CreatedBy: uid,
	})
	if err == nil {
		t.Error("expected error when automation project has no linked project")
	}
}

func TestCreateAndListProjects(t *testing.T) {
	svc, aRepo, _, _, _ := newTestService(t)
	wsID := uuid.New()
	uid := uuid.New()

	_, err := svc.CreateProject(context.Background(), CreateProjectInput{
		WorkspaceID: wsID,
		Name:        "",
		Framework:   "junit",
		CreatedBy:   uid,
	})
	if err == nil {
		t.Error("expected error for empty name")
	}

	proj, err := svc.CreateProject(context.Background(), CreateProjectInput{
		WorkspaceID:   wsID,
		Name:          "Web Tests",
		Framework:     "playwright",
		RepositoryURL: "https://github.com/org/repo",
		Branch:        "main",
		Command:       "npx playwright test",
		CreatedBy:     uid,
	})
	mustParse(t, err)
	if proj.Name != "Web Tests" {
		t.Errorf("expected Web Tests, got %s", proj.Name)
	}

	list, err := svc.ListProjects(context.Background(), wsID, "", 10)
	mustParse(t, err)
	if len(list) != 1 {
		t.Errorf("expected 1 project, got %d", len(list))
	}

	got, err := svc.GetProject(context.Background(), proj.ID)
	mustParse(t, err)
	if got.Framework != "playwright" {
		t.Errorf("expected playwright, got %s", got.Framework)
	}

	updated, err := svc.UpdateProject(context.Background(), proj.ID, UpdateProjectInput{
		Name:      "Web Tests Updated",
		Framework: "cypress",
	})
	mustParse(t, err)
	if updated.Name != "Web Tests Updated" || updated.Framework != "cypress" {
		t.Errorf("unexpected update: %+v", updated)
	}

	mustParse(t, svc.DeleteProject(context.Background(), proj.ID))
	if _, ok := aRepo.projects[proj.ID]; ok {
		t.Error("project not deleted")
	}
}

func TestUploadAndGetArtifact(t *testing.T) {
	svc, aRepo, _, _, st := newTestService(t)
	wsID := uuid.New()
	linkedProjectID := uuid.New()
	proj := newAutomationProject(svc, wsID, linkedProjectID)
	execID := uuid.New()
	aRepo.executions[execID] = &AutomationExecution{
		ID:          execID,
		ProjectID:   proj.ID,
		WorkspaceID: wsID,
		Name:        "Run",
		Status:      "running",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	artifact, err := svc.UploadArtifact(context.Background(), UploadArtifactInput{
		ExecutionID: execID,
		WorkspaceID: wsID,
		Kind:        ArtifactKindScreenshot,
		Name:        "screenshot.png",
		MimeType:    "image/png",
		Data:        []byte("pngdata"),
	})
	mustParse(t, err)
	if artifact.FileSize != 7 {
		t.Errorf("expected file size 7, got %d", artifact.FileSize)
	}

	got, data, err := svc.GetArtifact(context.Background(), artifact.ID)
	mustParse(t, err)
	if string(data) != "pngdata" {
		t.Errorf("unexpected artifact data: %s", string(data))
	}
	if got.Kind != ArtifactKindScreenshot {
		t.Errorf("unexpected artifact kind %s", got.Kind)
	}

	list, err := svc.ListArtifacts(context.Background(), execID, "", 10)
	mustParse(t, err)
	if len(list) != 1 {
		t.Errorf("expected 1 artifact, got %d", len(list))
	}

	mustParse(t, svc.DeleteArtifact(context.Background(), artifact.ID))
	if _, ok := aRepo.artifacts[artifact.ID]; ok {
		t.Error("artifact not deleted from repo")
	}
	if _, err := st.ReadArtifact(artifact.FilePath); err == nil {
		t.Error("artifact not removed from storage")
	}
}

func TestLogs(t *testing.T) {
	svc, aRepo, _, _, _ := newTestService(t)
	wsID := uuid.New()
	execID := uuid.New()
	aRepo.executions[execID] = &AutomationExecution{
		ID:          execID,
		ProjectID:   uuid.New(),
		WorkspaceID: wsID,
		Name:        "Run",
		Status:      "running",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	log, err := svc.AddLog(context.Background(), AddLogInput{
		ExecutionID: execID,
		WorkspaceID: wsID,
		Level:       "info",
		Message:     "started",
		LoggedAt:    time.Now(),
	})
	mustParse(t, err)
	if log.Message != "started" {
		t.Errorf("unexpected log message %s", log.Message)
	}

	logs, err := svc.ListLogs(context.Background(), execID, "", 10)
	mustParse(t, err)
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

func TestIsValidFormat(t *testing.T) {
	valid := []string{"junit", "playwright", "cypress", "newman", "pytest-junit", "robot"}
	for _, s := range valid {
		if !IsValidFormat(s) {
			t.Errorf("expected %s to be valid", s)
		}
	}
	if IsValidFormat("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}

func TestRunStatusFromCounts(t *testing.T) {
	if s := runStatusFromCounts(5, 0); s != results.RunStatusPassed {
		t.Errorf("expected passed, got %s", s)
	}
	if s := runStatusFromCounts(5, 2); s != results.RunStatusFailed {
		t.Errorf("expected failed, got %s", s)
	}
}

func TestArtifactKindValidation(t *testing.T) {
	if !IsValidArtifactKind("report") {
		t.Error("expected report to be valid")
	}
	if IsValidArtifactKind("unknown") {
		t.Error("expected unknown to be invalid")
	}
}

func TestMimeTypeFor(t *testing.T) {
	if mimeTypeFor(FormatJUnit) != "application/xml" {
		t.Errorf("unexpected mime type for junit: %s", mimeTypeFor(FormatJUnit))
	}
	if mimeTypeFor(FormatPlaywright) != "application/json" {
		t.Errorf("unexpected mime type for playwright: %s", mimeTypeFor(FormatPlaywright))
	}
}
