package apikeys

import "database/sql"

type Module struct {
	Handler *Handler
	Service *Service
}

func NewModule(db *sql.DB) *Module {
	repo := NewSQLRepository(db)
	service := NewService(repo)
	return &Module{
		Handler: NewHandler(service),
		Service: service,
	}
}
