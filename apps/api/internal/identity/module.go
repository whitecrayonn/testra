package identity

import (
	"database/sql"
	"time"

	"github.com/testra/testra/apps/api/internal/shared/jwt"
)

func NewModule(db *sql.DB, tokenManager *jwt.Manager, jwtExpiry time.Duration, refreshExpiry time.Duration, refreshAbsolute time.Duration, smtpCfg SMTPConfig) *Handler {
	repo := NewSQLRepository(db)
	service := NewService(repo, tokenManager, jwtExpiry, refreshExpiry, refreshAbsolute, smtpCfg)
	return NewHandler(service, jwtExpiry, refreshExpiry)
}
