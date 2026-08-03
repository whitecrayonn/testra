package workspace

import (
	"database/sql"

	apidb "github.com/testra/testra/apps/api/internal/shared/db"
)

func NewModule(db *sql.DB) *Handler {
	repo := NewSQLRepository(db)
	dbHandle := apidb.Wrap(db)
	service := NewService(repo, dbHandle)
	return NewHandler(service)
}
