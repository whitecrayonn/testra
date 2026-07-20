package results

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/middleware"
	"github.com/testra/testra/apps/api/internal/shared/pagination"
)

// ---------- Test run item execution ----------

type executeItemRequest struct {
	Status       string       `json:"status"`
	StepResults  []StepResult `json:"step_results"`
	Comment      string       `json:"comment"`
	DurationMs   int64        `json:"duration_ms"`
	ErrorMessage string       `json:"error_message"`
	StackTrace   string       `json:"stack_trace"`
}

func (h *Handler) ExecuteItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req executeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	uid, _ := middleware.UserIDFromContext(r.Context())
	item, err := h.service.ExecuteItem(r.Context(), itemID, ExecuteItemInput{
		Status:       RunItemStatus(req.Status),
		StepResults:  req.StepResults,
		Comment:      req.Comment,
		DurationMs:   req.DurationMs,
		ErrorMessage: req.ErrorMessage,
		StackTrace:   req.StackTrace,
		ExecutedBy:   uid,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapItemResponse(item)})
}

func (h *Handler) ListItemHistory(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	history, err := h.service.ListItemHistory(r.Context(), itemID)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]historyResponse, len(history))
	for i, h := range history {
		resp[i] = mapHistoryResponse(&h)
	}
	apihttp.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// ---------- Evidence ----------

type evidenceRequest struct {
	StepOrder   int    `json:"step_order"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	StoragePath string `json:"storage_path"`
}

func (h *Handler) AttachEvidence(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req evidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	uid, _ := middleware.UserIDFromContext(r.Context())
	evidence, err := h.service.AttachEvidence(r.Context(), EvidenceInput{
		RunItemID:   itemID,
		StepOrder:   req.StepOrder,
		FileName:    req.FileName,
		ContentType: req.ContentType,
		StoragePath: req.StoragePath,
		UploadedBy:  uid,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, map[string]any{"data": mapEvidenceResponse(evidence)})
}

func (h *Handler) ListEvidence(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	evidence, err := h.service.ListEvidence(r.Context(), itemID)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]evidenceResponse, len(evidence))
	for i, e := range evidence {
		resp[i] = mapEvidenceResponse(&e)
	}
	apihttp.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

func (h *Handler) DeleteEvidence(w http.ResponseWriter, r *http.Request) {
	evidenceID, err := uuid.Parse(chi.URLParam(r, "evidenceId"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid evidence id")
		return
	}

	if err := h.service.DeleteEvidence(r.Context(), evidenceID); err != nil {
		apihttp.MapError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Defect links ----------

type linkDefectRequest struct {
	DefectID string `json:"defect_id"`
}

func (h *Handler) LinkDefect(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req linkDefectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	defectID, err := uuid.Parse(req.DefectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid defect_id")
		return
	}

	if err := h.service.LinkDefect(r.Context(), itemID, defectID); err != nil {
		apihttp.MapError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListItemDefects(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	ids, err := h.service.ListItemDefects(r.Context(), itemID)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]string, len(ids))
	for i, id := range ids {
		resp[i] = id.String()
	}
	apihttp.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

func (h *Handler) UnlinkDefect(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	defectID, err := uuid.Parse(chi.URLParam(r, "defectId"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid defect id")
		return
	}

	if err := h.service.UnlinkDefect(r.Context(), itemID, defectID); err != nil {
		apihttp.MapError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Bulk and clone/run operations ----------

type bulkUpdateItemsRequest struct {
	ItemIDs []string `json:"item_ids"`
	Status  string   `json:"status"`
}

func (h *Handler) BulkUpdateItems(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req bulkUpdateItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	itemIDs := make([]uuid.UUID, 0, len(req.ItemIDs))
	for _, idStr := range req.ItemIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid item id")
			return
		}
		itemIDs = append(itemIDs, id)
	}

	uid, _ := middleware.UserIDFromContext(r.Context())
	items, err := h.service.BulkUpdateItems(r.Context(), runID, itemIDs, RunItemStatus(req.Status), uid)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]itemResponse, len(items))
	for i, item := range items {
		resp[i] = mapItemResponse(&item)
	}
	apihttp.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

func (h *Handler) CloneRun(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	uid, _ := middleware.UserIDFromContext(r.Context())
	run, err := h.service.CloneRun(r.Context(), runID, uid)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, map[string]any{"data": mapRunResponse(run)})
}

func (h *Handler) RerunRun(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	uid, _ := middleware.UserIDFromContext(r.Context())
	run, err := h.service.RerunRun(r.Context(), runID, uid)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, map[string]any{"data": mapRunResponse(run)})
}

// Override ListItems to support optional filtering and pagination.
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	params := pagination.ParseParams(r)

	items, err := h.service.ListItemsPaged(r.Context(), runID, status, search, params.Cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]itemResponse, len(items))
	for i, item := range items {
		resp[i] = mapItemResponse(&item)
	}

	meta := pagination.Meta{HasMore: len(items) == params.Limit}
	if meta.HasMore && len(items) > 0 {
		nextCursor, err := pagination.EncodeCursor(items[len(items)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

// ---------- Test plans ----------

type createPlanRequest struct {
	WorkspaceID   string                 `json:"workspace_id"`
	ProjectID     string                 `json:"project_id"`
	SuiteID       *string                `json:"suite_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Configuration map[string]interface{} `json:"configuration"`
	TestCaseIDs   []string               `json:"test_case_ids"`
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req createPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
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

	tcIDs := make([]uuid.UUID, 0, len(req.TestCaseIDs))
	for _, idStr := range req.TestCaseIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test_case_id")
			return
		}
		tcIDs = append(tcIDs, id)
	}

	uid, _ := middleware.UserIDFromContext(r.Context())
	plan, err := h.service.CreatePlan(r.Context(), CreatePlanInput{
		WorkspaceID:   wsID,
		ProjectID:     projectID,
		SuiteID:       suiteID,
		Name:          req.Name,
		Description:   req.Description,
		Configuration: req.Configuration,
		CreatedBy:     uid,
		TestCaseIDs:   tcIDs,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, map[string]any{"data": mapPlanResponse(plan)})
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
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
	plans, err := h.service.ListPlans(r.Context(), projectID, params.Cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]planResponse, len(plans))
	for i, plan := range plans {
		resp[i] = mapPlanResponse(&plan)
	}

	meta := pagination.Meta{HasMore: len(plans) == params.Limit}
	if meta.HasMore && len(plans) > 0 {
		nextCursor, err := pagination.EncodeCursor(plans[len(plans)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	plan, err := h.service.GetPlan(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapPlanResponse(plan)})
}

type updatePlanRequest struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Status        string                 `json:"status"`
	Configuration map[string]interface{} `json:"configuration"`
	TestCaseIDs   []string               `json:"test_case_ids"`
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req updatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	var tcIDs []uuid.UUID
	if req.TestCaseIDs != nil {
		tcIDs = make([]uuid.UUID, 0, len(req.TestCaseIDs))
		for _, idStr := range req.TestCaseIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test_case_id")
				return
			}
			tcIDs = append(tcIDs, id)
		}
	}

	var status TestPlanStatus
	if req.Status != "" {
		status = TestPlanStatus(req.Status)
	}

	plan, err := h.service.UpdatePlan(r.Context(), id, UpdatePlanInput{
		Name:          req.Name,
		Description:   req.Description,
		Status:        status,
		Configuration: req.Configuration,
		TestCaseIDs:   tcIDs,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapPlanResponse(plan)})
}

func (h *Handler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	if err := h.service.DeletePlan(r.Context(), id); err != nil {
		apihttp.MapError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPlanItems(w http.ResponseWriter, r *http.Request) {
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	items, err := h.service.GetPlanItems(r.Context(), planID)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]planItemResponse, len(items))
	for i, item := range items {
		resp[i] = mapPlanItemResponse(&item)
	}
	apihttp.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

func (h *Handler) CreateRunFromPlan(w http.ResponseWriter, r *http.Request) {
	planID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	uid, _ := middleware.UserIDFromContext(r.Context())
	run, err := h.service.CreateRunFromPlan(r.Context(), planID, uid)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, map[string]any{"data": mapRunResponse(run)})
}
