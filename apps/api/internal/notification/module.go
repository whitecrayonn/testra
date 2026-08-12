package notification

import (
	"database/sql"

	"github.com/testra/testra/apps/api/internal/shared/secrets"
)

type Module struct {
	Handler *Handler
	Service *Service
}

func NewModule(sqlDB *sql.DB, smtpHost, smtpPort, smtpFrom, smtpUsername string, secretProvider secrets.Provider, passwordSecret string) *Module {
	repo := NewSQLRepository(sqlDB)
	service := NewService(repo, SMTPConfig{
		Host:           smtpHost,
		Port:           smtpPort,
		From:           smtpFrom,
		Username:       smtpUsername,
		SecretProvider: secretProvider,
		PasswordSecret: passwordSecret,
	})
	return &Module{Handler: NewHandler(service), Service: service}
}
