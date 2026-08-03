package search

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id is required")
		return
	}

	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "q is required")
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := h.service.Search(r.Context(), wsID, query, limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, result)
}
