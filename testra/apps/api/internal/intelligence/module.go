package intelligence

import "database/sql"

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func New(sqlDB *sql.DB, mlServiceURL string) *Module {
	repo := NewSQLRepository(sqlDB)
	service := NewService(repo, NewMLClient(mlServiceURL))
	handler := NewHandler(service)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
