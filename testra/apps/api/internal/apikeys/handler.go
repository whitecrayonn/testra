package apikeys

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

type apiKeyResponse struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Name        string     `json:"name"`
	KeyPrefix   string     `json:"key_prefix"`
	Scopes      []string   `json:"scopes"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type createAPIKeyRequest struct {
	WorkspaceID string     `json:"workspace_id"`
	Name        string     `json:"name"`
	Scopes      []string   `json:"scopes"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type createAPIKeyResponse struct {
	APIKey apiKeyResponse `json:"api_key"`
	RawKey string         `json:"raw_key"`
}

func mapAPIKeyResponse(k APIKey) apiKeyResponse {
	return apiKeyResponse{
		ID:          k.ID.String(),
		WorkspaceID: k.WorkspaceID.String(),
		Name:        k.Name,
		KeyPrefix:   k.KeyPrefix,
		Scopes:      k.Scopes,
		ExpiresAt:   k.ExpiresAt,
		LastUsedAt:  k.LastUsedAt,
		CreatedAt:   k.CreatedAt,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	result, err := h.service.Create(r.Context(), CreateInput{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Scopes:      req.Scopes,
		ExpiresAt:   req.ExpiresAt,
		CreatedBy:   userID,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, createAPIKeyResponse{
		APIKey: mapAPIKeyResponse(result.APIKey),
		RawKey: result.RawKey,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id is required")
		return
	}

	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	params := pagination.ParseParams(r)
	keys, err := h.service.ListForWorkspacePaginated(r.Context(), workspaceID, params.Cursor, params.Limit)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]apiKeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = mapAPIKeyResponse(k)
	}

	meta := pagination.Meta{HasMore: len(keys) == params.Limit}
	if meta.HasMore && len(keys) > 0 {
		nextCursor, err := pagination.EncodeCursor(keys[len(keys)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid api key id")
		return
	}

	if err := h.service.Revoke(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{"status": "revoked"})
}

func mapError(w http.ResponseWriter, err error) {
	switch err {
	case sharederrors.ErrConflict:
		apihttp.ErrorJSON(w, http.StatusConflict, "CONFLICT", err.Error())
	case sharederrors.ErrNotFound:
		apihttp.ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case sharederrors.ErrInvalidInput:
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case sharederrors.ErrInvalidCredential:
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
	default:
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
