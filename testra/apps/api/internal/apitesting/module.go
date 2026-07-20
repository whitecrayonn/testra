package apitesting

import "database/sql"

// Module holds the API testing HTTP handler.
type Module struct {
	Handler *Handler
}

// NewModule wires repository, service, and handler for the API testing module.
func NewModule(db *sql.DB) *Module {
	repo := NewSQLRepository(db)
	service := NewService(repo)
	return &Module{Handler: NewHandler(service)}
}
