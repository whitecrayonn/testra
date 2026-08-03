package integrationhub

import (
	"database/sql"

	"github.com/testra/testra/apps/api/internal/audit"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func New(sqlDB *sql.DB, auditSvc *audit.Service, bus *eventbus.Bus) *Module {
	repo := NewSQLRepository(sqlDB)
	service := NewService(repo, auditSvc, bus, sqlDB)
	handler := NewHandler(service)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
