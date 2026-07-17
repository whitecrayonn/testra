package server

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/testra/testra/apps/api/internal/apikeys"
	"github.com/testra/testra/apps/api/internal/audit"
	"github.com/testra/testra/apps/api/internal/automationhub"
	"github.com/testra/testra/apps/api/internal/defects"
	"github.com/testra/testra/apps/api/internal/identity"
	"github.com/testra/testra/apps/api/internal/notification"
	"github.com/testra/testra/apps/api/internal/organization"
	"github.com/testra/testra/apps/api/internal/project"
	"github.com/testra/testra/apps/api/internal/rbac"
	"github.com/testra/testra/apps/api/internal/results"
	"github.com/testra/testra/apps/api/internal/shared/db"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/idempotency"
	sharedmiddleware "github.com/testra/testra/apps/api/internal/shared/middleware"
	"github.com/testra/testra/apps/api/internal/shared/tenant"
	"github.com/testra/testra/apps/api/internal/testmanagement"
	"github.com/testra/testra/apps/api/internal/workspace"
)

type Config struct {
	DB                  *sql.DB
	JWTSecret           string
	JWTExpiry           time.Duration
	RefreshExpiryDays   int
	RefreshAbsoluteDays int
	RedisAddr           string
	SMTPHost            string
	SMTPPort            string
	SMTPFrom            string
	CORSAllowedOrigins  string
	IdempotencyKeyTTL   time.Duration
}

type apiKeyValidatorAdapter struct {
	service *apikeys.Service
}

func (a *apiKeyValidatorAdapter) Validate(ctx context.Context, rawKey string) (sharedmiddleware.APIKeyInfo, error) {
	return a.service.Validate(ctx, rawKey)
}

func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, o := range origins {
				if o == origin {
					w.Header().Set("Access-Control-Allow-Origin", o)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					break
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func New(cfg Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(corsMiddleware(cfg.CORSAllowedOrigins))
	r.Use(sharedmiddleware.MaxBodySize(sharedmiddleware.DefaultMaxBodySize))

	refreshExpiry := time.Duration(cfg.RefreshExpiryDays) * 24 * time.Hour
	refreshAbsolute := time.Duration(cfg.RefreshAbsoluteDays) * 24 * time.Hour
	smtpCfg := identity.SMTPConfig{Host: cfg.SMTPHost, Port: cfg.SMTPPort, From: cfg.SMTPFrom}

	identityModule := identity.NewModule(cfg.DB, cfg.JWTSecret, cfg.JWTExpiry, refreshExpiry, refreshAbsolute, smtpCfg)
	organizationModule := organization.NewModule(cfg.DB)
	workspaceModule := workspace.NewModule(cfg.DB)
	projectModule := project.NewModule(cfg.DB)
	apiKeyModule := apikeys.NewModule(cfg.DB)
	testMgmtModule := testmanagement.NewModule(cfg.DB)
	resultsModule := results.NewModule(cfg.DB)
	defectsModule := defects.NewModule(cfg.DB)
	notificationModule := notification.NewModule(cfg.DB, cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
	automationHubModule := automationhub.NewModule(resultsModule.Repository)
	dbHandle := db.Wrap(cfg.DB)

	tenantResolver := tenant.NewResolver(dbHandle)
	rbacLoader := rbac.NewSQLPermissionLoader(dbHandle)

	idempotencyStore := idempotency.NewPostgresStore(dbHandle)
	auditSvc := audit.NewModule(dbHandle).Service

	authMiddleware := sharedmiddleware.Auth(sharedmiddleware.AuthConfig{JWTSecret: cfg.JWTSecret})
	rbacCfg := sharedmiddleware.RBACConfig{Loader: rbacLoader}

	localLimiter := sharedmiddleware.NewLocalRateLimiter()
	rateLimitCfg := sharedmiddleware.RateLimitConfig{Limiter: localLimiter}
	_ = rateLimitCfg

	auditLogFn := func(input sharedmiddleware.AuditLogInput) {
		auditSvc.Log(context.Background(), audit.LogInput{
			UserID:     input.UserID,
			Action:     input.Action,
			Resource:   input.Resource,
			ResourceID: input.ResourceID,
			IPAddress:  input.IPAddress,
		})
	}

	apiKeyValidator := &apiKeyValidatorAdapter{service: apiKeyModule.Service}

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(sharedmiddleware.RateLimit(rateLimitCfg, sharedmiddleware.RateLimitByIP(), sharedmiddleware.RateLimitRule{Limit: 20, Window: time.Minute}))
			r.Post("/auth/register", identityModule.Register)
			r.Post("/auth/login", identityModule.Login)
			r.Post("/auth/refresh", identityModule.RefreshToken)
			r.Post("/auth/password-reset/request", identityModule.RequestPasswordReset)
			r.Post("/auth/password-reset/confirm", identityModule.ResetPassword)
		})

		r.Group(func(r chi.Router) {
			r.Use(sharedmiddleware.RateLimit(rateLimitCfg, sharedmiddleware.RateLimitByAPIKey(), sharedmiddleware.RateLimitRule{Limit: 100, Window: time.Minute}))
			r.Use(sharedmiddleware.APIKeyAuth(cfg.DB, apiKeyValidator))
			r.Use(sharedmiddleware.RequireScope("runs:ingest"))
			r.Use(sharedmiddleware.AuditLog("test_run.ingest", "test_run",
				func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
				func(r *http.Request) string { return "" },
				auditLogFn,
			))
			r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, "ingest", cfg.IdempotencyKeyTTL))
			r.Post("/ingest", automationHubModule.Handler.Ingest)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/auth/me", identityModule.Me)
			r.Post("/auth/mfa/setup", identityModule.SetupMFA)
			r.Post("/auth/mfa/verify", identityModule.VerifyMFA)
			r.Post("/auth/mfa/disable", identityModule.DisableMFA)

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.OrgIDFromURLParam("id"),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "orgs:read")).Get("/organizations/{id}", organizationModule.Get)
			})

			r.Post("/organizations", organizationModule.Create)
			r.Get("/organizations", organizationModule.List)

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.OrgIDFromBody("organization_id"),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "workspaces:create")).Post("/workspaces", workspaceModule.Create)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.OrgIDFromQuery("organization_id"),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "workspaces:read")).Get("/workspaces", workspaceModule.List)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "workspaces:read")).Get("/workspaces/{id}", workspaceModule.Get)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "projects:create")).Post("/projects", projectModule.Create)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "apikeys:create"),
					sharedmiddleware.AuditLog("api_key.create", "api_key",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/api-keys", apiKeyModule.Handler.Create)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "projects:read")).Get("/projects", projectModule.List)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "apikeys:read")).Get("/api-keys", apiKeyModule.Handler.List)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "projects:read")).Get("/projects/{id}", projectModule.Get)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:create"),
					sharedmiddleware.AuditLog("test_folder.create", "test_folder",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/test-folders", testMgmtModule.Handler.CreateFolder)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:create"),
					sharedmiddleware.AuditLog("test_suite.create", "test_suite",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/test-suites", testMgmtModule.Handler.CreateSuite)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:create"),
					sharedmiddleware.AuditLog("test_case.create", "test_case",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/test-cases", testMgmtModule.Handler.CreateCase)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-folders", testMgmtModule.Handler.ListFolders)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-suites", testMgmtModule.Handler.ListSuites)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-cases/search", testMgmtModule.Handler.SearchCases)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromQuery("project_id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-cases", testMgmtModule.Handler.ListCases)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-cases/{id}", testMgmtModule.Handler.GetCase)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-cases/{id}/versions", testMgmtModule.Handler.ListVersions)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:update"),
					sharedmiddleware.AuditLog("test_case.update", "test_case",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/test-cases/{id}", testMgmtModule.Handler.UpdateCase)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:delete"),
					sharedmiddleware.AuditLog("test_case.delete", "test_case",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/test-cases/{id}", testMgmtModule.Handler.DeleteCase)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-folders/{id}", testMgmtModule.Handler.GetFolder)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-suites/{id}", testMgmtModule.Handler.GetSuite)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:update"),
					sharedmiddleware.AuditLog("test_folder.update", "test_folder",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/test-folders/{id}", testMgmtModule.Handler.UpdateFolder)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:update"),
					sharedmiddleware.AuditLog("test_suite.update", "test_suite",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/test-suites/{id}", testMgmtModule.Handler.UpdateSuite)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:delete"),
					sharedmiddleware.AuditLog("test_folder.delete", "test_folder",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/test-folders/{id}", testMgmtModule.Handler.DeleteFolder)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "tests:delete"),
					sharedmiddleware.AuditLog("test_suite.delete", "test_suite",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/test-suites/{id}", testMgmtModule.Handler.DeleteSuite)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.APIKeyToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "apikeys:delete"),
					sharedmiddleware.AuditLog("api_key.revoke", "api_key",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/api-keys/{id}", apiKeyModule.Handler.Revoke)
			})

			// --- Phase 3: Execution & Results ---

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:create"),
					sharedmiddleware.AuditLog("test_run.create", "test_run",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/test-runs", resultsModule.Handler.CreateRun)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromQuery("project_id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "runs:read")).Get("/test-runs", resultsModule.Handler.ListRuns)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.RunToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "runs:read")).Get("/test-runs/{id}", resultsModule.Handler.GetRun)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "runs:read")).Get("/test-runs/{id}/items", resultsModule.Handler.ListItems)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "runs:read")).Get("/test-runs/{id}/stream", resultsModule.Handler.StreamRunProgress)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:update"),
					sharedmiddleware.AuditLog("test_run.update", "test_run",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/test-runs/{id}", resultsModule.Handler.UpdateRunStatus)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:delete"),
					sharedmiddleware.AuditLog("test_run.delete", "test_run",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/test-runs/{id}", resultsModule.Handler.DeleteRun)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.RunItemToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:update"),
					sharedmiddleware.AuditLog("test_run_item.update", "test_run_item",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/test-run-items/{id}", resultsModule.Handler.UpdateItemStatus)
			})

		r.Group(func(r chi.Router) {
			r.Use(sharedmiddleware.TenantContext(cfg.DB,
				sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromBody("project_id"), tenantResolver),
				tenantResolver,
			))
			r.With(
				sharedmiddleware.RequirePermission(rbacCfg, "defects:create"),
				sharedmiddleware.AuditLog("defect.create", "defect",
					func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
					func(r *http.Request) string { return "" },
					auditLogFn,
				),
			).Post("/defects", defectsModule.Create)
		})

		r.Group(func(r chi.Router) {
			r.Use(sharedmiddleware.TenantContext(cfg.DB,
				sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromQuery("project_id"), tenantResolver),
				tenantResolver,
			))
			r.With(sharedmiddleware.RequirePermission(rbacCfg, "defects:read")).Get("/defects", defectsModule.List)
		})

		r.Group(func(r chi.Router) {
			r.Use(sharedmiddleware.TenantContext(cfg.DB,
				sharedmiddleware.DefectToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
				tenantResolver,
			))
			r.With(sharedmiddleware.RequirePermission(rbacCfg, "defects:read")).Get("/defects/{id}", defectsModule.Get)
			r.With(
				sharedmiddleware.RequirePermission(rbacCfg, "defects:update"),
				sharedmiddleware.AuditLog("defect.update", "defect",
					func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
					func(r *http.Request) string { return chi.URLParam(r, "id") },
					auditLogFn,
				),
			).Put("/defects/{id}", defectsModule.Update)
			r.With(
				sharedmiddleware.RequirePermission(rbacCfg, "defects:delete"),
				sharedmiddleware.AuditLog("defect.delete", "defect",
					func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
					func(r *http.Request) string { return chi.URLParam(r, "id") },
					auditLogFn,
				),
			).Delete("/defects/{id}", defectsModule.Delete)
		})

			// --- Notification Center ---
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))

				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notifications:read")).Get("/notifications", notificationModule.Handler.List)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notifications:read")).Get("/notifications/unread-count", notificationModule.Handler.UnreadCount)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notifications:update")).Patch("/notifications/{id}", notificationModule.Handler.MarkRead)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notifications:delete")).Delete("/notifications/{id}", notificationModule.Handler.Delete)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "notifications:create"),
					sharedmiddleware.AuditLog("notification.create", "notification",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/notifications", notificationModule.Handler.Create)

				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notification_preferences:read")).Get("/notification-preferences", notificationModule.Handler.GetPreferences)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notification_preferences:update")).Put("/notification-preferences", notificationModule.Handler.UpdatePreferences)

				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notification_channels:read")).Get("/notification-channels", notificationModule.Handler.ListChannels)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "notification_channels:create"),
					sharedmiddleware.AuditLog("notification_channel.create", "notification_channel",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/notification-channels", notificationModule.Handler.CreateChannel)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "notification_channels:update"),
					sharedmiddleware.AuditLog("notification_channel.update", "notification_channel",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/notification-channels/{id}", notificationModule.Handler.UpdateChannel)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "notification_channels:delete"),
					sharedmiddleware.AuditLog("notification_channel.delete", "notification_channel",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/notification-channels/{id}", notificationModule.Handler.DeleteChannel)
			})
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		apihttp.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return r
}
