package apitesting

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type fakeRepository struct {
	collections  map[uuid.UUID]*Collection
	folders      map[uuid.UUID]*Folder
	environments map[uuid.UUID]*Environment
	requests     map[uuid.UUID]*Request
	history      map[uuid.UUID]*RequestHistory
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		collections:  make(map[uuid.UUID]*Collection),
		folders:      make(map[uuid.UUID]*Folder),
		environments: make(map[uuid.UUID]*Environment),
		requests:     make(map[uuid.UUID]*Request),
		history:      make(map[uuid.UUID]*RequestHistory),
	}
}

func (f *fakeRepository) CreateCollection(_ context.Context, c *Collection) error {
	f.collections[c.ID] = c
	return nil
}

func (f *fakeRepository) GetCollectionByID(_ context.Context, id uuid.UUID) (*Collection, error) {
	if c, ok := f.collections[id]; ok {
		return c, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListCollections(_ context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Collection, error) {
	var out []Collection
	for _, c := range f.collections {
		if c.WorkspaceID == workspaceID {
			out = append(out, *c)
		}
	}
	return out, nil
}

func (f *fakeRepository) UpdateCollection(_ context.Context, c *Collection) error {
	if _, ok := f.collections[c.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.collections[c.ID] = c
	return nil
}

func (f *fakeRepository) DeleteCollection(_ context.Context, id uuid.UUID) error {
	if _, ok := f.collections[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.collections, id)
	return nil
}

func (f *fakeRepository) CreateFolder(_ context.Context, folder *Folder) error {
	f.folders[folder.ID] = folder
	return nil
}

func (f *fakeRepository) GetFolderByID(_ context.Context, id uuid.UUID) (*Folder, error) {
	if folder, ok := f.folders[id]; ok {
		return folder, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListFolders(_ context.Context, collectionID uuid.UUID, parentID *uuid.UUID, cursor string, limit int) ([]Folder, error) {
	var out []Folder
	for _, folder := range f.folders {
		if folder.CollectionID != collectionID {
			continue
		}
		if parentID != nil && folder.ParentID != nil && *folder.ParentID != *parentID {
			continue
		}
		if parentID == nil && folder.ParentID != nil {
			continue
		}
		out = append(out, *folder)
	}
	return out, nil
}

func (f *fakeRepository) UpdateFolder(_ context.Context, folder *Folder) error {
	if _, ok := f.folders[folder.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.folders[folder.ID] = folder
	return nil
}

func (f *fakeRepository) DeleteFolder(_ context.Context, id uuid.UUID) error {
	if _, ok := f.folders[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.folders, id)
	return nil
}

func (f *fakeRepository) CreateEnvironment(_ context.Context, env *Environment) error {
	f.environments[env.ID] = env
	return nil
}

func (f *fakeRepository) GetEnvironmentByID(_ context.Context, id uuid.UUID) (*Environment, error) {
	if env, ok := f.environments[id]; ok {
		return env, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListEnvironments(_ context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]Environment, error) {
	var out []Environment
	for _, env := range f.environments {
		if env.WorkspaceID == workspaceID {
			out = append(out, *env)
		}
	}
	return out, nil
}

func (f *fakeRepository) UpdateEnvironment(_ context.Context, env *Environment) error {
	if _, ok := f.environments[env.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.environments[env.ID] = env
	return nil
}

func (f *fakeRepository) DeleteEnvironment(_ context.Context, id uuid.UUID) error {
	if _, ok := f.environments[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.environments, id)
	return nil
}

func (f *fakeRepository) CreateRequest(_ context.Context, req *Request) error {
	f.requests[req.ID] = req
	return nil
}

func (f *fakeRepository) GetRequestByID(_ context.Context, id uuid.UUID) (*Request, error) {
	if req, ok := f.requests[id]; ok {
		return req, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListRequests(_ context.Context, collectionID uuid.UUID, folderID *uuid.UUID, cursor string, limit int) ([]Request, error) {
	var out []Request
	for _, req := range f.requests {
		if req.CollectionID != collectionID {
			continue
		}
		if folderID != nil && req.FolderID != nil && *req.FolderID != *folderID {
			continue
		}
		if folderID == nil && req.FolderID != nil {
			continue
		}
		out = append(out, *req)
	}
	return out, nil
}

func (f *fakeRepository) SearchRequests(_ context.Context, workspaceID uuid.UUID, query string, cursor string, limit int) ([]Request, string, error) {
	var out []Request
	q := strings.ToLower(query)
	for _, req := range f.requests {
		if req.WorkspaceID != workspaceID {
			continue
		}
		if strings.Contains(strings.ToLower(req.Name), q) || strings.Contains(strings.ToLower(req.URL), q) {
			out = append(out, *req)
		}
	}
	return out, "", nil
}

func (f *fakeRepository) UpdateRequest(_ context.Context, req *Request) error {
	if _, ok := f.requests[req.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.requests[req.ID] = req
	return nil
}

func (f *fakeRepository) DeleteRequest(_ context.Context, id uuid.UUID) error {
	if _, ok := f.requests[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.requests, id)
	return nil
}

func (f *fakeRepository) CreateRequestHistory(_ context.Context, h *RequestHistory) error {
	f.history[h.ID] = h
	return nil
}

func (f *fakeRepository) GetRequestHistoryByID(_ context.Context, id uuid.UUID) (*RequestHistory, error) {
	if h, ok := f.history[id]; ok {
		return h, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepository) ListRequestHistory(_ context.Context, requestID *uuid.UUID, workspaceID uuid.UUID, cursor string, limit int) ([]RequestHistory, error) {
	var out []RequestHistory
	for _, h := range f.history {
		if requestID != nil && (h.RequestID == nil || *h.RequestID != *requestID) {
			continue
		}
		if requestID == nil && h.WorkspaceID != workspaceID {
			continue
		}
		out = append(out, *h)
	}
	return out, nil
}

func (f *fakeRepository) RunInTx(_ context.Context, fn func(Repository) error) error {
	return fn(f)
}

func TestServiceCreateCollection(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	userID := uuid.New()

	c, err := svc.CreateCollection(context.Background(), CreateCollectionInput{
		WorkspaceID: wsID,
		Name:        "Core API",
		Description: "Core service endpoints",
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID == uuid.Nil {
		t.Fatal("expected collection id")
	}
	if c.Name != "Core API" {
		t.Errorf("expected name 'Core API', got %s", c.Name)
	}

	_, err = svc.CreateCollection(context.Background(), CreateCollectionInput{
		WorkspaceID: wsID,
		Name:        "X",
		CreatedBy:   userID,
	})
	if err != sharederrors.ErrInvalidInput {
		t.Errorf("expected invalid input, got %v", err)
	}
}

func TestServiceCreateRequest(t *testing.T) {
	svc := NewService(newFakeRepository())
	wsID := uuid.New()
	collID := uuid.New()
	userID := uuid.New()

	req, err := svc.CreateRequest(context.Background(), CreateRequestInput{
		WorkspaceID:  wsID,
		CollectionID: collID,
		Name:         "Get users",
		Method:       MethodGET,
		URL:          "https://api.example.com/users",
		CreatedBy:    userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ID == uuid.Nil {
		t.Fatal("expected request id")
	}
	if req.Method != MethodGET {
		t.Errorf("expected method GET, got %s", req.Method)
	}

	_, err = svc.CreateRequest(context.Background(), CreateRequestInput{
		WorkspaceID:  wsID,
		CollectionID: collID,
		Name:         "Bad",
		Method:       HTTPMethod("FAKE"),
		CreatedBy:    userID,
	})
	if err != sharederrors.ErrInvalidInput {
		t.Errorf("expected invalid input, got %v", err)
	}
}

func TestServiceUpdateRequest(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)
	wsID := uuid.New()
	collID := uuid.New()
	userID := uuid.New()

	req, err := svc.CreateRequest(context.Background(), CreateRequestInput{
		WorkspaceID:  wsID,
		CollectionID: collID,
		Name:         "Get users",
		Method:       MethodGET,
		URL:          "https://api.example.com/users",
		CreatedBy:    userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := svc.UpdateRequest(context.Background(), req.ID, UpdateRequestInput{
		CollectionID: collID,
		Name:         "List users",
		Method:       MethodGET,
		URL:          "https://api.example.com/users",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "List users" {
		t.Errorf("expected name 'List users', got %s", updated.Name)
	}
}

func TestServiceExecuteRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"ok":true}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	repo := newFakeRepository()
	svc := NewService(repo)
	svc.SetHTTPClient(server.Client())

	wsID := uuid.New()
	userID := uuid.New()

	result, history, err := svc.Execute(context.Background(), ExecuteInput{
		WorkspaceID: wsID,
		Request: &InlineRequest{
			Name:   "Get users",
			Method: MethodGET,
			URL:    server.URL + "/users",
			Headers: []KeyValuePair{
				{Key: "Accept", Value: "application/json", Enabled: true},
			},
		},
		Save:      true,
		CreatedBy: userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != http.StatusOK {
		t.Errorf("expected status 200, got %d", result.Status)
	}
	if history == nil {
		t.Fatal("expected history to be saved")
	}
	if history.ResponseStatus != http.StatusOK {
		t.Errorf("expected history status 200, got %d", history.ResponseStatus)
	}
}

func TestServiceVariableSubstitution(t *testing.T) {
	tests := []struct {
		template string
		vars     map[string]string
		want     string
	}{
		{"https://api.{{host}}/users", map[string]string{"host": "example.com"}, "https://api.example.com/users"},
		{"{{missing}}", map[string]string{}, "{{missing}}"},
		{"{{ a }}", map[string]string{"a": "b"}, "b"},
	}

	for _, tt := range tests {
		got := applyVariables(tt.template, tt.vars)
		if got != tt.want {
			t.Errorf("applyVariables(%q) = %q, want %q", tt.template, got, tt.want)
		}
	}
}

func TestServiceExecuteWithEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users" && r.Header.Get("Authorization") == "Bearer token123" {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"ok":true}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	repo := newFakeRepository()
	svc := NewService(repo)
	svc.SetHTTPClient(server.Client())

	wsID := uuid.New()
	userID := uuid.New()
	envID := uuid.New()
	repo.environments[envID] = &Environment{
		ID:          envID,
		WorkspaceID: wsID,
		Name:        "Prod",
		Variables: []KeyValuePair{
			{Key: "token", Value: "token123", Enabled: true},
			{Key: "baseUrl", Value: server.URL, Enabled: true},
		},
	}

	result, _, err := svc.Execute(context.Background(), ExecuteInput{
		WorkspaceID: wsID,
		Request: &InlineRequest{
			Name:          "Get users",
			Method:        MethodGET,
			URL:           "{{baseUrl}}/users",
			EnvironmentID: &envID,
			AuthType:      AuthTypeBearer,
			AuthConfig: AuthConfig{
				BearerToken: "{{token}}",
			},
		},
		Save:      true,
		CreatedBy: userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", result.Status, result.Error)
	}
}
