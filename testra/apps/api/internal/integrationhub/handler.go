package integrationhub

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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
	userID, _ := middleware.UserIDFromContext(r.Context())
	i, err := h.service.UpdateIntegration(r.Context(), id, UpdateIntegrationInput{
		Type:    req.Type,
		Name:    req.Name,
		Config:  req.Config,
		Enabled: req.Enabled,
	}, userID)
	if err != nil {
		apihttp.MapError(w, err)
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
	userID, _ := middleware.UserIDFromContext(r.Context())
	if err := h.service.DeleteIntegration(r.Context(), id, userID); err != nil {
		apihttp.MapError(w, err)
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
	userID, _ := middleware.UserIDFromContext(r.Context())
	if err := h.service.TestIntegration(r.Context(), id, userID); err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetIntegrationHealth(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	status, err := h.service.HealthStatus(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) EnableIntegration(w http.ResponseWriter, r *http.Request) {
	h.setEnabled(w, r, true)
}

func (h *Handler) DisableIntegration(w http.ResponseWriter, r *http.Request) {
	h.setEnabled(w, r, false)
}

func (h *Handler) setEnabled(w http.ResponseWriter, r *http.Request, enabled bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	var err2 error
	if enabled {
		err2 = h.service.EnableIntegration(r.Context(), id, userID)
	} else {
		err2 = h.service.DisableIntegration(r.Context(), id, userID)
	}
	if err2 != nil {
		apihttp.MapError(w, err2)
		return
	}
	apihttp.JSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
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
		apihttp.MapError(w, err)
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
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	events, err := h.service.ListEvents(r.Context(), wsID, status, limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	resp := make([]eventResponse, len(events))
	for i, e := range events {
		resp[i] = toEventResponse(&e)
	}
	apihttp.JSON(w, http.StatusOK, resp)
}

func (h *Handler) ListDeadLetterEvents(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	events, err := h.service.ListDeadLetterEvents(r.Context(), wsID, limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	resp := make([]eventResponse, len(events))
	for i, e := range events {
		resp[i] = toEventResponse(&e)
	}
	apihttp.JSON(w, http.StatusOK, resp)
}

func (h *Handler) RetryEvent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid event id")
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	e, err := h.service.RetryEvent(r.Context(), id, userID)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toEventResponse(e))
}

func (h *Handler) ReplayDeadLetter(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid event id")
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	e, err := h.service.ReplayDeadLetter(r.Context(), id, userID)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toEventResponse(e))
}

func (h *Handler) ReceiveIncomingWebhook(w http.ResponseWriter, r *http.Request) {
	provider := IntegrationType(chi.URLParam(r, "provider"))
	if !IsValidIntegrationType(string(provider)) {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid provider")
		return
	}
	var integrationID uuid.UUID
	if idParam := chi.URLParam(r, "integration_id"); idParam != "" {
		var err error
		integrationID, err = uuid.Parse(idParam)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid integration_id")
			return
		}
	}
	signature := r.Header.Get("X-Webhook-Signature")
	if signature == "" {
		signature = r.Header.Get("X-Hub-Signature-256")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "failed to read body")
		return
	}
	e, err := h.service.ProcessIncomingWebhook(r.Context(), provider, integrationID, signature, body)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusAccepted, toEventResponse(e))
}

// response helpers

type integrationResponse struct {
	ID           uuid.UUID         `json:"id"`
	Workspace    uuid.UUID         `json:"workspace_id"`
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	Config       map[string]string `json:"config"`
	Enabled      bool              `json:"enabled"`
	HealthStatus string            `json:"health_status"`
	LastTestedAt *string           `json:"last_tested_at,omitempty"`
	LastError    string            `json:"last_error,omitempty"`
	SyncStatus   string            `json:"sync_status"`
	RetryCount   int               `json:"retry_count"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

func toIntegrationResponse(i *Integration) integrationResponse {
	resp := integrationResponse{
		ID:           i.ID,
		Workspace:    i.WorkspaceID,
		Type:         string(i.Type),
		Name:         i.Name,
		Config:       maskSecrets(i.Config),
		Enabled:      i.Enabled,
		HealthStatus: i.HealthStatus,
		LastError:    i.LastError,
		SyncStatus:   i.SyncStatus,
		RetryCount:   i.RetryCount,
		CreatedAt:    i.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    i.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if i.LastTestedAt != nil {
		t := i.LastTestedAt.Format("2006-01-02T15:04:05Z")
		resp.LastTestedAt = &t
	}
	return resp
}

func maskSecrets(config map[string]string) map[string]string {
	if config == nil {
		return nil
	}
	masked := make(map[string]string, len(config))
	for k, v := range config {
		kl := strings.ToLower(k)
		if strings.Contains(kl, "token") || strings.Contains(kl, "secret") || strings.Contains(kl, "password") || strings.Contains(kl, "private_key") || strings.Contains(kl, "api_key") || strings.Contains(kl, "app_password") {
			masked[k] = ""
		} else {
			masked[k] = v
		}
	}
	return masked
}

type eventResponse struct {
	ID            uuid.UUID              `json:"id"`
	WorkspaceID   uuid.UUID              `json:"workspace_id,omitempty"`
	IntegrationID *uuid.UUID             `json:"integration_id,omitempty"`
	EventType     string                 `json:"event_type"`
	Payload       map[string]interface{} `json:"payload"`
	Status        string                 `json:"status"`
	ExternalID    string                 `json:"external_id"`
	RetryCount    int                    `json:"retry_count"`
	CreatedAt     string                 `json:"created_at"`
}

func toEventResponse(e *IntegrationEvent) eventResponse {
	resp := eventResponse{
		ID:            e.ID,
		WorkspaceID:   e.WorkspaceID,
		IntegrationID: e.IntegrationID,
		EventType:     e.EventType,
		Payload:       e.Payload,
		Status:        e.Status,
		ExternalID:    e.ExternalID,
		RetryCount:    e.RetryCount,
		CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if e.WorkspaceID == uuid.Nil {
		// omit workspace for incoming webhooks without a known workspace
		resp.WorkspaceID = uuid.Nil
	}
	return resp
}
