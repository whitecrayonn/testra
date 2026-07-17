package automationhub

import (
	"github.com/testra/testra/apps/api/internal/results"
)

type Module struct {
	Service *Service
	Handler *Handler
}

func NewModule(resultsRepo results.Repository) *Module {
	svc := NewService(resultsRepo)
	h := NewHandler(svc)
	return &Module{
		Service: svc,
		Handler: h,
	}
}
