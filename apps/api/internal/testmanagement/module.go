package testmanagement

import (
	"database/sql"
)

type Module struct {
	Handler *Handler
}

func NewModule(db *sql.DB) *Module {
	repo := NewSQLRepository(db)
	service := NewService(repo)
	return &Module{Handler: NewHandler(service)}
}
