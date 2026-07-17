package testmanagement

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

type folderResponse struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	ParentID    *string `json:"parent_id"`
	Name        string  `json:"name"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func mapFolderResponse(f *TestFolder) folderResponse {
	var parentID *string
	if f.ParentID != nil {
		s := f.ParentID.String()
		parentID = &s
	}
	return folderResponse{
		ID:          f.ID.String(),
		WorkspaceID: f.WorkspaceID.String(),
		ParentID:    parentID,
		Name:        f.Name,
		CreatedAt:   f.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   f.UpdatedAt.Format(time.RFC3339),
	}
}

type suiteResponse struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	FolderID    *string `json:"folder_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func mapSuiteResponse(s *TestSuite) suiteResponse {
	var folderID *string
	if s.FolderID != nil {
		f := s.FolderID.String()
		folderID = &f
	}
	return suiteResponse{
		ID:          s.ID.String(),
		WorkspaceID: s.WorkspaceID.String(),
		FolderID:    folderID,
		Name:        s.Name,
		Description: s.Description,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}

type stepResponse struct {
	Order    int    `json:"order"`
	Action   string `json:"action"`
	Expected string `json:"expected"`
	TestData string `json:"test_data"`
}

type caseResponse struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspace_id"`
	ProjectID     string         `json:"project_id"`
	SuiteID       *string        `json:"suite_id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Preconditions string         `json:"preconditions"`
	Steps         []stepResponse `json:"steps"`
	Status        string         `json:"status"`
	Priority      string         `json:"priority"`
	Tags          []string       `json:"tags"`
	Version       int            `json:"version"`
	CreatedBy     string         `json:"created_by"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

func mapCaseResponse(tc *TestCase) caseResponse {
	var suiteID *string
	if tc.SuiteID != nil {
		s := tc.SuiteID.String()
		suiteID = &s
	}
	steps := make([]stepResponse, len(tc.Steps))
	for i, st := range tc.Steps {
		steps[i] = stepResponse{
			Order:    st.Order,
			Action:   st.Action,
			Expected: st.Expected,
			TestData: st.TestData,
		}
	}
	if tc.Tags == nil {
		tc.Tags = []string{}
	}
	return caseResponse{
		ID:            tc.ID.String(),
		WorkspaceID:   tc.WorkspaceID.String(),
		ProjectID:     tc.ProjectID.String(),
		SuiteID:       suiteID,
		Title:         tc.Title,
		Description:   tc.Description,
		Preconditions: tc.Preconditions,
		Steps:         steps,
		Status:        string(tc.Status),
		Priority:      string(tc.Priority),
		Tags:          tc.Tags,
		Version:       tc.Version,
		CreatedBy:     tc.CreatedBy.String(),
		CreatedAt:     tc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     tc.UpdatedAt.Format(time.RFC3339),
	}
}

type versionResponse struct {
	ID            string         `json:"id"`
	TestCaseID    string         `json:"test_case_id"`
	Version       int            `json:"version"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Preconditions string         `json:"preconditions"`
	Steps         []stepResponse `json:"steps"`
	ChangedBy     string         `json:"changed_by"`
	CreatedAt     string         `json:"created_at"`
}

func mapVersionResponse(v *TestCaseVersion) versionResponse {
	steps := make([]stepResponse, len(v.Steps))
	for i, st := range v.Steps {
		steps[i] = stepResponse{
			Order:    st.Order,
			Action:   st.Action,
			Expected: st.Expected,
			TestData: st.TestData,
		}
	}
	return versionResponse{
		ID:            v.ID.String(),
		TestCaseID:    v.TestCaseID.String(),
		Version:       v.Version,
		Title:         v.Title,
		Description:   v.Description,
		Preconditions: v.Preconditions,
		Steps:         steps,
		ChangedBy:     v.ChangedBy.String(),
		CreatedAt:     v.CreatedAt.Format(time.RFC3339),
	}
}

type createFolderRequest struct {
	WorkspaceID string  `json:"workspace_id"`
	ParentID    *string `json:"parent_id"`
	Name        string  `json:"name"`
}

func (h *Handler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	var req createFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid parent id")
			return
		}
		parentID = &pid
	}

	folder, err := h.service.CreateFolder(r.Context(), CreateFolderInput{
		WorkspaceID: wsID,
		ParentID:    parentID,
		Name:        req.Name,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapFolderResponse(folder))
}

func (h *Handler) GetFolder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid folder id")
		return
	}

	folder, err := h.service.GetFolder(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapFolderResponse(folder))
}

func (h *Handler) ListFolders(w http.ResponseWriter, r *http.Request) {
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

	var parentID *uuid.UUID
	if pidStr := r.URL.Query().Get("parent_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid parent id")
			return
		}
		parentID = &pid
	}

	folders, err := h.service.ListFolders(r.Context(), wsID, parentID)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]folderResponse, len(folders))
	for i, f := range folders {
		resp[i] = mapFolderResponse(&f)
	}
	apihttp.JSON(w, http.StatusOK, resp)
}

type updateFolderRequest struct {
	Name string `json:"name"`
}

func (h *Handler) UpdateFolder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid folder id")
		return
	}

	var req updateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	folder, err := h.service.UpdateFolder(r.Context(), id, UpdateFolderInput{Name: req.Name})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapFolderResponse(folder))
}

func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid folder id")
		return
	}

	if err := h.service.DeleteFolder(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type createSuiteRequest struct {
	WorkspaceID string  `json:"workspace_id"`
	FolderID    *string `json:"folder_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
}

func (h *Handler) CreateSuite(w http.ResponseWriter, r *http.Request) {
	var req createSuiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	wsID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace id")
		return
	}

	var folderID *uuid.UUID
	if req.FolderID != nil && *req.FolderID != "" {
		fid, err := uuid.Parse(*req.FolderID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid folder id")
			return
		}
		folderID = &fid
	}

	suite, err := h.service.CreateSuite(r.Context(), CreateSuiteInput{
		WorkspaceID: wsID,
		FolderID:    folderID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapSuiteResponse(suite))
}

func (h *Handler) GetSuite(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite id")
		return
	}

	suite, err := h.service.GetSuite(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapSuiteResponse(suite))
}

func (h *Handler) ListSuites(w http.ResponseWriter, r *http.Request) {
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

	var folderID *uuid.UUID
	if fidStr := r.URL.Query().Get("folder_id"); fidStr != "" {
		fid, err := uuid.Parse(fidStr)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid folder id")
			return
		}
		folderID = &fid
	}

	suites, err := h.service.ListSuites(r.Context(), wsID, folderID)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]suiteResponse, len(suites))
	for i, s := range suites {
		resp[i] = mapSuiteResponse(&s)
	}
	apihttp.JSON(w, http.StatusOK, resp)
}

type updateSuiteRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) UpdateSuite(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite id")
		return
	}

	var req updateSuiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	suite, err := h.service.UpdateSuite(r.Context(), id, UpdateSuiteInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapSuiteResponse(suite))
}

func (h *Handler) DeleteSuite(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite id")
		return
	}

	if err := h.service.DeleteSuite(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type createStepRequest struct {
	Action   string `json:"action"`
	Expected string `json:"expected"`
	TestData string `json:"test_data"`
}

type createCaseRequest struct {
	WorkspaceID   string              `json:"workspace_id"`
	ProjectID     string              `json:"project_id"`
	SuiteID       *string             `json:"suite_id"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	Preconditions string              `json:"preconditions"`
	Steps         []createStepRequest `json:"steps"`
	Status        string              `json:"status"`
	Priority      string              `json:"priority"`
	Tags          []string            `json:"tags"`
}

func (h *Handler) CreateCase(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	var req createCaseRequest
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

	var suiteID *uuid.UUID
	if req.SuiteID != nil && *req.SuiteID != "" {
		sid, err := uuid.Parse(*req.SuiteID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite id")
			return
		}
		suiteID = &sid
	}

	steps := make([]TestStep, len(req.Steps))
	for i, s := range req.Steps {
		steps[i] = TestStep{
			Order:    i + 1,
			Action:   s.Action,
			Expected: s.Expected,
			TestData: s.TestData,
		}
	}

	tc, err := h.service.CreateCase(r.Context(), CreateCaseInput{
		WorkspaceID:   wsID,
		ProjectID:     projID,
		SuiteID:       suiteID,
		Title:         req.Title,
		Description:   req.Description,
		Preconditions: req.Preconditions,
		Steps:         steps,
		Status:        TestCaseStatus(req.Status),
		Priority:      TestCasePriority(req.Priority),
		Tags:          req.Tags,
		CreatedBy:     userID,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, mapCaseResponse(tc))
}

func (h *Handler) GetCase(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test case id")
		return
	}

	tc, err := h.service.GetCase(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapCaseResponse(tc))
}

func (h *Handler) ListCases(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr == "" {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "project_id is required")
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid project id")
		return
	}

	var suiteID *uuid.UUID
	if sidStr := r.URL.Query().Get("suite_id"); sidStr != "" {
		sid, err := uuid.Parse(sidStr)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite id")
			return
		}
		suiteID = &sid
	}

	params := pagination.ParseParams(r)
	cases, err := h.service.ListCases(r.Context(), projectID, suiteID, params.Cursor, params.Limit)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]caseResponse, len(cases))
	for i, tc := range cases {
		resp[i] = mapCaseResponse(&tc)
	}

	meta := pagination.Meta{HasMore: len(cases) == params.Limit}
	if meta.HasMore && len(cases) > 0 {
		nextCursor, err := pagination.EncodeCursor(cases[len(cases)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) SearchCases(w http.ResponseWriter, r *http.Request) {
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

	params := pagination.ParseParams(r)
	cases, nextCursor, err := h.service.SearchCases(r.Context(), wsID, query, params.Cursor, params.Limit)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]caseResponse, len(cases))
	for i, tc := range cases {
		resp[i] = mapCaseResponse(&tc)
	}

	meta := pagination.Meta{HasMore: nextCursor != ""}
	if meta.HasMore {
		meta.NextCursor = nextCursor
	}

	apihttp.JSON(w, http.StatusOK, map[string]any{
		"data": resp,
		"meta": meta,
	})
}

type updateCaseRequest struct {
	SuiteID       *string             `json:"suite_id"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	Preconditions string              `json:"preconditions"`
	Steps         []createStepRequest `json:"steps"`
	Status        string              `json:"status"`
	Priority      string              `json:"priority"`
	Tags          []string            `json:"tags"`
}

func (h *Handler) UpdateCase(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test case id")
		return
	}

	var req updateCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	var suiteID *uuid.UUID
	if req.SuiteID != nil && *req.SuiteID != "" {
		sid, err := uuid.Parse(*req.SuiteID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid suite id")
			return
		}
		suiteID = &sid
	}

	steps := make([]TestStep, len(req.Steps))
	for i, s := range req.Steps {
		steps[i] = TestStep{
			Order:    i + 1,
			Action:   s.Action,
			Expected: s.Expected,
			TestData: s.TestData,
		}
	}

	tc, err := h.service.UpdateCase(r.Context(), id, UpdateCaseInput{
		SuiteID:       suiteID,
		Title:         req.Title,
		Description:   req.Description,
		Preconditions: req.Preconditions,
		Steps:         steps,
		Status:        TestCaseStatus(req.Status),
		Priority:      TestCasePriority(req.Priority),
		Tags:          req.Tags,
		ChangedBy:     userID,
	})
	if err != nil {
		mapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, mapCaseResponse(tc))
}

func (h *Handler) DeleteCase(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test case id")
		return
	}

	if err := h.service.DeleteCase(r.Context(), id); err != nil {
		mapError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid test case id")
		return
	}

	versions, err := h.service.ListVersions(r.Context(), id)
	if err != nil {
		mapError(w, err)
		return
	}

	resp := make([]versionResponse, len(versions))
	for i, v := range versions {
		resp[i] = mapVersionResponse(&v)
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
		apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
