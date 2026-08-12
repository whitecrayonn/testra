package audit

import (
	"github.com/testra/testra/apps/api/internal/shared/db"
)

type Module struct {
	Service *Service
	Handler *Handler
}

func NewModule(db db.DBTX) *Module {
	repo := NewSQLRepository(db)
	service := NewService(repo)
	return &Module{Service: service, Handler: NewHandler(service)}
}
