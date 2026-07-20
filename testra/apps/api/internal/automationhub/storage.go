package automationhub

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ArtifactStorage persists execution artifacts on the local filesystem.
// It is intentionally simple (no cloud abstraction) per ADR-007.
type ArtifactStorage struct {
	basePath string
}

func NewArtifactStorage(basePath string) *ArtifactStorage {
	return &ArtifactStorage{basePath: basePath}
}

func (s *ArtifactStorage) ensureBase() error {
	return os.MkdirAll(s.basePath, 0o755)
}

var safeNameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func safeFileName(name string) string {
	name = strings.ReplaceAll(name, "..", "_")
	return safeNameRe.ReplaceAllString(name, "_")
}

// SaveArtifact writes data to a deterministic path under the execution directory.
func (s *ArtifactStorage) SaveArtifact(executionID uuid.UUID, kind ArtifactKind, name string, data []byte) (string, int64, error) {
	if err := s.ensureBase(); err != nil {
		return "", 0, fmt.Errorf("create storage dir: %w", err)
	}
	dir := filepath.Join(s.basePath, executionID.String(), string(kind))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", 0, fmt.Errorf("create artifact dir: %w", err)
	}
	filename := safeFileName(name)
	if filename == "" {
		filename = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", 0, fmt.Errorf("write artifact: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, fmt.Errorf("stat artifact: %w", err)
	}
	return path, info.Size(), nil
}

// ReadArtifact returns the bytes for a stored artifact path.
func (s *ArtifactStorage) ReadArtifact(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// DeleteArtifact removes an artifact file.
func (s *ArtifactStorage) DeleteArtifact(path string) error {
	return os.Remove(path)
}

// RelativePath returns a path relative to the storage base for portability.
func (s *ArtifactStorage) RelativePath(path string) string {
	rel, err := filepath.Rel(s.basePath, path)
	if err != nil {
		return path
	}
	return rel
}

// FullPath joins a relative path with the storage base.
func (s *ArtifactStorage) FullPath(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(s.basePath, rel)
}
