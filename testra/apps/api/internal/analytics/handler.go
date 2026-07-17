package analytics

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

func (h *Handler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkspaceID string                 `json:"workspace_id"`
		Name        string                 `json:"name"`
		Type        string                 `json:"type"`
		Config      map[string]interface{} `json:"config"`
	}
	if err := parseJSON(r, &req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	dashboard, err := h.service.CreateDashboard(r.Context(), CreateDashboardInput{WorkspaceID: wsID, Name: req.Name, Type: req.Type, Config: req.Config}, userID)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, toDashboardResponse(dashboard))
}

func (h *Handler) ListDashboards(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return
	}
	dashboards, err := h.service.ListDashboards(r.Context(), wsID)
	if err != nil {
		mapError(w, err)
		return
	}
	resp := make([]dashboardResponse, len(dashboards))
	for i, d := range dashboards {
		resp[i] = toDashboardResponse(&d)
	}
	apihttp.JSON(w, http.StatusOK, resp)
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	d, err := h.service.GetDashboard(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toDashboardResponse(d))
}

func (h *Handler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	var req struct {
		Name   string                 `json:"name"`
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}
	if err := parseJSON(r, &req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	d, err := h.service.UpdateDashboard(r.Context(), id, UpdateDashboardInput{Name: req.Name, Type: req.Type, Config: req.Config})
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toDashboardResponse(d))
}

func (h *Handler) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	if err := h.service.DeleteDashboard(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusNoContent, nil)
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return
	}
	var projectID *uuid.UUID
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		p, err := uuid.Parse(pid)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
			return
		}
		projectID = &p
	}
	summary, err := h.service.GetSummary(r.Context(), wsID, projectID)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, summary)
}

func (h *Handler) GetTrends(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return
	}
	var projectID *uuid.UUID
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		p, err := uuid.Parse(pid)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
			return
		}
		projectID = &p
	}
	var start, end *time.Time
	if s := r.URL.Query().Get("start"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "start must be YYYY-MM-DD")
			return
		}
		start = &t
	}
	if e := r.URL.Query().Get("end"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "end must be YYYY-MM-DD")
			return
		}
		end = &t
	}
	trends, err := h.service.GetTrends(r.Context(), wsID, projectID, start, end)
	if err != nil {
		mapError(w, err)
		return
	}
	meta := pagination.Meta{HasMore: false}
	apihttp.JSON(w, http.StatusOK, map[string]interface{}{"trends": trends, "meta": meta})
}

// Helpers

type dashboardResponse struct {
	ID          uuid.UUID              `json:"id"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func toDashboardResponse(d *Dashboard) dashboardResponse {
	return dashboardResponse{
		ID:          d.ID,
		WorkspaceID: d.WorkspaceID,
		Name:        d.Name,
		Type:        d.Type,
		Config:      d.Config,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func mapError(w http.ResponseWriter, err error) {
	switch err {
	case sharederrors.ErrNotFound:
		apihttp.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case sharederrors.ErrInvalidInput:
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case sharederrors.ErrForbidden:
		apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	default:
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseJSON(r *http.Request, v interface{}) error {
	// The shared middleware already ensures JSON bodies are readable.
	return json.NewDecoder(r.Body).Decode(v)
}
