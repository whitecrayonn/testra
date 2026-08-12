package billing

import "database/sql"

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func New(sqlDB *sql.DB, stripeSecretKey string) *Module {
	repo := NewSQLRepository(sqlDB)
	service := NewService(repo, NewPaymentProvider(stripeSecretKey))
	handler := NewHandler(service)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
