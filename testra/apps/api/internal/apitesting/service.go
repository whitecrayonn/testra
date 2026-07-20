package apitesting

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

// Service implements the API testing business logic.
type Service struct {
	repo       Repository
	httpClient *http.Client
}

// NewService creates a new service with the default HTTP client.
func NewService(repo Repository) *Service {
	return &Service{
		repo:       repo,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetHTTPClient allows tests to inject a custom HTTP client.
func (s *Service) SetHTTPClient(client *http.Client) {
	s.httpClient = client
}

// Collection inputs

type CreateCollectionInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Description string
	CreatedBy   uuid.UUID
}

type UpdateCollectionInput struct {
	Name        string
	Description string
}

// Folder inputs

type CreateFolderInput struct {
	WorkspaceID  uuid.UUID
	CollectionID uuid.UUID
	ParentID     *uuid.UUID
	Name         string
}

type UpdateFolderInput struct {
	ParentID *uuid.UUID
	Name     string
}

// Environment inputs

type CreateEnvironmentInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Variables   []KeyValuePair
}

type UpdateEnvironmentInput struct {
	Name      string
	Variables []KeyValuePair
}

// Request inputs

type CreateRequestInput struct {
	WorkspaceID   uuid.UUID
	CollectionID  uuid.UUID
	FolderID      *uuid.UUID
	EnvironmentID *uuid.UUID
	Name          string
	Method        HTTPMethod
	URL           string
	Headers       []KeyValuePair
	QueryParams   []KeyValuePair
	AuthType      AuthType
	AuthConfig    AuthConfig
	BodyType      BodyType
	BodyContent   string
	Variables     []KeyValuePair
	CreatedBy     uuid.UUID
}

type UpdateRequestInput struct {
	CollectionID  uuid.UUID
	FolderID      *uuid.UUID
	EnvironmentID *uuid.UUID
	Name          string
	Method        HTTPMethod
	URL           string
	Headers       []KeyValuePair
	QueryParams   []KeyValuePair
	AuthType      AuthType
	AuthConfig    AuthConfig
	BodyType      BodyType
	BodyContent   string
	Variables     []KeyValuePair
}

// Execution inputs

type InlineRequest struct {
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
	EnvironmentID *uuid.UUID     `json:"environment_id"`
}

type ExecuteInput struct {
	WorkspaceID   uuid.UUID
	RequestID     *uuid.UUID
	EnvironmentID *uuid.UUID
	Request       *InlineRequest
	Save          bool
	CreatedBy     uuid.UUID
}

// Collection CRUD

func (s *Service) CreateCollection(ctx context.Context, input CreateCollectionInput) (*Collection, error) {
	if input.WorkspaceID == uuid.Nil || !IsValidName(input.Name) || input.CreatedBy == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	now := time.Now().UTC()
	c := &Collection{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		CreatedBy:   input.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateCollection(ctx, c); err != nil {
		return nil, err
	}

	eventbus.Default().Publish(ctx, eventbus.Event{
		Type:     "api_test.collection_created",
		TenantID: c.WorkspaceID.String(),
		Payload: map[string]interface{}{
			"collection_id": c.ID.String(),
			"workspace_id":  c.WorkspaceID.String(),
			"name":          c.Name,
		},
	})

	return c, nil
}

func (s *Service) GetCollection(ctx context.Context, id uuid.UUID) (*Collection, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetCollectionByID(ctx, id)
}

func (s *Service) ListCollections(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Collection, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListCollections(ctx, workspaceID, cursor, limit)
}

func (s *Service) UpdateCollection(ctx context.Context, id uuid.UUID, input UpdateCollectionInput) (*Collection, error) {
	if id == uuid.Nil || !IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	c, err := s.repo.GetCollectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.Name = strings.TrimSpace(input.Name)
	c.Description = strings.TrimSpace(input.Description)
	c.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateCollection(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteCollection(ctx, id)
}

// Folder CRUD

func (s *Service) CreateFolder(ctx context.Context, input CreateFolderInput) (*Folder, error) {
	if input.WorkspaceID == uuid.Nil || input.CollectionID == uuid.Nil || !IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	now := time.Now().UTC()
	f := &Folder{
		ID:           uuid.New(),
		WorkspaceID:  input.WorkspaceID,
		CollectionID: input.CollectionID,
		ParentID:     input.ParentID,
		Name:         strings.TrimSpace(input.Name),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateFolder(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) GetFolder(ctx context.Context, id uuid.UUID) (*Folder, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetFolderByID(ctx, id)
}

func (s *Service) ListFolders(ctx context.Context, collectionID uuid.UUID, parentID *uuid.UUID, cursor string, limit int) ([]Folder, error) {
	if collectionID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListFolders(ctx, collectionID, parentID, cursor, limit)
}

func (s *Service) UpdateFolder(ctx context.Context, id uuid.UUID, input UpdateFolderInput) (*Folder, error) {
	if id == uuid.Nil || !IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	f, err := s.repo.GetFolderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	f.Name = strings.TrimSpace(input.Name)
	f.ParentID = input.ParentID
	f.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateFolder(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteFolder(ctx, id)
}

// Environment CRUD

func (s *Service) CreateEnvironment(ctx context.Context, input CreateEnvironmentInput) (*Environment, error) {
	if input.WorkspaceID == uuid.Nil || !IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	now := time.Now().UTC()
	env := &Environment{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		Name:        strings.TrimSpace(input.Name),
		Variables:   normalizeKeyValuePairs(input.Variables),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateEnvironment(ctx, env); err != nil {
		return nil, err
	}
	return env, nil
}

func (s *Service) GetEnvironment(ctx context.Context, id uuid.UUID) (*Environment, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetEnvironmentByID(ctx, id)
}

func (s *Service) ListEnvironments(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Environment, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListEnvironments(ctx, workspaceID, cursor, limit)
}

func (s *Service) UpdateEnvironment(ctx context.Context, id uuid.UUID, input UpdateEnvironmentInput) (*Environment, error) {
	if id == uuid.Nil || !IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}

	env, err := s.repo.GetEnvironmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	env.Name = strings.TrimSpace(input.Name)
	env.Variables = normalizeKeyValuePairs(input.Variables)
	env.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateEnvironment(ctx, env); err != nil {
		return nil, err
	}
	return env, nil
}

func (s *Service) DeleteEnvironment(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteEnvironment(ctx, id)
}

// Request CRUD

func (s *Service) CreateRequest(ctx context.Context, input CreateRequestInput) (*Request, error) {
	if input.WorkspaceID == uuid.Nil || input.CollectionID == uuid.Nil || !IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}
	if !IsValidHTTPMethod(input.Method) {
		return nil, sharederrors.ErrInvalidInput
	}

	now := time.Now().UTC()
	req := &Request{
		ID:            uuid.New(),
		WorkspaceID:   input.WorkspaceID,
		CollectionID:  input.CollectionID,
		FolderID:      input.FolderID,
		EnvironmentID: input.EnvironmentID,
		Name:          strings.TrimSpace(input.Name),
		Method:        input.Method,
		URL:           strings.TrimSpace(input.URL),
		Headers:       normalizeKeyValuePairs(input.Headers),
		QueryParams:   normalizeKeyValuePairs(input.QueryParams),
		AuthType:      input.AuthType,
		AuthConfig:    input.AuthConfig,
		BodyType:      input.BodyType,
		BodyContent:   input.BodyContent,
		Variables:     normalizeKeyValuePairs(input.Variables),
		CreatedBy:     input.CreatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if req.AuthType == "" {
		req.AuthType = AuthTypeNone
	}
	if req.BodyType == "" {
		req.BodyType = BodyTypeNone
	}
	if !IsValidAuthType(req.AuthType) || !IsValidBodyType(req.BodyType) {
		return nil, sharederrors.ErrInvalidInput
	}

	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) GetRequest(ctx context.Context, id uuid.UUID) (*Request, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetRequestByID(ctx, id)
}

func (s *Service) ListRequests(ctx context.Context, collectionID uuid.UUID, folderID *uuid.UUID, cursor string, limit int) ([]Request, error) {
	if collectionID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListRequests(ctx, collectionID, folderID, cursor, limit)
}

func (s *Service) SearchRequests(ctx context.Context, workspaceID uuid.UUID, query string, cursor string, limit int) ([]Request, string, error) {
	if workspaceID == uuid.Nil {
		return nil, "", sharederrors.ErrInvalidInput
	}
	return s.repo.SearchRequests(ctx, workspaceID, query, cursor, limit)
}

func (s *Service) UpdateRequest(ctx context.Context, id uuid.UUID, input UpdateRequestInput) (*Request, error) {
	if id == uuid.Nil || !IsValidName(input.Name) || !IsValidHTTPMethod(input.Method) {
		return nil, sharederrors.ErrInvalidInput
	}

	req, err := s.repo.GetRequestByID(ctx, id)
	if err != nil {
		return nil, err
	}

	req.CollectionID = input.CollectionID
	req.FolderID = input.FolderID
	req.EnvironmentID = input.EnvironmentID
	req.Name = strings.TrimSpace(input.Name)
	req.Method = input.Method
	req.URL = strings.TrimSpace(input.URL)
	req.Headers = normalizeKeyValuePairs(input.Headers)
	req.QueryParams = normalizeKeyValuePairs(input.QueryParams)
	req.AuthType = input.AuthType
	req.AuthConfig = input.AuthConfig
	req.BodyType = input.BodyType
	req.BodyContent = input.BodyContent
	req.Variables = normalizeKeyValuePairs(input.Variables)
	req.UpdatedAt = time.Now().UTC()

	if req.AuthType == "" {
		req.AuthType = AuthTypeNone
	}
	if req.BodyType == "" {
		req.BodyType = BodyTypeNone
	}
	if !IsValidAuthType(req.AuthType) || !IsValidBodyType(req.BodyType) {
		return nil, sharederrors.ErrInvalidInput
	}

	if err := s.repo.UpdateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteRequest(ctx, id)
}

// Execution

func (s *Service) Execute(ctx context.Context, input ExecuteInput) (*ExecutionResult, *RequestHistory, error) {
	if input.WorkspaceID == uuid.Nil {
		return nil, nil, sharederrors.ErrInvalidInput
	}

	var req Request
	var requestID *uuid.UUID
	if input.RequestID != nil {
		stored, err := s.repo.GetRequestByID(ctx, *input.RequestID)
		if err != nil {
			return nil, nil, err
		}
		req = *stored
		requestID = &req.ID
	} else if input.Request != nil {
		inline := input.Request
		if !IsValidHTTPMethod(inline.Method) || strings.TrimSpace(inline.URL) == "" {
			return nil, nil, sharederrors.ErrInvalidInput
		}
		req = Request{
			WorkspaceID: input.WorkspaceID,
			Name:        inline.Name,
			Method:      inline.Method,
			URL:         strings.TrimSpace(inline.URL),
			Headers:     normalizeKeyValuePairs(inline.Headers),
			QueryParams: normalizeKeyValuePairs(inline.QueryParams),
			AuthType:    inline.AuthType,
			AuthConfig:  inline.AuthConfig,
			BodyType:    inline.BodyType,
			BodyContent: inline.BodyContent,
			Variables:   normalizeKeyValuePairs(inline.Variables),
		}
		if req.AuthType == "" {
			req.AuthType = AuthTypeNone
		}
		if req.BodyType == "" {
			req.BodyType = BodyTypeNone
		}
		if !IsValidAuthType(req.AuthType) || !IsValidBodyType(req.BodyType) {
			return nil, nil, sharederrors.ErrInvalidInput
		}
		if inline.EnvironmentID != nil {
			req.EnvironmentID = inline.EnvironmentID
		}
	} else {
		return nil, nil, sharederrors.ErrInvalidInput
	}

	environmentID := req.EnvironmentID
	if input.EnvironmentID != nil {
		environmentID = input.EnvironmentID
	}

	vars, err := s.buildVariables(ctx, input.WorkspaceID, environmentID, req.Variables)
	if err != nil {
		return nil, nil, err
	}

	result, err := s.runHTTPRequest(ctx, req, vars)
	if err != nil {
		return nil, nil, err
	}

	history := &RequestHistory{
		ID:                 uuid.New(),
		WorkspaceID:        input.WorkspaceID,
		RequestID:          requestID,
		EnvironmentID:      environmentID,
		Name:               req.Name,
		Method:             string(req.Method),
		URL:                req.URL,
		RequestHeaders:     req.Headers,
		RequestBody:        req.BodyContent,
		ResponseStatus:     result.Status,
		ResponseStatusText: result.StatusText,
		ResponseHeaders:    keyValuePairsFromHeaders(result.Headers),
		ResponseBody:       string(result.Body),
		ResponseTimeMs:     result.ResponseTimeMs,
		Error:              result.Error,
		CreatedBy:          input.CreatedBy,
		CreatedAt:          time.Now().UTC(),
	}

	if input.Save {
		if err := s.repo.CreateRequestHistory(ctx, history); err != nil {
			return result, nil, err
		}
	}

	status := "passed"
	if result == nil || result.Status < 200 || result.Status >= 300 || result.Error != "" {
		status = "failed"
	}
	var statusCode, duration int
	if result != nil {
		statusCode = result.Status
		duration = result.ResponseTimeMs
	}
	eventbus.Default().Publish(ctx, eventbus.Event{
		Type:     "api_test.executed",
		TenantID: input.WorkspaceID.String(),
		Payload: map[string]interface{}{
			"request_id":   requestID,
			"workspace_id": input.WorkspaceID.String(),
			"status":       status,
			"status_code":  statusCode,
			"duration_ms":  duration,
		},
	})

	return result, history, nil
}

func (s *Service) GetRequestHistory(ctx context.Context, id uuid.UUID) (*RequestHistory, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetRequestHistoryByID(ctx, id)
}

func (s *Service) ListRequestHistory(ctx context.Context, requestID *uuid.UUID, workspaceID uuid.UUID, cursor string, limit int) ([]RequestHistory, error) {
	if requestID == nil && workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}

	resolvedWorkspace := workspaceID
	if requestID != nil && workspaceID == uuid.Nil {
		req, err := s.repo.GetRequestByID(ctx, *requestID)
		if err != nil {
			return nil, err
		}
		resolvedWorkspace = req.WorkspaceID
	}

	return s.repo.ListRequestHistory(ctx, requestID, resolvedWorkspace, cursor, limit)
}

func (s *Service) buildVariables(ctx context.Context, workspaceID uuid.UUID, environmentID *uuid.UUID, requestVars []KeyValuePair) (map[string]string, error) {
	vars := make(map[string]string)

	if environmentID != nil {
		env, err := s.repo.GetEnvironmentByID(ctx, *environmentID)
		if err != nil {
			return nil, err
		}
		if env.WorkspaceID != workspaceID {
			return nil, sharederrors.ErrInvalidInput
		}
		for _, v := range env.Variables {
			if v.Enabled && v.Key != "" {
				vars[v.Key] = v.Value
			}
		}
	}

	for _, v := range requestVars {
		if v.Enabled && v.Key != "" {
			vars[v.Key] = v.Value
		}
	}

	return vars, nil
}

var variableRegex = regexp.MustCompile(`\{\{([^{}]+)\}\}`)

func applyVariables(template string, vars map[string]string) string {
	return variableRegex.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.TrimSpace(match[2 : len(match)-2])
		if val, ok := vars[key]; ok {
			return val
		}
		return match
	})
}

func (s *Service) runHTTPRequest(ctx context.Context, req Request, vars map[string]string) (*ExecutionResult, error) {
	rawURL := applyVariables(req.URL, vars)

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}

	q := u.Query()
	for _, p := range req.QueryParams {
		if !p.Enabled {
			continue
		}
		key := applyVariables(p.Key, vars)
		value := applyVariables(p.Value, vars)
		if key == "" {
			continue
		}
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	var bodyReader io.Reader
	var bodyContent string
	if req.BodyType != BodyTypeNone && req.BodyContent != "" {
		bodyContent = applyVariables(req.BodyContent, vars)
		bodyReader = strings.NewReader(bodyContent)
	}

	httpReq, err := http.NewRequestWithContext(ctx, string(req.Method), u.String(), bodyReader)
	if err != nil {
		return nil, sharederrors.ErrInvalidInput
	}

	for _, h := range req.Headers {
		if !h.Enabled {
			continue
		}
		key := applyVariables(h.Key, vars)
		value := applyVariables(h.Value, vars)
		if key == "" {
			continue
		}
		httpReq.Header.Add(key, value)
	}

	switch req.AuthType {
	case AuthTypeBearer:
		token := applyVariables(req.AuthConfig.BearerToken, vars)
		if token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+token)
		}
	case AuthTypeBasic:
		username := applyVariables(req.AuthConfig.Username, vars)
		password := applyVariables(req.AuthConfig.Password, vars)
		if username != "" {
			httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
		}
	case AuthTypeAPIKey:
		key := applyVariables(req.AuthConfig.APIKey, vars)
		value := applyVariables(req.AuthConfig.APIValue, vars)
		location := req.AuthConfig.APILocation
		if key != "" && value != "" {
			if location == "query" {
				q := httpReq.URL.Query()
				q.Set(key, value)
				httpReq.URL.RawQuery = q.Encode()
			} else {
				httpReq.Header.Set(key, value)
			}
		}
	}

	if req.BodyType == BodyTypeJSON && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if req.BodyType == BodyTypeURLEncoded && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	start := time.Now()
	httpResp, err := s.httpClient.Do(httpReq)
	elapsed := time.Since(start)

	result := &ExecutionResult{
		Headers:        make(map[string][]string),
		ResponseTimeMs: int(elapsed.Milliseconds()),
	}

	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer httpResp.Body.Close()

	result.Status = httpResp.StatusCode
	result.StatusText = httpResp.Status
	result.Headers = httpResp.Header

	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 2*1024*1024))
	if err != nil {
		result.Error = "failed to read response body: " + err.Error()
		return result, nil
	}
	result.Body = body

	return result, nil
}

func keyValuePairsFromHeaders(headers map[string][]string) []KeyValuePair {
	var pairs []KeyValuePair
	for key, values := range headers {
		for _, value := range values {
			pairs = append(pairs, KeyValuePair{Key: key, Value: value, Enabled: true})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key < pairs[j].Key
	})
	return pairs
}

func normalizeKeyValuePairs(pairs []KeyValuePair) []KeyValuePair {
	if pairs == nil {
		return []KeyValuePair{}
	}
	return pairs
}

// IsValidName is exposed for handler use; it delegates to the domain helper.
func (s *Service) IsValidName(name string) bool {
	return validation.IsValidName(name) && len(strings.TrimSpace(name)) >= 2
}
