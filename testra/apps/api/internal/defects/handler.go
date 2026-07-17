package defects

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

type defectResponse struct {
	ID            string  `json:"id"`
	WorkspaceID   string  `json:"workspace_id"`
	ProjectID     string  `json:"project_id"`
	TestRunItemID *string `json:"test_run_item_id,omitempty"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	Severity      string  `json:"severity"`
	Priority      string  `json:"priority"`
	Status        string  `json:"status"`
	AssignedTo    *string `json:"assigned_to,omitempty"`
	CreatedBy     string  `json:"created_by"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type createDefectRequest struct {
	WorkspaceID   string  `json:"workspace_id"`
	ProjectID     string  `json:"project_id"`
	TestRunItemID *string `json:"test_run_item_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Severity      string  `json:"severity"`
	Priority      string  `json:"priority"`
	Status        string  `json:"status"`
	AssignedTo    *string `json:"assigned_to"`
}

type updateDefectRequest struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	Severity      *string `json:"severity"`
	Priority      *string `json:"priority"`
	Status        *string `json:"status"`
	AssignedTo    *string `json:"assigned_to"`
	TestRunItemID *string `json:"test_run_item_id"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createDefectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
		return
	}

	var testRunItemID *uuid.UUID
	if req.TestRunItemID != nil && *req.TestRunItemID != "" {
		id, err := uuid.Parse(*req.TestRunItemID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test_run_item_id")
			return
		}
		testRunItemID = &id
	}

	var assignedTo *uuid.UUID
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		id, err := uuid.Parse(*req.AssignedTo)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid assigned_to")
			return
		}
		assignedTo = &id
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	d, err := h.service.Create(r.Context(), CreateInput{
		WorkspaceID:   workspaceID,
		ProjectID:     projectID,
		TestRunItemID: testRunItemID,
		Title:         req.Title,
		Description:   req.Description,
		Severity:      DefectSeverity(req.Severity),
		Priority:      DefectPriority(req.Priority),
		Status:        DefectStatus(req.Status),
		AssignedTo:    assignedTo,
		CreatedBy:     userID,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, toDefectResponse(d))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid defect id")
		return
	}

	d, err := h.service.Get(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toDefectResponse(d))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(r.URL.Query().Get("project_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
		return
	}

	params := pagination.ParseParams(r)
	cursor := params.Cursor
	if cursor != "" {
		decoded, err := pagination.DecodeCursor(cursor)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid cursor")
			return
		}
		cursor = decoded
	}

	defects, err := h.service.List(r.Context(), projectID, cursor, params.Limit)
	if err != nil {
		mapError(w, err)
		return
	}

	meta := pagination.Meta{HasMore: len(defects) == params.Limit}
	if meta.HasMore && len(defects) > 0 {
		nextCursor, err := pagination.EncodeCursor(defects[len(defects)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	resp := make([]defectResponse, len(defects))
	for i, d := range defects {
		resp[i] = toDefectResponse(&d)
	}

	apihttp.JSON(w, http.StatusOK, map[string]interface{}{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid defect id")
		return
	}

	var req updateDefectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	input := UpdateInput{}
	if req.Title != nil {
		input.Title = req.Title
	}
	if req.Description != nil {
		input.Description = req.Description
	}
	if req.Severity != nil {
		sev := DefectSeverity(*req.Severity)
		input.Severity = &sev
	}
	if req.Priority != nil {
		pri := DefectPriority(*req.Priority)
		input.Priority = &pri
	}
	if req.Status != nil {
		st := DefectStatus(*req.Status)
		input.Status = &st
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo == "" {
			empty := uuid.Nil
			input.AssignedTo = &empty
		} else {
			id, err := uuid.Parse(*req.AssignedTo)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid assigned_to")
				return
			}
			input.AssignedTo = &id
		}
	}
	if req.TestRunItemID != nil {
		if *req.TestRunItemID == "" {
			input.TestRunItemID = &uuid.Nil
		} else {
			id, err := uuid.Parse(*req.TestRunItemID)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test_run_item_id")
				return
			}
			input.TestRunItemID = &id
		}
	}

	d, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toDefectResponse(d))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid defect id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toDefectResponse(d *Defect) defectResponse {
	resp := defectResponse{
		ID:          d.ID.String(),
		WorkspaceID: d.WorkspaceID.String(),
		ProjectID:   d.ProjectID.String(),
		Title:       d.Title,
		Description: d.Description,
		Severity:    string(d.Severity),
		Priority:    string(d.Priority),
		Status:      string(d.Status),
		CreatedBy:   d.CreatedBy.String(),
		CreatedAt:   d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   d.UpdatedAt.Format(time.RFC3339),
	}
	if d.TestRunItemID != nil {
		s := d.TestRunItemID.String()
		resp.TestRunItemID = &s
	}
	if d.AssignedTo != nil {
		s := d.AssignedTo.String()
		resp.AssignedTo = &s
	}
	return resp
}

func mapError(w http.ResponseWriter, err error) {
	switch err {
	case sharederrors.ErrNotFound:
		apihttp.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case sharederrors.ErrInvalidInput:
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case sharederrors.ErrUnauthorized:
		apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	default:
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
	}
}
