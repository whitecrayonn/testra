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
// is wired in server.go.
func NewModule(db *sql.DB, testMgmtRepo testmanagement.Repository) *Module {
	repo := NewSQLRepository(db)
	service := NewService(repo, testMgmtRepo)
	return &Module{Handler: NewHandler(service)}
}
