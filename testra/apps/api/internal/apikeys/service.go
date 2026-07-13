package apikeys

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	WorkspaceID uuid.UUID
	Name        string
	Scopes      []string
	CreatedBy   uuid.UUID
}

type CreateResult struct {
	APIKey    APIKey
	RawKey    string
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateResult, error) {
	if input.Name == "" || input.WorkspaceID == uuid.Nil {
		return CreateResult{}, sharederrors.ErrInvalidInput
	}

	rawKey, err := generateAPIKey()
	if err != nil {
		return CreateResult{}, err
	}

	hash := hashKey(rawKey)
	prefix := rawKey[:12]

	key := APIKey{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		KeyHash:     hash,
		KeyPrefix:   prefix,
		Scopes:      input.Scopes,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, &key); err != nil {
		return CreateResult{}, err
	}

	return CreateResult{APIKey: key, RawKey: rawKey}, nil
}

func (s *Service) ListForWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]APIKey, error) {
	return s.repo.ListForWorkspace(ctx, workspaceID)
}

func (s *Service) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.repo.Revoke(ctx, id)
}

func (s *Service) Validate(ctx context.Context, rawKey string) (*APIKey, error) {
	hash := hashKey(rawKey)
	key, err := s.repo.GetByHash(ctx, hash)
	if err != nil {
		return nil, sharederrors.ErrInvalidCredential
	}

	if key.ExpiresAt != nil && time.Now().UTC().After(*key.ExpiresAt) {
		return nil, sharederrors.ErrInvalidCredential
	}

	_ = s.repo.UpdateLastUsed(ctx, key.ID)
	return key, nil
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate api key: %w", err)
	}
	return "testra_" + hex.EncodeToString(b), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
