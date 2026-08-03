package analytics

import "database/sql"

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func New(sqlDB *sql.DB) *Module {
	repo := NewSQLRepository(sqlDB)
	service := NewService(repo)
	handler := NewHandler(service)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
