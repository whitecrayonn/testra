package results

import "database/sql"

type Module struct {
	Repository *SQLRepository
	Service    *Service
	Handler    *Handler
}

func NewModule(db *sql.DB) *Module {
	repo := NewSQLRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)
	return &Module{
		Repository: repo,
		Service:    svc,
		Handler:    h,
	}
}
