package automationhub

import (
	"github.com/testra/testra/apps/api/internal/defects"
	"github.com/testra/testra/apps/api/internal/results"
	"github.com/testra/testra/apps/api/internal/testmanagement"
)

type Module struct {
	Service *Service
	Handler *Handler
}

func NewModule(repo Repository, resultsRepo results.Repository, defectsRepo defects.Repository, testMgmtRepo testmanagement.Repository, storage *ArtifactStorage) *Module {
	svc := NewService(repo, resultsRepo, defectsRepo, testMgmtRepo, storage)
	h := NewHandler(svc)
	return &Module{
		Service: svc,
		Handler: h,
	}
}
