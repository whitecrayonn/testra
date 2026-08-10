package testgen

import (
	"encoding/json"
	"net/http"
	"time"

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

type generateFromSpecRequest struct {
	WorkspaceID  string                 `json:"workspace_id"`
	ProjectID    string                 `json:"project_id"`
	SpecFilename string                 `json:"spec_filename"`
	Spec         map[string]interface{} `json:"spec"`
}

type generationRunResponse struct {
	ID            string `json:"id"`
	WorkspaceID   string `json:"workspace_id"`
	ProjectID     string `json:"project_id"`
	Source        string `json:"source"`
	SpecFilename  string `json:"spec_filename"`
	EndpointCount int    `json:"endpoint_count"`
	CaseCount     int    `json:"case_count"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     string `json:"created_at"`
}

func mapRunResponse(run *GenerationRun) generationRunResponse {
	return generationRunResponse{
		ID:            run.ID.String(),
		WorkspaceID:   run.WorkspaceID.String(),
		ProjectID:     run.ProjectID.String(),
		Source:        run.Source,
		SpecFilename:  run.SpecFilename,
		EndpointCount: run.EndpointCount,
		CaseCount:     run.CaseCount,
		CreatedBy:     run.CreatedBy.String(),
		CreatedAt:     run.CreatedAt.Format(time.RFC3339),
	}
}

// generatedCaseResponse is a minimal projection of the created test cases —
// the full case (with all fields) is available via the normal
// GET /test-cases/{id} endpoint once generation completes.
type generatedCaseResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Priority string `json:"priority"`
	Status   string `json:"status"`
}

type generateFromSpecResponse struct {
	Run   generationRunResponse   `json:"run"`
	Cases []generatedCaseResponse `json:"cases"`
}

// GenerateFromSpec accepts an OpenAPI 3.0/3.1 document as JSON (not YAML —
// this handler has no YAML dependency by design) and creates one
// pending_review test case per rule match. See openapi_parser.go for the
// deterministic rule set; no AI/ML is used anywhere in this path.
func (h *Handler) GenerateFromSpec(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	var req generateFromSpecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}
	projID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project id")
		return
	}
	if req.Spec == nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "spec is required")
		return
	}

	result, err := h.service.GenerateFromSpec(r.Context(), GenerateFromSpecInput{
		WorkspaceID:  wsID,
		ProjectID:    projID,
		Spec:         req.Spec,
		SpecFilename: req.SpecFilename,
		CreatedBy:    userID,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	caseResp := make([]generatedCaseResponse, len(result.Cases))
	for i, c := range result.Cases {
		caseResp[i] = generatedCaseResponse{
			ID:       c.ID.String(),
			Title:    c.Title,
			Priority: string(c.Priority),
			Status:   string(c.Status),
		}
	}

	apihttp.JSON(w, http.StatusCreated, generateFromSpecResponse{
		Run:   mapRunResponse(result.Run),
		Cases: caseResp,
	})
}

type endpointFieldRequest struct {
	Name     string   `json:"name"`
	Location string   `json:"location"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Enum     []string `json:"enum"`
}

type generateFromEndpointRequest struct {
	WorkspaceID  string                 `json:"workspace_id"`
	ProjectID    string                 `json:"project_id"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	Fields       []endpointFieldRequest `json:"fields"`
	RequiresAuth bool                   `json:"requires_auth"`
}

// GenerateFromEndpoint accepts a single method/path/fields description (no
// OpenAPI document) and creates one pending_review test case per rule match,
// via the same deterministic rule set GenerateFromSpec uses — see
// Service.GenerateFromEndpoint for how the two entry points converge.
func (h *Handler) GenerateFromEndpoint(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	var req generateFromEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}
	projID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project id")
		return
	}

	fields := make([]EndpointField, len(req.Fields))
	for i, f := range req.Fields {
		fields[i] = EndpointField{
			Name:     f.Name,
			Location: f.Location,
			Type:     f.Type,
			Required: f.Required,
			Enum:     f.Enum,
		}
	}

	result, err := h.service.GenerateFromEndpoint(r.Context(), GenerateFromEndpointInput{
		WorkspaceID:  wsID,
		ProjectID:    projID,
		Method:       req.Method,
		Path:         req.Path,
		Fields:       fields,
		RequiresAuth: req.RequiresAuth,
		CreatedBy:    userID,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	caseResp := make([]generatedCaseResponse, len(result.Cases))
	for i, c := range result.Cases {
		caseResp[i] = generatedCaseResponse{
			ID:       c.ID.String(),
			Title:    c.Title,
			Priority: string(c.Priority),
			Status:   string(c.Status),
		}
	}

	apihttp.JSON(w, http.StatusCreated, generateFromSpecResponse{
		Run:   mapRunResponse(result.Run),
		Cases: caseResp,
	})
}
