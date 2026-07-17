package integrationhub

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkspaceID string            `json:"workspace_id"`
		Type        string            `json:"type"`
		Name        string            `json:"name"`
		Config      map[string]string `json:"config"`
		Enabled     bool              `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	i, err := h.service.CreateIntegration(r.Context(), CreateIntegrationInput{
		WorkspaceID: wsID,
		Type:        req.Type,
		Name:        req.Name,
		Config:      req.Config,
		Enabled:     req.Enabled,
	}, userID)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, toIntegrationResponse(i))
}

func (h *Handler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return
	}
	integrations, err := h.service.ListIntegrations(r.Context(), wsID)
	if err != nil {
		mapError(w, err)
		return
	}
	resp := make([]integrationResponse, len(integrations))
	for i, in := range integrations {
		resp[i] = toIntegrationResponse(&in)
	}
	apihttp.JSON(w, http.StatusOK, resp)
}

func (h *Handler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	i, err := h.service.GetIntegration(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toIntegrationResponse(i))
}

func (h *Handler) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	var req struct {
		Type    string            `json:"type"`
		Name    string            `json:"name"`
		Config  map[string]string `json:"config"`
		Enabled *bool             `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	i, err := h.service.UpdateIntegration(r.Context(), id, UpdateIntegrationInput{
		Type:    req.Type,
		Name:    req.Name,
		Config:  req.Config,
		Enabled: req.Enabled,
	})
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toIntegrationResponse(i))
}

func (h *Handler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	if err := h.service.DeleteIntegration(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusNoContent, nil)
}

func (h *Handler) TestIntegration(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	if err := h.service.TestIntegration(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) DispatchEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkspaceID   string                 `json:"workspace_id"`
		IntegrationID string                 `json:"integration_id"`
		EventType     string                 `json:"event_type"`
		Payload       map[string]interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}
	intID, err := uuid.Parse(req.IntegrationID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid integration_id")
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	e, err := h.service.DispatchEvent(r.Context(), DispatchEventInput{
		WorkspaceID:   wsID,
		IntegrationID: intID,
		EventType:     req.EventType,
		Payload:       req.Payload,
	}, userID)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, toEventResponse(e))
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	events, err := h.service.ListEvents(r.Context(), wsID, limit)
	if err != nil {
		mapError(w, err)
		return
	}
	resp := make([]eventResponse, len(events))
	for i, e := range events {
		resp[i] = toEventResponse(&e)
	}
	apihttp.JSON(w, http.StatusOK, resp)
}

// response helpers

type integrationResponse struct {
	ID        uuid.UUID         `json:"id"`
	Workspace uuid.UUID         `json:"workspace_id"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Config    map[string]string `json:"config"`
	Enabled   bool              `json:"enabled"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

func toIntegrationResponse(i *Integration) integrationResponse {
	return integrationResponse{
		ID:        i.ID,
		Workspace: i.WorkspaceID,
		Type:      string(i.Type),
		Name:      i.Name,
		Config:    i.Config,
		Enabled:   i.Enabled,
		CreatedAt: i.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: i.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type eventResponse struct {
	ID            uuid.UUID              `json:"id"`
	WorkspaceID   uuid.UUID              `json:"workspace_id"`
	IntegrationID *uuid.UUID             `json:"integration_id,omitempty"`
	EventType     string                 `json:"event_type"`
	Payload       map[string]interface{} `json:"payload"`
	Status        string                 `json:"status"`
	ExternalID    string                 `json:"external_id"`
	CreatedAt     string                 `json:"created_at"`
}

func toEventResponse(e *IntegrationEvent) eventResponse {
	return eventResponse{
		ID:            e.ID,
		WorkspaceID:   e.WorkspaceID,
		IntegrationID: e.IntegrationID,
		EventType:     e.EventType,
		Payload:       e.Payload,
		Status:        e.Status,
		ExternalID:    e.ExternalID,
		CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
