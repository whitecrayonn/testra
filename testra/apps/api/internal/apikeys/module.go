package apikeys

import "database/sql"

func NewModule(db *sql.DB) *Handler {
	repo := NewSQLRepository(db)
	service := NewService(repo)
	return NewHandler(service)
}
