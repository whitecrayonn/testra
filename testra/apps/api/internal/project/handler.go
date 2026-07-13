package project

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type projectResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
}

func mapProjectResponse(p *Project) projectResponse {
	return projectResponse{
		ID:          p.ID.String(),
		WorkspaceID: p.WorkspaceID.String(),
		Name:        p.Name,
		Key:         p.Key,
		Description: p.Description,
	}
}

type createProjectRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	project, err := h.service.Create(r.Context(), CreateInput{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Key:         req.Key,
		Description: req.Description,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapProjectResponse(project))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project id")
		return
	}

	project, err := h.service.Get(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapProjectResponse(project))
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

	projects, err := h.service.ListForWorkspace(r.Context(), workspaceID)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]projectResponse, len(projects))
	for i, p := range projects {
		resp[i] = mapProjectResponse(&p)
	}
	apihttp.JSON(w, http.StatusOK, resp)
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
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
