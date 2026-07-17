package audit

import (
	"github.com/testra/testra/apps/api/internal/shared/db"
)

type Module struct {
	Service *Service
}

func NewModule(db db.DBTX) *Module {
	repo := NewSQLRepository(db)
	return &Module{Service: NewService(repo)}
}
