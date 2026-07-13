package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/testra/testra/apps/api/internal/apikeys"
	"github.com/testra/testra/apps/api/internal/identity"
	"github.com/testra/testra/apps/api/internal/organization"
	"github.com/testra/testra/apps/api/internal/project"
	sharedmiddleware "github.com/testra/testra/apps/api/internal/shared/middleware"
	"github.com/testra/testra/apps/api/internal/workspace"
)

type Config struct {
	DB        *sql.DB
	JWTSecret string
	JWTExpiry time.Duration
}

func New(cfg Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	identityModule := identity.NewModule(cfg.DB, cfg.JWTSecret, cfg.JWTExpiry)
	organizationModule := organization.NewModule(cfg.DB)
	workspaceModule := workspace.NewModule(cfg.DB)
	projectModule := project.NewModule(cfg.DB)
	apiKeyModule := apikeys.NewModule(cfg.DB)

	authMiddleware := sharedmiddleware.Auth(sharedmiddleware.AuthConfig{JWTSecret: cfg.JWTSecret})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", identityModule.Register)
		r.Post("/auth/login", identityModule.Login)
		r.Post("/auth/password-reset/request", identityModule.RequestPasswordReset)
		r.Post("/auth/password-reset/confirm", identityModule.ResetPassword)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/auth/me", identityModule.Me)
			r.Post("/auth/mfa/setup", identityModule.SetupMFA)
			r.Post("/auth/mfa/verify", identityModule.VerifyMFA)
			r.Post("/auth/mfa/disable", identityModule.DisableMFA)

			r.Post("/organizations", organizationModule.Create)
			r.Get("/organizations", organizationModule.List)
			r.Get("/organizations/{id}", organizationModule.Get)

			r.Post("/workspaces", workspaceModule.Create)
			r.Get("/workspaces", workspaceModule.List)
			r.Get("/workspaces/{id}", workspaceModule.Get)

			r.Post("/projects", projectModule.Create)
			r.Get("/projects", projectModule.List)
			r.Get("/projects/{id}", projectModule.Get)

			r.Post("/api-keys", apiKeyModule.Create)
			r.Get("/api-keys", apiKeyModule.List)
			r.Delete("/api-keys/{id}", apiKeyModule.Revoke)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return r
}
