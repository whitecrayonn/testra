package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, summary)
}

func (h *Handler) GetTrends(w http.ResponseWriter, r *http.Request) {
	wsID, projectID, start, end := parseTrendParams(w, r)
	if wsID == uuid.Nil {
		return
	}
	trends, err := h.service.GetTrends(r.Context(), wsID, projectID, start, end)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	meta := pagination.Meta{HasMore: false}
	apihttp.JSON(w, http.StatusOK, map[string]interface{}{"trends": trends, "meta": meta})
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseMetricsFilter(w, r)
	if !ok {
		return
	}
	metrics, err := h.service.GetMetrics(r.Context(), filter)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, metrics)
}

func (h *Handler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseMetricsFilter(w, r)
	if !ok {
		return
	}
	activity, err := h.service.GetRecentActivity(r.Context(), filter)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, activity)
}

func (h *Handler) ExportMetricsCSV(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseMetricsFilter(w, r)
	if !ok {
		return
	}
	rows, err := h.service.GetMetricsCSV(r.Context(), filter)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=metrics.csv")
	for _, row := range rows {
		_, _ = w.Write([]byte(strings.Join(row, ",") + "\n"))
	}
}

func parseTrendParams(w http.ResponseWriter, r *http.Request) (uuid.UUID, *uuid.UUID, *time.Time, *time.Time) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return uuid.Nil, nil, nil, nil
	}
	var projectID *uuid.UUID
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		p, err := uuid.Parse(pid)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
			return uuid.Nil, nil, nil, nil
		}
		projectID = &p
	}
	var start, end *time.Time
	if s := r.URL.Query().Get("start"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "start must be YYYY-MM-DD")
			return uuid.Nil, nil, nil, nil
		}
		start = &t
	}
	if e := r.URL.Query().Get("end"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "end must be YYYY-MM-DD")
			return uuid.Nil, nil, nil, nil
		}
		end = &t
	}
	return wsID, projectID, start, end
}

func parseMetricsFilter(w http.ResponseWriter, r *http.Request) (MetricsFilter, bool) {
	filter := MetricsFilter{}
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return filter, false
	}
	filter.WorkspaceID = wsID
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		p, err := uuid.Parse(pid)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
			return filter, false
		}
		filter.ProjectID = &p
	}
	filter.Release = r.URL.Query().Get("release")
	filter.Sprint = r.URL.Query().Get("sprint")
	filter.Environment = r.URL.Query().Get("environment")
	filter.Source = r.URL.Query().Get("source")
	if tester := r.URL.Query().Get("tester"); tester != "" {
		t, err := uuid.Parse(tester)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid tester")
			return filter, false
		}
		filter.TesterID = &t
	}
	if s := r.URL.Query().Get("start"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "start must be YYYY-MM-DD")
			return filter, false
		}
		filter.Start = &t
	}
	if e := r.URL.Query().Get("end"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "end must be YYYY-MM-DD")
			return filter, false
		}
		filter.End = &t
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	return filter, true
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

func parseJSON(r *http.Request, v interface{}) error {
	// The shared middleware already ensures JSON bodies are readable.
	return json.NewDecoder(r.Body).Decode(v)
}
