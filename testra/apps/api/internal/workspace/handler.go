package workspace

import (
	"encoding/json"
	"net/http"

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

type workspaceResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
}

func mapWorkspaceResponse(w *Workspace) workspaceResponse {
	return workspaceResponse{
		ID:             w.ID.String(),
		OrganizationID: w.OrganizationID.String(),
		Name:           w.Name,
		Slug:           w.Slug,
	}
}

type createWorkspaceRequest struct {
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid organization id")
		return
	}

	workspace, err := h.service.Create(r.Context(), CreateInput{
		OrganizationID: orgID,
		Name:           req.Name,
		Slug:           req.Slug,
		OwnerID:        userID,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapWorkspaceResponse(workspace))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	workspace, err := h.service.Get(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapWorkspaceResponse(workspace))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("organization_id")
	if orgIDStr == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "organization_id is required")
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid organization id")
		return
	}

	params := pagination.ParseParams(r)
	workspaces, err := h.service.ListForOrganizationPaginated(r.Context(), orgID, params.Cursor, params.Limit)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]workspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		resp[i] = mapWorkspaceResponse(&ws)
	}

	meta := pagination.Meta{HasMore: len(workspaces) == params.Limit}
	if meta.HasMore && len(workspaces) > 0 {
		nextCursor, err := pagination.EncodeCursor(workspaces[len(workspaces)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

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
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
