package identity

import (
	"database/sql"
	"time"
)

func NewModule(db *sql.DB, jwtSecret string, jwtExpiry time.Duration, refreshExpiry time.Duration, refreshAbsolute time.Duration, smtpCfg SMTPConfig) *Handler {
	repo := NewSQLRepository(db)
	service := NewService(repo, jwtSecret, jwtExpiry, refreshExpiry, refreshAbsolute, smtpCfg)
	return NewHandler(service)
}
