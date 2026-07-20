package apikeys

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type fakeRepo struct {
	keys            map[uuid.UUID]*APIKey
	hashToKey       map[string]*APIKey
	workspaceOrg    map[uuid.UUID]uuid.UUID
	updatedLastUsed map[uuid.UUID]bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		keys:            make(map[uuid.UUID]*APIKey),
		hashToKey:       make(map[string]*APIKey),
		workspaceOrg:    make(map[uuid.UUID]uuid.UUID),
		updatedLastUsed: make(map[uuid.UUID]bool),
	}
}

func (f *fakeRepo) Create(_ context.Context, key *APIKey) error {
	f.keys[key.ID] = key
	f.hashToKey[key.KeyHash] = key
	return nil
}

func (f *fakeRepo) GetByHash(_ context.Context, hash string) (*APIKey, error) {
	if key, ok := f.hashToKey[hash]; ok {
		return key, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepo) ListForWorkspace(_ context.Context, workspaceID uuid.UUID) ([]APIKey, error) {
	var out []APIKey
	for _, k := range f.keys {
		if k.WorkspaceID == workspaceID {
			out = append(out, *k)
		}
	}
	return out, nil
}

func (f *fakeRepo) ListForWorkspacePaginated(_ context.Context, workspaceID uuid.UUID, _ string, _ int) ([]APIKey, error) {
	return f.ListForWorkspace(context.Background(), workspaceID)
}

func (f *fakeRepo) Revoke(_ context.Context, id uuid.UUID) error {
	if _, ok := f.keys[id]; !ok {
		return sharederrors.ErrNotFound
	}
	now := time.Now().UTC()
	f.keys[id].RevokedAt = &now
	return nil
}

func (f *fakeRepo) UpdateLastUsed(_ context.Context, id uuid.UUID) error {
	f.updatedLastUsed[id] = true
	return nil
}

func (f *fakeRepo) GetWorkspaceOrganization(_ context.Context, workspaceID uuid.UUID) (uuid.UUID, error) {
	if orgID, ok := f.workspaceOrg[workspaceID]; ok {
		return orgID, nil
	}
	return uuid.Nil, sharederrors.ErrNotFound
}

func TestServiceCreate(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	wsID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()
	repo.workspaceOrg[wsID] = orgID

	res, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: wsID,
		Name:        "CI Key",
		Scopes:      []string{"runs:ingest"},
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RawKey == "" {
		t.Fatal("expected raw key")
	}
	if res.APIKey.OrganizationID != orgID {
		t.Fatal("expected organization to be resolved")
	}
	if res.APIKey.KeyHash != hashKey(res.RawKey) {
		t.Fatal("key hash mismatch")
	}

	_, err = svc.Create(context.Background(), CreateInput{
		WorkspaceID: wsID,
		Name:        "",
		Scopes:      []string{"runs:ingest"},
		CreatedBy:   userID,
	})
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected invalid input, got %v", err)
	}

	future := time.Now().UTC().Add(400 * 24 * time.Hour)
	_, err = svc.Create(context.Background(), CreateInput{
		WorkspaceID: wsID,
		Name:        "Too long",
		CreatedBy:   userID,
		ExpiresAt:   &future,
	})
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected invalid input for too long expiry, got %v", err)
	}
}

func TestServiceValidate(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	wsID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()
	repo.workspaceOrg[wsID] = orgID

	res, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: wsID,
		Name:        "CI Key",
		Scopes:      []string{"runs:ingest"},
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validated, err := svc.Validate(context.Background(), res.RawKey)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if validated.ID != res.APIKey.ID {
		t.Fatal("validated key mismatch")
	}
	if !repo.updatedLastUsed[validated.ID] {
		t.Fatal("expected last used to be updated")
	}

	_, err = svc.Validate(context.Background(), "testra_invalid")
	if err != sharederrors.ErrInvalidCredential {
		t.Fatalf("expected invalid credential, got %v", err)
	}

	past := time.Now().UTC().Add(-time.Hour)
	expired := APIKey{
		ID:             uuid.New(),
		WorkspaceID:    wsID,
		OrganizationID: orgID,
		KeyHash:        hashKey("testra_expired"),
		KeyPrefix:      "testra_expired"[:12],
		ExpiresAt:      &past,
		CreatedBy:      userID,
	}
	_ = repo.Create(context.Background(), &expired)
	_, err = svc.Validate(context.Background(), "testra_expired")
	if err != sharederrors.ErrInvalidCredential {
		t.Fatalf("expected expired credential, got %v", err)
	}
}

func TestServiceRevokeAndList(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	wsID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()
	repo.workspaceOrg[wsID] = orgID

	res, err := svc.Create(context.Background(), CreateInput{
		WorkspaceID: wsID,
		Name:        "Key",
		Scopes:      []string{"runs:ingest"},
		CreatedBy:   userID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Create(context.Background(), CreateInput{
		WorkspaceID: wsID,
		Name:        "Bad scope",
		Scopes:      []string{"unknown:scope"},
		CreatedBy:   userID,
	})
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected invalid input for unknown scope, got %v", err)
	}

	keys, err := svc.ListForWorkspace(context.Background(), wsID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}

	if err := svc.Revoke(context.Background(), res.APIKey.ID); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
	if repo.keys[res.APIKey.ID].RevokedAt == nil {
		t.Fatal("expected revoked_at to be set")
	}

	if err := svc.Revoke(context.Background(), uuid.New()); err != sharederrors.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
