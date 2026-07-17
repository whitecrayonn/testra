package notification

import (
	"encoding/json"
	"net/http"
	"strconv"
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

type notificationResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	Link           string    `json:"link"`
	Read           bool      `json:"read"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func mapNotificationResponse(n *Notification) notificationResponse {
	return notificationResponse{
		ID:             n.ID.String(),
		OrganizationID: n.OrganizationID.String(),
		UserID:         n.UserID.String(),
		Type:           string(n.Type),
		Title:          n.Title,
		Body:           n.Body,
		Link:           n.Link,
		Read:           n.Read,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
	}
}

type preferencesResponse struct {
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	InAppEnabled   bool      `json:"in_app_enabled"`
	EmailEnabled   bool      `json:"email_enabled"`
	SlackEnabled   bool      `json:"slack_enabled"`
	TeamsEnabled   bool      `json:"teams_enabled"`
	WebhookEnabled bool      `json:"webhook_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func mapPreferencesResponse(p *NotificationPreferences) preferencesResponse {
	return preferencesResponse{
		OrganizationID: p.OrganizationID.String(),
		UserID:         p.UserID.String(),
		InAppEnabled:   p.InAppEnabled,
		EmailEnabled:   p.EmailEnabled,
		SlackEnabled:   p.SlackEnabled,
		TeamsEnabled:   p.TeamsEnabled,
		WebhookEnabled: p.WebhookEnabled,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

type channelResponse struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organization_id"`
	WorkspaceID    string            `json:"workspace_id"`
	Type           string            `json:"type"`
	Name           string            `json:"name"`
	Config         map[string]string `json:"config"`
	CreatedBy      string            `json:"created_by"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func mapChannelResponse(ch *NotificationChannel) channelResponse {
	return channelResponse{
		ID:             ch.ID.String(),
		OrganizationID: ch.OrganizationID.String(),
		WorkspaceID:    ch.WorkspaceID.String(),
		Type:           string(ch.Type),
		Name:           ch.Name,
		Config:         ch.Config,
		CreatedBy:      ch.CreatedBy.String(),
		CreatedAt:      ch.CreatedAt,
		UpdatedAt:      ch.UpdatedAt,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	var read *bool
	if r.URL.Query().Get("read") != "" {
		b, err := strconv.ParseBool(r.URL.Query().Get("read"))
		if err == nil {
			read = &b
		}
	}

	params := pagination.ParseParams(r)
	notifications, nextCursor, err := h.service.ListNotifications(r.Context(), userID, read, params.Cursor, params.Limit)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]notificationResponse, len(notifications))
	for i, n := range notifications {
		resp[i] = mapNotificationResponse(&n)
	}

	meta := pagination.Meta{HasMore: nextCursor != "", NextCursor: nextCursor}
	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	count, err := h.service.CountUnread(r.Context(), userID)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"unread_count": count})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid notification id")
		return
	}

	var req struct {
		Read bool `json:"read"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.service.MarkRead(r.Context(), id, userID, req.Read); err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid notification id")
		return
	}

	if err := h.service.DeleteNotification(r.Context(), id, userID); err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

type createNotificationRequest struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Link   string `json:"link"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	orgID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "missing tenant context")
		return
	}

	workspaceID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	var req createNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid user id")
		return
	}

	n, err := h.service.CreateNotification(r.Context(), CreateNotificationInput{
		OrganizationID: orgID,
		WorkspaceID:    workspaceID,
		UserID:         targetUserID,
		Type:           req.Type,
		Title:          req.Title,
		Body:           req.Body,
		Link:           req.Link,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, map[string]any{"data": mapNotificationResponse(n)})
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	orgID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "missing tenant context")
		return
	}

	p, err := h.service.GetPreferences(r.Context(), orgID, userID)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapPreferencesResponse(p)})
}

type updatePreferencesRequest struct {
	InAppEnabled   bool `json:"in_app_enabled"`
	EmailEnabled   bool `json:"email_enabled"`
	SlackEnabled   bool `json:"slack_enabled"`
	TeamsEnabled   bool `json:"teams_enabled"`
	WebhookEnabled bool `json:"webhook_enabled"`
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	orgID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "missing tenant context")
		return
	}

	var req updatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	p, err := h.service.UpdatePreferences(r.Context(), UpdatePreferencesInput{
		OrganizationID: orgID,
		InAppEnabled:   req.InAppEnabled,
		EmailEnabled:   req.EmailEnabled,
		SlackEnabled:   req.SlackEnabled,
		TeamsEnabled:   req.TeamsEnabled,
		WebhookEnabled: req.WebhookEnabled,
	}, userID)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapPreferencesResponse(p)})
}

func (h *Handler) ListChannels(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id is required")
		return
	}

	channels, err := h.service.ListChannels(r.Context(), workspaceID)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]channelResponse, len(channels))
	for i, ch := range channels {
		resp[i] = mapChannelResponse(&ch)
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

type createChannelRequest struct {
	WorkspaceID string            `json:"workspace_id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Config      map[string]string `json:"config"`
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	createdBy, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	orgID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "missing tenant context")
		return
	}

	var req createChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	ch, err := h.service.CreateChannel(r.Context(), CreateChannelInput{
		OrganizationID: orgID,
		WorkspaceID:    workspaceID,
		Type:           req.Type,
		Name:           req.Name,
		Config:         req.Config,
	}, createdBy)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, map[string]any{"data": mapChannelResponse(ch)})
}

type updateChannelRequest struct {
	Type   string            `json:"type"`
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

func (h *Handler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid channel id")
		return
	}

	var req updateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	ch, err := h.service.UpdateChannel(r.Context(), id, UpdateChannelInput{
		Type:   req.Type,
		Name:   req.Name,
		Config: req.Config,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"data": mapChannelResponse(ch)})
}

func (h *Handler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid channel id")
		return
	}

	if err := h.service.DeleteChannel(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func mapError(w http.ResponseWriter, err error) {
	switch err {
	case sharederrors.ErrConflict:
		apihttp.ErrorJSON(w, http.StatusConflict, "CONFLICT", err.Error())
	case sharederrors.ErrNotFound:
		apihttp.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case sharederrors.ErrInvalidInput:
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case sharederrors.ErrUnauthorized:
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case sharederrors.ErrForbidden:
		apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	default:
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
