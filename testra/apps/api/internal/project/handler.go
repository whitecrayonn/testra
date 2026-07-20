package project

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/pagination"
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
		apihttp.MapError(w, err)
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
		apihttp.MapError(w, err)
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

	params := pagination.ParseParams(r)
	projects, err := h.service.ListForWorkspacePaginated(r.Context(), workspaceID, params.Cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]projectResponse, len(projects))
	for i, p := range projects {
		resp[i] = mapProjectResponse(&p)
	}

	meta := pagination.Meta{HasMore: len(projects) == params.Limit}
	if meta.HasMore && len(projects) > 0 {
		nextCursor, err := pagination.EncodeCursor(projects[len(projects)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}
