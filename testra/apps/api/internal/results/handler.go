package results

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
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

type runResponse struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	ProjectID   string     `json:"project_id"`
	SuiteID     *string    `json:"suite_id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Total       int        `json:"total"`
	Passed      int        `json:"passed"`
	Failed      int        `json:"failed"`
	Skipped     int        `json:"skipped"`
	Blocked     int        `json:"blocked"`
	DurationMs  int64      `json:"duration_ms"`
	Source      string     `json:"source"`
	CreatedBy   string     `json:"created_by"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type itemResponse struct {
	ID           string    `json:"id"`
	RunID        string    `json:"run_id"`
	TestCaseID   *string   `json:"test_case_id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	DurationMs   int64     `json:"duration_ms"`
	ErrorMessage string    `json:"error_message"`
	StackTrace   string    `json:"stack_trace"`
	Artifacts    []string  `json:"artifacts"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func mapRunResponse(run *TestRun) runResponse {
	resp := runResponse{
		ID:          run.ID.String(),
		WorkspaceID: run.WorkspaceID.String(),
		ProjectID:   run.ProjectID.String(),
		Name:        run.Name,
		Status:      string(run.Status),
		Total:       run.Total,
		Passed:      run.Passed,
		Failed:      run.Failed,
		Skipped:     run.Skipped,
		Blocked:     run.Blocked,
		DurationMs:  run.DurationMs,
		Source:      string(run.Source),
		CreatedBy:   run.CreatedBy.String(),
		StartedAt:   run.StartedAt,
		CompletedAt: run.CompletedAt,
		CreatedAt:   run.CreatedAt,
		UpdatedAt:   run.UpdatedAt,
	}
	if run.SuiteID != nil {
		resp.SuiteID = strPtr(run.SuiteID.String())
	}
	return resp
}

func mapItemResponse(item *TestRunItem) itemResponse {
	resp := itemResponse{
		ID:           item.ID.String(),
		RunID:        item.RunID.String(),
		Title:        item.Title,
		Status:       string(item.Status),
		DurationMs:   item.DurationMs,
		ErrorMessage: item.ErrorMessage,
		StackTrace:   item.StackTrace,
		Artifacts:    item.Artifacts,
		SortOrder:    item.SortOrder,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
	if item.TestCaseID != nil {
		resp.TestCaseID = strPtr(item.TestCaseID.String())
	}
	return resp
}

func strPtr(s string) *string { return &s }

func mapError(w http.ResponseWriter, err error) {
	switch err {
	case sharederrors.ErrConflict:
		apihttp.ErrorJSON(w, http.StatusConflict, "CONFLICT", err.Error())
	case sharederrors.ErrNotFound:
		apihttp.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case sharederrors.ErrInvalidInput:
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case sharederrors.ErrForbidden:
		apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	default:
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred")
	}
}

type createRunRequest struct {
	WorkspaceID string   `json:"workspace_id"`
	ProjectID   string   `json:"project_id"`
	SuiteID     *string  `json:"suite_id"`
	Name        string   `json:"name"`
	TestCaseIDs []string `json:"test_case_ids"`
	Source      string   `json:"source"`
}

func (h *Handler) CreateRun(w http.ResponseWriter, r *http.Request) {
	var req createRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
		return
	}

	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	var suiteID *uuid.UUID
	if req.SuiteID != nil && *req.SuiteID != "" {
		sid, err := uuid.Parse(*req.SuiteID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite_id")
			return
		}
		suiteID = &sid
	}

	uid, _ := middleware.UserIDFromContext(r.Context())

	source := RunSourceManual
	if req.Source != "" {
		source = RunSource(req.Source)
	}

	var tcIDs []uuid.UUID
	for _, idStr := range req.TestCaseIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test_case_id")
			return
		}
		tcIDs = append(tcIDs, id)
	}

	run, err := h.service.CreateRun(r.Context(), CreateRunInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		SuiteID:     suiteID,
		Name:        req.Name,
		Source:      source,
		CreatedBy:   uid,
		TestCaseIDs: tcIDs,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, map[string]any{"data": mapRunResponse(run)})
}

func (h *Handler) GetRun(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	run, err := h.service.GetRun(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapRunResponse(run)})
}

func (h *Handler) ListRuns(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "project_id is required")
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
		return
	}

	params := pagination.ParseParams(r)
	runs, err := h.service.ListRuns(r.Context(), projectID, params.Cursor, params.Limit)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]runResponse, len(runs))
	for i, run := range runs {
		resp[i] = mapRunResponse(&run)
	}

	meta := pagination.Meta{HasMore: len(runs) == params.Limit}
	if meta.HasMore && len(runs) > 0 {
		nextCursor, err := pagination.EncodeCursor(runs[len(runs)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

type updateRunStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateRunStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req updateRunStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	run, err := h.service.UpdateRunStatus(r.Context(), id, RunStatus(req.Status))
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapRunResponse(run)})
}

func (h *Handler) DeleteRun(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	if err := h.service.DeleteRun(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	items, err := h.service.ListItems(r.Context(), runID)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]itemResponse, len(items))
	for i, item := range items {
		resp[i] = mapItemResponse(&item)
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

type updateItemStatusRequest struct {
	Status       string `json:"status"`
	DurationMs   int64  `json:"duration_ms"`
	ErrorMessage string `json:"error_message"`
	StackTrace   string `json:"stack_trace"`
}

func (h *Handler) UpdateItemStatus(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req updateItemStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	item, err := h.service.UpdateItemStatus(r.Context(), itemID, RunItemStatus(req.Status), req.DurationMs, req.ErrorMessage, req.StackTrace)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapItemResponse(item)})
}

func (h *Handler) StreamRunProgress(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	ch, err := h.service.SubscribeRunProgress(r.Context(), runID)
	if err != nil {
		mapError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	ctx := r.Context()
	for {
		select {
		case event, open := <-ch:
			if !open {
				return
			}
			data, _ := json.Marshal(event)
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(data)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}
