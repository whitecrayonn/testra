package apitesting

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

// Handler exposes the API testing HTTP endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a Handler from a Service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func userID(r *http.Request) (uuid.UUID, bool) {
	return middleware.UserIDFromContext(r.Context())
}

func parseUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

func parseUUIDString(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Collections

type createCollectionRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type collectionResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toCollectionResponse(c *Collection) collectionResponse {
	return collectionResponse{
		ID:          c.ID.String(),
		WorkspaceID: c.WorkspaceID.String(),
		Name:        c.Name,
		Description: c.Description,
		CreatedBy:   c.CreatedBy.String(),
		CreatedAt:   c.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:   c.UpdatedAt.Format(http.TimeFormat),
	}
}

func (h *Handler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	var req createCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	uid, ok := userID(r)
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	c, err := h.service.CreateCollection(r.Context(), CreateCollectionInput{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   uid,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, toCollectionResponse(c))
}

func (h *Handler) GetCollection(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	c, err := h.service.GetCollection(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toCollectionResponse(c))
}

func (h *Handler) ListCollections(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	params := pagination.ParseParams(r)
	cursor := params.Cursor
	if cursor != "" {
		decoded, err := pagination.DecodeCursor(cursor)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid cursor")
			return
		}
		cursor = decoded
	}

	collections, err := h.service.ListCollections(r.Context(), workspaceID, cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]collectionResponse, len(collections))
	for i, c := range collections {
		resp[i] = toCollectionResponse(&c)
	}

	meta := pagination.Meta{HasMore: len(collections) == params.Limit}
	if meta.HasMore && len(collections) > 0 {
		nextCursor, err := pagination.EncodeCursor(collections[len(collections)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]interface{}{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req updateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	c, err := h.service.UpdateCollection(r.Context(), id, UpdateCollectionInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toCollectionResponse(c))
}

func (h *Handler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	if err := h.service.DeleteCollection(r.Context(), id); err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusNoContent, nil)
}

// Folders

type createFolderRequest struct {
	WorkspaceID  string `json:"workspace_id"`
	CollectionID string `json:"collection_id"`
	ParentID     string `json:"parent_id"`
	Name         string `json:"name"`
}

type updateFolderRequest struct {
	ParentID string `json:"parent_id"`
	Name     string `json:"name"`
}

type folderResponse struct {
	ID           string  `json:"id"`
	WorkspaceID  string  `json:"workspace_id"`
	CollectionID string  `json:"collection_id"`
	ParentID     *string `json:"parent_id"`
	Name         string  `json:"name"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func toFolderResponse(f *Folder) folderResponse {
	resp := folderResponse{
		ID:           f.ID.String(),
		WorkspaceID:  f.WorkspaceID.String(),
		CollectionID: f.CollectionID.String(),
		Name:         f.Name,
		CreatedAt:    f.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:    f.UpdatedAt.Format(http.TimeFormat),
	}
	if f.ParentID != nil {
		s := f.ParentID.String()
		resp.ParentID = &s
	}
	return resp
}

func (h *Handler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	var req createFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}
	collectionID, err := uuid.Parse(req.CollectionID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid collection_id")
		return
	}
	parentID, err := parseUUIDString(req.ParentID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid parent_id")
		return
	}

	f, err := h.service.CreateFolder(r.Context(), CreateFolderInput{
		WorkspaceID:  workspaceID,
		CollectionID: collectionID,
		ParentID:     parentID,
		Name:         req.Name,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, toFolderResponse(f))
}

func (h *Handler) GetFolder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	f, err := h.service.GetFolder(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toFolderResponse(f))
}

func (h *Handler) ListFolders(w http.ResponseWriter, r *http.Request) {
	collectionID, err := uuid.Parse(r.URL.Query().Get("collection_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid collection_id")
		return
	}

	parentID, err := parseUUIDString(r.URL.Query().Get("parent_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid parent_id")
		return
	}

	params := pagination.ParseParams(r)
	cursor := params.Cursor
	if cursor != "" {
		decoded, err := pagination.DecodeCursor(cursor)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid cursor")
			return
		}
		cursor = decoded
	}

	folders, err := h.service.ListFolders(r.Context(), collectionID, parentID, cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]folderResponse, len(folders))
	for i, f := range folders {
		resp[i] = toFolderResponse(&f)
	}

	meta := pagination.Meta{HasMore: len(folders) == params.Limit}
	if meta.HasMore && len(folders) > 0 {
		nextCursor, err := pagination.EncodeCursor(folders[len(folders)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]interface{}{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) UpdateFolder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req updateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	parentID, err := parseUUIDString(req.ParentID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid parent_id")
		return
	}

	f, err := h.service.UpdateFolder(r.Context(), id, UpdateFolderInput{
		ParentID: parentID,
		Name:     req.Name,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toFolderResponse(f))
}

func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	if err := h.service.DeleteFolder(r.Context(), id); err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusNoContent, nil)
}

// Environments

type createEnvironmentRequest struct {
	WorkspaceID string         `json:"workspace_id"`
	Name        string         `json:"name"`
	Variables   []KeyValuePair `json:"variables"`
}

type updateEnvironmentRequest struct {
	Name      string         `json:"name"`
	Variables []KeyValuePair `json:"variables"`
}

type environmentResponse struct {
	ID          string         `json:"id"`
	WorkspaceID string         `json:"workspace_id"`
	Name        string         `json:"name"`
	Variables   []KeyValuePair `json:"variables"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

func toEnvironmentResponse(env *Environment) environmentResponse {
	return environmentResponse{
		ID:          env.ID.String(),
		WorkspaceID: env.WorkspaceID.String(),
		Name:        env.Name,
		Variables:   env.Variables,
		CreatedAt:   env.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:   env.UpdatedAt.Format(http.TimeFormat),
	}
}

func (h *Handler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var req createEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	env, err := h.service.CreateEnvironment(r.Context(), CreateEnvironmentInput{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Variables:   req.Variables,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, toEnvironmentResponse(env))
}

func (h *Handler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	env, err := h.service.GetEnvironment(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toEnvironmentResponse(env))
}

func (h *Handler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	params := pagination.ParseParams(r)
	cursor := params.Cursor
	if cursor != "" {
		decoded, err := pagination.DecodeCursor(cursor)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid cursor")
			return
		}
		cursor = decoded
	}

	environments, err := h.service.ListEnvironments(r.Context(), workspaceID, cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]environmentResponse, len(environments))
	for i, env := range environments {
		resp[i] = toEnvironmentResponse(&env)
	}

	meta := pagination.Meta{HasMore: len(environments) == params.Limit}
	if meta.HasMore && len(environments) > 0 {
		nextCursor, err := pagination.EncodeCursor(environments[len(environments)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]interface{}{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req updateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	env, err := h.service.UpdateEnvironment(r.Context(), id, UpdateEnvironmentInput{
		Name:      req.Name,
		Variables: req.Variables,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toEnvironmentResponse(env))
}

func (h *Handler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	if err := h.service.DeleteEnvironment(r.Context(), id); err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusNoContent, nil)
}

// Requests

type createRequestRequest struct {
	WorkspaceID   string         `json:"workspace_id"`
	CollectionID  string         `json:"collection_id"`
	FolderID      string         `json:"folder_id"`
	EnvironmentID string         `json:"environment_id"`
	Name          string         `json:"name"`
	Method        HTTPMethod     `json:"method"`
	URL           string         `json:"url"`
	Headers       []KeyValuePair `json:"headers"`
	QueryParams   []KeyValuePair `json:"query_params"`
	AuthType      AuthType       `json:"auth_type"`
	AuthConfig    AuthConfig     `json:"auth_config"`
	BodyType      BodyType       `json:"body_type"`
	BodyContent   string         `json:"body_content"`
	Variables     []KeyValuePair `json:"variables"`
}

type updateRequestRequest struct {
	CollectionID  string         `json:"collection_id"`
	FolderID      string         `json:"folder_id"`
	EnvironmentID string         `json:"environment_id"`
	Name          string         `json:"name"`
	Method        HTTPMethod     `json:"method"`
	URL           string         `json:"url"`
	Headers       []KeyValuePair `json:"headers"`
	QueryParams   []KeyValuePair `json:"query_params"`
	AuthType      AuthType       `json:"auth_type"`
	AuthConfig    AuthConfig     `json:"auth_config"`
	BodyType      BodyType       `json:"body_type"`
	BodyContent   string         `json:"body_content"`
	Variables     []KeyValuePair `json:"variables"`
}

type requestResponse struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspace_id"`
	CollectionID  string         `json:"collection_id"`
	FolderID      *string        `json:"folder_id"`
	EnvironmentID *string        `json:"environment_id"`
	Name          string         `json:"name"`
	Method        HTTPMethod     `json:"method"`
	URL           string         `json:"url"`
	Headers       []KeyValuePair `json:"headers"`
	QueryParams   []KeyValuePair `json:"query_params"`
	AuthType      AuthType       `json:"auth_type"`
	AuthConfig    AuthConfig     `json:"auth_config"`
	BodyType      BodyType       `json:"body_type"`
	BodyContent   string         `json:"body_content"`
	Variables     []KeyValuePair `json:"variables"`
	CreatedBy     string         `json:"created_by"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

func toRequestResponse(req *Request) requestResponse {
	resp := requestResponse{
		ID:           req.ID.String(),
		WorkspaceID:  req.WorkspaceID.String(),
		CollectionID: req.CollectionID.String(),
		Name:         req.Name,
		Method:       req.Method,
		URL:          req.URL,
		Headers:      req.Headers,
		QueryParams:  req.QueryParams,
		AuthType:     req.AuthType,
		AuthConfig:   req.AuthConfig,
		BodyType:     req.BodyType,
		BodyContent:  req.BodyContent,
		Variables:    req.Variables,
		CreatedBy:    req.CreatedBy.String(),
		CreatedAt:    req.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:    req.UpdatedAt.Format(http.TimeFormat),
	}
	if req.FolderID != nil {
		s := req.FolderID.String()
		resp.FolderID = &s
	}
	if req.EnvironmentID != nil {
		s := req.EnvironmentID.String()
		resp.EnvironmentID = &s
	}
	return resp
}

func decodeCreateRequest(req *createRequestRequest, uid uuid.UUID) (*CreateRequestInput, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}
	collectionID, err := uuid.Parse(req.CollectionID)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}
	folderID, err := parseUUIDString(req.FolderID)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}
	environmentID, err := parseUUIDString(req.EnvironmentID)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}

	return &CreateRequestInput{
		WorkspaceID:   workspaceID,
		CollectionID:  collectionID,
		FolderID:      folderID,
		EnvironmentID: environmentID,
		Name:          req.Name,
		Method:        req.Method,
		URL:           req.URL,
		Headers:       req.Headers,
		QueryParams:   req.QueryParams,
		AuthType:      req.AuthType,
		AuthConfig:    req.AuthConfig,
		BodyType:      req.BodyType,
		BodyContent:   req.BodyContent,
		Variables:     req.Variables,
		CreatedBy:     uid,
	}, nil
}

func decodeUpdateRequest(req *updateRequestRequest) (*UpdateRequestInput, error) {
	collectionID, err := uuid.Parse(req.CollectionID)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}
	folderID, err := parseUUIDString(req.FolderID)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}
	environmentID, err := parseUUIDString(req.EnvironmentID)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}

	return &UpdateRequestInput{
		CollectionID:  collectionID,
		FolderID:      folderID,
		EnvironmentID: environmentID,
		Name:          req.Name,
		Method:        req.Method,
		URL:           req.URL,
		Headers:       req.Headers,
		QueryParams:   req.QueryParams,
		AuthType:      req.AuthType,
		AuthConfig:    req.AuthConfig,
		BodyType:      req.BodyType,
		BodyContent:   req.BodyContent,
		Variables:     req.Variables,
	}, nil
}

func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var req createRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	uid, ok := userID(r)
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	input, err := decodeCreateRequest(&req, uid)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	created, err := h.service.CreateRequest(r.Context(), *input)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusCreated, toRequestResponse(created))
}

func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	req, err := h.service.GetRequest(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toRequestResponse(req))
}

func (h *Handler) ListRequests(w http.ResponseWriter, r *http.Request) {
	collectionID, err := uuid.Parse(r.URL.Query().Get("collection_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid collection_id")
		return
	}

	folderID, err := parseUUIDString(r.URL.Query().Get("folder_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid folder_id")
		return
	}

	params := pagination.ParseParams(r)
	cursor := params.Cursor
	if cursor != "" {
		decoded, err := pagination.DecodeCursor(cursor)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid cursor")
			return
		}
		cursor = decoded
	}

	requests, err := h.service.ListRequests(r.Context(), collectionID, folderID, cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]requestResponse, len(requests))
	for i, req := range requests {
		resp[i] = toRequestResponse(&req)
	}

	meta := pagination.Meta{HasMore: len(requests) == params.Limit}
	if meta.HasMore && len(requests) > 0 {
		nextCursor, err := pagination.EncodeCursor(requests[len(requests)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]interface{}{
		"data": resp,
		"meta": meta,
	})
}

func (h *Handler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	var req updateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	input, err := decodeUpdateRequest(&req)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	updated, err := h.service.UpdateRequest(r.Context(), id, *input)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toRequestResponse(updated))
}

func (h *Handler) DeleteRequest(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	if err := h.service.DeleteRequest(r.Context(), id); err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusNoContent, nil)
}

// Search requests

func (h *Handler) SearchRequests(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := uuid.Parse(r.URL.Query().Get("workspace_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	query := r.URL.Query().Get("q")
	params := pagination.ParseParams(r)

	requests, nextCursor, err := h.service.SearchRequests(r.Context(), workspaceID, query, params.Cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]requestResponse, len(requests))
	for i, req := range requests {
		resp[i] = toRequestResponse(&req)
	}

	meta := pagination.Meta{
		HasMore:    nextCursor != "",
		NextCursor: nextCursor,
	}

	apihttp.JSON(w, http.StatusOK, map[string]interface{}{
		"data": resp,
		"meta": meta,
	})
}

// Executions

type executeRequestRequest struct {
	WorkspaceID   string         `json:"workspace_id"`
	RequestID     string         `json:"request_id"`
	EnvironmentID string         `json:"environment_id"`
	Request       *InlineRequest `json:"request"`
	Save          bool           `json:"save"`
}

type executionResponse struct {
	History *requestHistoryResponse `json:"history"`
	Result  executionResultResponse `json:"result"`
}

type executionResultResponse struct {
	Status         int                 `json:"status"`
	StatusText     string              `json:"status_text"`
	Headers        map[string][]string `json:"headers"`
	Body           string              `json:"body"`
	ResponseTimeMs int                 `json:"response_time_ms"`
	Error          string              `json:"error"`
}

type requestHistoryResponse struct {
	ID                 string         `json:"id"`
	WorkspaceID        string         `json:"workspace_id"`
	RequestID          *string        `json:"request_id"`
	EnvironmentID      *string        `json:"environment_id"`
	Name               string         `json:"name"`
	Method             string         `json:"method"`
	URL                string         `json:"url"`
	RequestHeaders     []KeyValuePair `json:"request_headers"`
	RequestBody        string         `json:"request_body"`
	ResponseStatus     int            `json:"response_status"`
	ResponseStatusText string         `json:"response_status_text"`
	ResponseHeaders    []KeyValuePair `json:"response_headers"`
	ResponseBody       string         `json:"response_body"`
	ResponseTimeMs     int            `json:"response_time_ms"`
	Error              string         `json:"error"`
	CreatedBy          string         `json:"created_by"`
	CreatedAt          string         `json:"created_at"`
}

func toRequestHistoryResponse(h *RequestHistory) requestHistoryResponse {
	resp := requestHistoryResponse{
		ID:                 h.ID.String(),
		WorkspaceID:        h.WorkspaceID.String(),
		Name:               h.Name,
		Method:             h.Method,
		URL:                h.URL,
		RequestHeaders:     h.RequestHeaders,
		RequestBody:        h.RequestBody,
		ResponseStatus:     h.ResponseStatus,
		ResponseStatusText: h.ResponseStatusText,
		ResponseHeaders:    h.ResponseHeaders,
		ResponseBody:       h.ResponseBody,
		ResponseTimeMs:     h.ResponseTimeMs,
		Error:              h.Error,
		CreatedBy:          h.CreatedBy.String(),
		CreatedAt:          h.CreatedAt.Format(http.TimeFormat),
	}
	if h.RequestID != nil {
		s := h.RequestID.String()
		resp.RequestID = &s
	}
	if h.EnvironmentID != nil {
		s := h.EnvironmentID.String()
		resp.EnvironmentID = &s
	}
	return resp
}

func toExecutionResultResponse(result *ExecutionResult) executionResultResponse {
	return executionResultResponse{
		Status:         result.Status,
		StatusText:     result.StatusText,
		Headers:        result.Headers,
		Body:           string(result.Body),
		ResponseTimeMs: result.ResponseTimeMs,
		Error:          result.Error,
	}
}

func (h *Handler) ExecuteRequest(w http.ResponseWriter, r *http.Request) {
	var req executeRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
		return
	}

	var requestID *uuid.UUID
	if req.RequestID != "" {
		id, err := uuid.Parse(req.RequestID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request_id")
			return
		}
		requestID = &id
	}

	var environmentID *uuid.UUID
	if req.EnvironmentID != "" {
		id, err := uuid.Parse(req.EnvironmentID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid environment_id")
			return
		}
		environmentID = &id
	}

	uid, ok := userID(r)
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}

	result, history, err := h.service.Execute(r.Context(), ExecuteInput{
		WorkspaceID:   workspaceID,
		RequestID:     requestID,
		EnvironmentID: environmentID,
		Request:       req.Request,
		Save:          req.Save,
		CreatedBy:     uid,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := executionResponse{
		Result: toExecutionResultResponse(result),
	}
	if history != nil {
		h := toRequestHistoryResponse(history)
		resp.History = &h
	}

	apihttp.JSON(w, http.StatusOK, resp)
}

func (h *Handler) GetRequestHistory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
		return
	}

	history, err := h.service.GetRequestHistory(r.Context(), id)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	apihttp.JSON(w, http.StatusOK, toRequestHistoryResponse(history))
}

func (h *Handler) ListRequestHistory(w http.ResponseWriter, r *http.Request) {
	var requestID *uuid.UUID
	if urlID := chi.URLParam(r, "id"); urlID != "" {
		id, err := uuid.Parse(urlID)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid id")
			return
		}
		requestID = &id
	} else if rid := r.URL.Query().Get("request_id"); rid != "" {
		id, err := uuid.Parse(rid)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request_id")
			return
		}
		requestID = &id
	}

	var workspaceID uuid.UUID
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		id, err := uuid.Parse(ws)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid workspace_id")
			return
		}
		workspaceID = id
	}

	params := pagination.ParseParams(r)
	cursor := params.Cursor
	if cursor != "" {
		decoded, err := pagination.DecodeCursor(cursor)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid cursor")
			return
		}
		cursor = decoded
	}

	history, err := h.service.ListRequestHistory(r.Context(), requestID, workspaceID, cursor, params.Limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]requestHistoryResponse, len(history))
	for i, h := range history {
		resp[i] = toRequestHistoryResponse(&h)
	}

	meta := pagination.Meta{HasMore: len(history) == params.Limit}
	if meta.HasMore && len(history) > 0 {
		nextCursor, err := pagination.EncodeCursor(history[len(history)-1].ID.String())
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSON(w, http.StatusOK, map[string]interface{}{
		"data": resp,
		"meta": meta,
	})
}
