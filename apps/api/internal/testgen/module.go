package testgen

import (
	"database/sql"

	"github.com/testra/testra/apps/api/internal/testmanagement"
)

type Module struct {
	Handler *Handler
}

// NewModule takes a testmanagement.Repository (not just *sql.DB) so the
// composition root (server.go) can pass the exact same repository instance
// used to build the testmanagement module — mirroring how automationhub.NewModule
// is wired in server.go. mlServiceURL/mlAPIKey configure the LLM-backed
// GenerateFromFile path; when mlServiceURL is empty that path fails loudly
// with "not configured" instead of silently doing nothing.
func NewModule(db *sql.DB, testMgmtRepo testmanagement.Repository, mlServiceURL, mlAPIKey string) *Module {
	repo := NewSQLRepository(db)
	fileGen := NewFileGenerationClient(mlServiceURL, mlAPIKey)
	service := NewService(repo, testMgmtRepo, fileGen)
	return &Module{Handler: NewHandler(service)}
}
