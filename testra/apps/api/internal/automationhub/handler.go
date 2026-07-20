package automationhub

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/middleware"
	"github.com/testra/testra/apps/api/internal/shared/pagination"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func userID(r *http.Request) uuid.UUID {
	uid, _ := middleware.UserIDFromContext(r.Context())
	return uid
}

// ----------------- Legacy /ingest -----------------

type ingestRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ProjectID   string `json:"project_id"`
	SuiteID     string `json:"suite_id"`
	Name        string `json:"name"`
	Format      string `json:"format"`
	Payload     string `json:"payload"`
}

type ingestResponse struct {
	RunID       string `json:"run_id"`
	ExecutionID string `json:"execution_id,omitempty"`
	Total       int    `json:"total"`
	Passed      int    `json:"passed"`
	Failed      int    `json:"failed"`
	Skipped     int    `json:"skipped"`
	DurationMs  int64  `json:"duration_ms"`
}

func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "failed to read request body")
		return
	}
	defer r.Body.Close()

	var meta ingestRequest
	if err := json.Unmarshal(body, &meta); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	wsID, err := uuid.Parse(meta.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	projID, err := uuid.Parse(meta.ProjectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
		return
	}

	var suiteID *uuid.UUID
	if meta.SuiteID != "" {
		sid, err := uuid.Parse(meta.SuiteID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite_id")
			return
		}
		suiteID = &sid
	}

	if !IsValidFormat(meta.Format) {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "unsupported format")
		return
	}

	if meta.Payload == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "payload is required")
		return
	}

	result, err := h.service.Ingest(r.Context(), IngestInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		SuiteID:     suiteID,
		Name:        meta.Name,
		Format:      IngestionFormat(meta.Format),
		Body:        []byte(meta.Payload),
		CreatedBy:   userID(r),
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, ingestResponse{
		RunID:      result.RunID.String(),
		Total:      result.Total,
		Passed:     result.Passed,
		Failed:     result.Failed,
		Skipped:    result.Skipped,
		DurationMs: result.DurationMs,
	})
}

// ----------------- Projects -----------------

type createProjectRequest struct {
	WorkspaceID   string `json:"workspace_id"`
	ProjectID     string `json:"project_id,omitempty"`
	Name          string `json:"name"`
	Framework     string `json:"framework"`
	RepositoryURL string `json:"repository_url"`
	Branch        string `json:"branch"`
	Command       string `json:"command"`
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	var projectID *uuid.UUID
	if req.ProjectID != "" {
		pid, err := uuid.Parse(req.ProjectID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
			return
		}
		projectID = &pid
	}

	project, err := h.service.CreateProject(r.Context(), CreateProjectInput{
		WorkspaceID:   wsID,
		ProjectID:     projectID,
		Name:          req.Name,
		Framework:     req.Framework,
		RepositoryURL: req.RepositoryURL,
		Branch:        req.Branch,
		Command:       req.Command,
		CreatedBy:     userID(r),
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapProjectResponse(project))
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project id")
		return
	}

	project, err := h.service.GetProject(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapProjectResponse(project))
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id is required")
		return
	}
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	params := pagination.ParseParams(r)
	projects, err := h.service.ListProjects(r.Context(), wsID, params.Cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]projectResponse, len(projects))
	for i, p := range projects {
		resp[i] = mapProjectResponse(&p)
	}

	meta := pagination.Meta{HasMore: len(projects) == params.Limit}
	if meta.HasMore && len(projects) > 0 {
		nextCursor, err := pagination.EncodeCursor(projects[len(projects)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

type updateProjectRequest struct {
	Name          string `json:"name"`
	Framework     string `json:"framework"`
	RepositoryURL string `json:"repository_url"`
	Branch        string `json:"branch"`
	Command       string `json:"command"`
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project id")
		return
	}

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	project, err := h.service.UpdateProject(r.Context(), id, UpdateProjectInput{
		Name:          req.Name,
		Framework:     req.Framework,
		RepositoryURL: req.RepositoryURL,
		Branch:        req.Branch,
		Command:       req.Command,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapProjectResponse(project))
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project id")
		return
	}

	if err := h.service.DeleteProject(r.Context(), id); err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ----------------- Executions -----------------

type importExecutionForm struct {
	ProjectID         string
	Name              string
	Format            string
	Report            []byte
	AutoCreateDefects bool
	MapTestCases      bool
}

func parseImportExecutionForm(r *http.Request) (importExecutionForm, error) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		return importExecutionForm{}, err
	}
	defer r.MultipartForm.RemoveAll()

	projIDStr := chi.URLParam(r, "project_id")
	if projIDStr == "" {
		projIDStr = r.FormValue("project_id")
	}
	projID, err := uuid.Parse(projIDStr)
	if err != nil {
		return importExecutionForm{}, err
	}

	file, _, err := r.FormFile("report")
	if err != nil {
		return importExecutionForm{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return importExecutionForm{}, err
	}

	form := importExecutionForm{
		ProjectID:         projID.String(),
		Name:              r.FormValue("name"),
		Format:            r.FormValue("format"),
		Report:            data,
		AutoCreateDefects: r.FormValue("auto_create_defects") == "true",
		MapTestCases:      r.FormValue("map_test_cases") == "true",
	}
	return form, nil
}

func (h *Handler) ImportExecution(w http.ResponseWriter, r *http.Request) {
	form, err := parseImportExecutionForm(r)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	projID, err := uuid.Parse(form.ProjectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
		return
	}

	format := IngestionFormat(form.Format)
	if !IsValidFormat(string(format)) {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "unsupported format")
		return
	}

	result, err := h.service.ImportExecution(r.Context(), ImportExecutionInput{
		ProjectID:         projID,
		Name:              form.Name,
		Format:            format,
		Report:            form.Report,
		CreatedBy:         userID(r),
		AutoCreateDefects: form.AutoCreateDefects,
		MapTestCases:      form.MapTestCases,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, ingestResponse{
		RunID:       result.RunID.String(),
		ExecutionID: result.ExecutionID.String(),
		Total:       result.Total,
		Passed:      result.Passed,
		Failed:      result.Failed,
		Skipped:     result.Skipped,
		DurationMs:  result.DurationMs,
	})
}

func (h *Handler) GetExecution(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid execution id")
		return
	}

	execution, err := h.service.GetExecution(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapExecutionResponse(execution))
}

func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	params := pagination.ParseParams(r)

	projIDStr := chi.URLParam(r, "project_id")
	if projIDStr == "" {
		projIDStr = r.URL.Query().Get("project_id")
	}
	if projIDStr != "" {
		projID, err := uuid.Parse(projIDStr)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
			return
		}
		executions, err := h.service.ListExecutions(r.Context(), projID, params.Cursor, params.Limit)
		if err != nil {
			apihttp.MapError(w, err)
			return
		}
		writeExecutionList(w, executions, params.Limit)
		return
	}

	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "project_id or workspace_id is required")
		return
	}
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	executions, err := h.service.ListExecutionsByWorkspace(r.Context(), wsID, params.Cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	writeExecutionList(w, executions, params.Limit)
}

func (h *Handler) RerunExecution(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid execution id")
		return
	}

	execution, err := h.service.RerunExecution(r.Context(), id, userID(r))
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapExecutionResponse(execution))
}

func (h *Handler) DeleteExecution(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid execution id")
		return
	}

	if err := h.service.DeleteExecution(r.Context(), id); err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ----------------- Artifacts -----------------

func (h *Handler) UploadArtifact(w http.ResponseWriter, r *http.Request) {
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid execution id")
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "file is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	kind := ArtifactKind(r.FormValue("kind"))
	if !IsValidArtifactKind(string(kind)) {
		kind = ArtifactKindArtifact
	}

	wsID, err := uuid.Parse(r.FormValue("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	var itemID *uuid.UUID
	if s := r.FormValue("test_run_item_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test_run_item_id")
			return
		}
		itemID = &id
	}

	artifact, err := h.service.UploadArtifact(r.Context(), UploadArtifactInput{
		ExecutionID:   execID,
		WorkspaceID:   wsID,
		TestRunItemID: itemID,
		Kind:          kind,
		Name:          header.Filename,
		MimeType:      header.Header.Get("Content-Type"),
		Data:          data,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapArtifactResponse(artifact))
}

func (h *Handler) ListArtifacts(w http.ResponseWriter, r *http.Request) {
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid execution id")
		return
	}

	params := pagination.ParseParams(r)
	artifacts, err := h.service.ListArtifacts(r.Context(), execID, params.Cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]artifactResponse, len(artifacts))
	for i, a := range artifacts {
		resp[i] = mapArtifactResponse(&a)
	}

	meta := pagination.Meta{HasMore: len(artifacts) == params.Limit}
	if meta.HasMore && len(artifacts) > 0 {
		nextCursor, err := pagination.EncodeCursor(artifacts[len(artifacts)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) GetArtifact(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid artifact id")
		return
	}

	artifact, data, err := h.service.GetArtifact(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	if artifact.MimeType != "" {
		w.Header().Set("Content-Type", artifact.MimeType)
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+artifact.Name+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) DeleteArtifact(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid artifact id")
		return
	}

	if err := h.service.DeleteArtifact(r.Context(), id); err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ----------------- Logs -----------------

type addLogRequest struct {
	Level    string `json:"level"`
	Message  string `json:"message"`
	LoggedAt string `json:"logged_at,omitempty"`
}

func (h *Handler) AddLog(w http.ResponseWriter, r *http.Request) {
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid execution id")
		return
	}

	var req addLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	loggedAt, _ := parseTime(req.LoggedAt)
	logEntry, err := h.service.AddLog(r.Context(), AddLogInput{
		ExecutionID: execID,
		WorkspaceID: wsID,
		Level:       req.Level,
		Message:     req.Message,
		LoggedAt:    loggedAt,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapLogResponse(logEntry))
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	execID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid execution id")
		return
	}

	params := pagination.ParseParams(r)
	logs, err := h.service.ListLogs(r.Context(), execID, params.Cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]logResponse, len(logs))
	for i, l := range logs {
		resp[i] = mapLogResponse(&l)
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": pagination.Meta{HasMore: len(logs) == params.Limit},
	})
}

// ----------------- response mapping -----------------

type projectResponse struct {
	ID            string  `json:"id"`
	WorkspaceID   string  `json:"workspace_id"`
	ProjectID     *string `json:"project_id,omitempty"`
	Name          string  `json:"name"`
	Framework     string  `json:"framework"`
	RepositoryURL string  `json:"repository_url"`
	Branch        string  `json:"branch"`
	Command       string  `json:"command"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func mapProjectResponse(p *AutomationProject) projectResponse {
	resp := projectResponse{
		ID:            p.ID.String(),
		WorkspaceID:   p.WorkspaceID.String(),
		Name:          p.Name,
		Framework:     p.Framework,
		RepositoryURL: p.RepositoryURL,
		Branch:        p.Branch,
		Command:       p.Command,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
	if p.ProjectID != nil {
		s := p.ProjectID.String()
		resp.ProjectID = &s
	}
	return resp
}

type executionResponse struct {
	ID           string  `json:"id"`
	ProjectID    string  `json:"project_id"`
	WorkspaceID  string  `json:"workspace_id"`
	TestRunID    *string `json:"test_run_id,omitempty"`
	Name         string  `json:"name"`
	Status       string  `json:"status"`
	ReportFormat string  `json:"report_format"`
	ReportPath   string  `json:"report_path,omitempty"`
	RetryOf      *string `json:"retry_of,omitempty"`
	DurationMs   int64   `json:"duration_ms"`
	Total        int     `json:"total"`
	Passed       int     `json:"passed"`
	Failed       int     `json:"failed"`
	Skipped      int     `json:"skipped"`
	Blocked      int     `json:"blocked"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func mapExecutionResponse(e *AutomationExecution) executionResponse {
	resp := executionResponse{
		ID:           e.ID.String(),
		ProjectID:    e.ProjectID.String(),
		WorkspaceID:  e.WorkspaceID.String(),
		Name:         e.Name,
		Status:       e.Status,
		ReportFormat: e.ReportFormat,
		ReportPath:   e.ReportPath,
		DurationMs:   e.DurationMs,
		Total:        e.Total,
		Passed:       e.Passed,
		Failed:       e.Failed,
		Skipped:      e.Skipped,
		Blocked:      e.Blocked,
		CreatedAt:    e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    e.UpdatedAt.Format(time.RFC3339),
	}
	if e.TestRunID != nil {
		s := e.TestRunID.String()
		resp.TestRunID = &s
	}
	if e.RetryOf != nil {
		s := e.RetryOf.String()
		resp.RetryOf = &s
	}
	return resp
}

type artifactResponse struct {
	ID            string                 `json:"id"`
	ExecutionID   string                 `json:"execution_id"`
	TestRunItemID *string                `json:"test_run_item_id,omitempty"`
	Kind          string                 `json:"kind"`
	Name          string                 `json:"name"`
	MimeType      string                 `json:"mime_type"`
	FileSize      int64                  `json:"file_size"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     string                 `json:"created_at"`
}

func mapArtifactResponse(a *AutomationArtifact) artifactResponse {
	resp := artifactResponse{
		ID:          a.ID.String(),
		ExecutionID: a.ExecutionID.String(),
		Kind:        string(a.Kind),
		Name:        a.Name,
		MimeType:    a.MimeType,
		FileSize:    a.FileSize,
		Metadata:    a.Metadata,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
	}
	if a.TestRunItemID != nil {
		s := a.TestRunItemID.String()
		resp.TestRunItemID = &s
	}
	return resp
}

type logResponse struct {
	ID        string `json:"id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	LoggedAt  string `json:"logged_at"`
	CreatedAt string `json:"created_at"`
}

func mapLogResponse(l *AutomationLog) logResponse {
	return logResponse{
		ID:        l.ID.String(),
		Level:     l.Level,
		Message:   l.Message,
		LoggedAt:  l.LoggedAt.Format(time.RFC3339),
		CreatedAt: l.CreatedAt.Format(time.RFC3339),
	}
}

func writeExecutionList(w http.ResponseWriter, executions []AutomationExecution, limit int) {
	resp := make([]executionResponse, len(executions))
	for i, e := range executions {
		resp[i] = mapExecutionResponse(&e)
	}

	meta := pagination.Meta{HasMore: len(executions) == limit}
	if meta.HasMore && len(executions) > 0 {
		nextCursor, err := pagination.EncodeCursor(executions[len(executions)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	return t, err
}
