package automationhub

import (
	"encoding/json"
	"io"
	"net/http"

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

type ingestRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ProjectID   string `json:"project_id"`
	SuiteID     string `json:"suite_id"`
	Name        string `json:"name"`
	Format      string `json:"format"`
	Payload     string `json:"payload"`
}

type ingestResponse struct {
	RunID      string `json:"run_id"`
	Total      int    `json:"total"`
	Passed     int    `json:"passed"`
	Failed     int    `json:"failed"`
	Skipped    int    `json:"skipped"`
	DurationMs int64  `json:"duration_ms"`
}

func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "failed to read request body")
		return
	}
	defer r.Body.Close()

	var meta ingestRequest
	if err := json.Unmarshal(body, &meta); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	wsID, err := uuid.Parse(meta.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	projID, err := uuid.Parse(meta.ProjectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project_id")
		return
	}

	var suiteID *uuid.UUID
	if meta.SuiteID != "" {
		sid, err := uuid.Parse(meta.SuiteID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite_id")
			return
		}
		suiteID = &sid
	}

	if !IsValidFormat(meta.Format) {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "unsupported format")
		return
	}

	if meta.Payload == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "payload is required")
		return
	}

	uid, _ := middleware.UserIDFromContext(r.Context())

	result, err := h.service.Ingest(r.Context(), IngestInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		SuiteID:     suiteID,
		Name:        meta.Name,
		Format:      IngestionFormat(meta.Format),
		Body:        []byte(meta.Payload),
		CreatedBy:   uid,
	})
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	apihttp.JSON(w, http.StatusCreated, ingestResponse{
		RunID:      result.RunID.String(),
		Total:      result.Total,
		Passed:     result.Passed,
		Failed:     result.Failed,
		Skipped:    result.Skipped,
		DurationMs: result.DurationMs,
	})
}
