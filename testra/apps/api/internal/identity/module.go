package identity

import (
	"database/sql"
	"time"
)

func NewModule(db *sql.DB, jwtSecret string, jwtExpiry time.Duration) *Handler {
	repo := NewSQLRepository(db)
	service := NewService(repo, jwtSecret, jwtExpiry)
	return NewHandler(service)
}
