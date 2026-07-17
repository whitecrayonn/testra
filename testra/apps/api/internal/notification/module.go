package notification

import (
	"database/sql"

	"github.com/testra/testra/apps/api/internal/shared/db"
)

type Module struct {
	Handler *Handler
	Service *Service
}

func NewModule(sqlDB *sql.DB, smtpHost, smtpPort, smtpFrom string) *Module {
	dbHandle := db.Wrap(sqlDB)
	repo := NewSQLRepository(sqlDB)
	service := NewService(repo, SMTPConfig{Host: smtpHost, Port: smtpPort, From: smtpFrom})
	_ = dbHandle
	return &Module{Handler: NewHandler(service), Service: service}
}
