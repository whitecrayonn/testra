package intelligence

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) PredictFlaky(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkspaceID   string            `json:"workspace_id"`
		TestCaseID    string            `json:"test_case_id"`
		TestCaseTitle string            `json:"test_case_title"`
		History       []RunHistoryPoint `json:"history"`
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

	p, err := h.service.PredictFlaky(r.Context(), PredictFlakyInput{
		TestCaseID:    req.TestCaseID,
		TestCaseTitle: req.TestCaseTitle,
		History:       req.History,
	}, wsID)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusCreated, toPredictionResponse(p))
}

func (h *Handler) ListPredictions(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return
	}
	minScore, _ := strconv.ParseFloat(r.URL.Query().Get("min_score"), 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	predictions, err := h.service.ListPredictions(r.Context(), wsID, minScore, limit)
	if err != nil {
		mapError(w, err)
		return
	}
	resp := make([]predictionResponse, len(predictions))
	for i, p := range predictions {
		resp[i] = toPredictionResponse(&p)
	}
	apihttp.JSON(w, http.StatusOK, resp)
}

func (h *Handler) GetPrediction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}
	p, err := h.service.GetPrediction(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toPredictionResponse(p))
}

func (h *Handler) ClassifyFailure(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkspaceID  string `json:"workspace_id"`
		ErrorMessage string `json:"error_message"`
		StackTrace   string `json:"stack_trace"`
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
	result, err := h.service.ClassifyFailure(r.Context(), ClassifyFailureInput{ErrorMessage: req.ErrorMessage, StackTrace: req.StackTrace}, wsID)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, result)
}

func (h *Handler) ListClusters(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id required")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	clusters, err := h.service.ListClusters(r.Context(), wsID, limit)
	if err != nil {
		mapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, clusters)
}

type predictionResponse struct {
	ID             uuid.UUID              `json:"id"`
	WorkspaceID    uuid.UUID              `json:"workspace_id"`
	TestCaseID     *uuid.UUID             `json:"test_case_id,omitempty"`
	TestCaseTitle  string                 `json:"test_case_title"`
	FlakinessScore float64                `json:"flakiness_score"`
	Confidence     float64                `json:"confidence"`
	Features       map[string]interface{} `json:"features"`
	PredictedAt    string                 `json:"predicted_at"`
}

func toPredictionResponse(p *FlakyPrediction) predictionResponse {
	return predictionResponse{
		ID:             p.ID,
		WorkspaceID:    p.WorkspaceID,
		TestCaseID:     p.TestCaseID,
		TestCaseTitle:  p.TestCaseTitle,
		FlakinessScore: p.FlakinessScore,
		Confidence:     p.Confidence,
		Features:       p.Features,
		PredictedAt:    p.PredictedAt.Format(time.RFC3339),
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
