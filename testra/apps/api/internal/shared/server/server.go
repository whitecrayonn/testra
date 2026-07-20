package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/testra/testra/apps/api/internal/analytics"
	"github.com/testra/testra/apps/api/internal/apikeys"
	"github.com/testra/testra/apps/api/internal/apitesting"
	"github.com/testra/testra/apps/api/internal/audit"
	"github.com/testra/testra/apps/api/internal/automationhub"
	"github.com/testra/testra/apps/api/internal/billing"
	"github.com/testra/testra/apps/api/internal/defects"
	"github.com/testra/testra/apps/api/internal/identity"
	"github.com/testra/testra/apps/api/internal/integrationhub"
	"github.com/testra/testra/apps/api/internal/intelligence"
	"github.com/testra/testra/apps/api/internal/notification"
	"github.com/testra/testra/apps/api/internal/organization"
	"github.com/testra/testra/apps/api/internal/project"
	"github.com/testra/testra/apps/api/internal/rbac"
	"github.com/testra/testra/apps/api/internal/results"
	"github.com/testra/testra/apps/api/internal/shared/db"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/idempotency"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
	sharedmiddleware "github.com/testra/testra/apps/api/internal/shared/middleware"
	"github.com/testra/testra/apps/api/internal/shared/secrets"
	"github.com/testra/testra/apps/api/internal/shared/tenant"
	"github.com/testra/testra/apps/api/internal/testmanagement"
	"github.com/testra/testra/apps/api/internal/workspace"
)

type Config struct {
	DB                  *sql.DB
	JWTManager          *jwt.Manager
	JWTExpiry           time.Duration
	RefreshExpiryDays   int
	RefreshAbsoluteDays int
	RedisAddr           string
	SMTPHost            string
	SMTPPort            string
	SMTPFrom            string
	SMTPUsername        string
	SMTPPasswordSecret  string
	SecretProvider      secrets.Provider
	CORSAllowedOrigins  string
	IdempotencyKeyTTL   time.Duration
	MLServiceURL        string
	StripeSecretKey     string
	StripePriceID       string
}

type apiKeyValidatorAdapter struct {
	service *apikeys.Service
}

func (a *apiKeyValidatorAdapter) Validate(ctx context.Context, rawKey string) (sharedmiddleware.APIKeyInfo, error) {
	return a.service.Validate(ctx, rawKey)
}

func apiSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, private, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Vary", "Authorization, Origin, Cookie")
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Vary: Origin prevents caches from serving the CORS response to a
			// request with a different Origin header.
			w.Header().Add("Vary", "Origin")

			origin := r.Header.Get("Origin")
			for _, o := range origins {
				if o != "" && o == origin {
					w.Header().Set("Access-Control-Allow-Origin", o)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-API-Key, X-CSRF-Token")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Max-Age", "600")
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
	r.Use(middleware.RequestID)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				w.Header().Set(middleware.RequestIDHeader, reqID)
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sharedmiddleware.RequestLogger(sharedmiddleware.NewStructuredLogFormatter(slog.New(slog.NewJSONHandler(os.Stdout, nil)))))
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(corsMiddleware(cfg.CORSAllowedOrigins))
	r.Use(sharedmiddleware.MaxBodySize(sharedmiddleware.DefaultMaxBodySize))

	refreshExpiry := time.Duration(cfg.RefreshExpiryDays) * 24 * time.Hour
	refreshAbsolute := time.Duration(cfg.RefreshAbsoluteDays) * 24 * time.Hour
	smtpCfg := identity.SMTPConfig{
		Host:           cfg.SMTPHost,
		Port:           cfg.SMTPPort,
		From:           cfg.SMTPFrom,
		Username:       cfg.SMTPUsername,
		SecretProvider: cfg.SecretProvider,
		PasswordSecret: cfg.SMTPPasswordSecret,
	}

	identityModule := identity.NewModule(cfg.DB, cfg.JWTManager, cfg.JWTExpiry, refreshExpiry, refreshAbsolute, smtpCfg)
	organizationModule := organization.NewModule(cfg.DB)
	workspaceModule := workspace.NewModule(cfg.DB)
	projectModule := project.NewModule(cfg.DB)
	apiKeyModule := apikeys.NewModule(cfg.DB)
	testMgmtModule := testmanagement.NewModule(cfg.DB)
	resultsModule := results.NewModule(cfg.DB)
	defectsModule := defects.NewModule(cfg.DB)
	apiTestingModule := apitesting.NewModule(cfg.DB)
	notificationModule := notification.NewModule(cfg.DB, cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, cfg.SMTPUsername, cfg.SecretProvider, cfg.SMTPPasswordSecret)

	automationStoragePath := os.Getenv("AUTOMATION_STORAGE_PATH")
	if automationStoragePath == "" {
		automationStoragePath = "storage/automation"
	}
	automationRepo := automationhub.NewSQLRepository(cfg.DB)
	automationStorage := automationhub.NewArtifactStorage(automationStoragePath)
	automationHubModule := automationhub.NewModule(automationRepo, resultsModule.Repository, defects.NewSQLRepository(cfg.DB), testmanagement.NewSQLRepository(cfg.DB), automationStorage)
	analyticsModule := analytics.New(cfg.DB)
	intelligenceModule := intelligence.New(cfg.DB, cfg.MLServiceURL)
	dbHandle := db.Wrap(cfg.DB)

	tenantResolver := tenant.NewResolver(dbHandle)
	rbacLoader := rbac.NewSQLPermissionLoader(dbHandle)

	idempotencyStore := idempotency.NewPostgresStore(dbHandle)
	auditSvc := audit.NewModule(dbHandle).Service

	eventBus := eventbus.New(256)
	eventbus.SetDefault(eventBus)
	integrationhubModule := integrationhub.New(cfg.DB, auditSvc, eventBus)
	billingModule := billing.New(cfg.DB, cfg.StripeSecretKey)

	authMiddleware := sharedmiddleware.Auth(sharedmiddleware.AuthConfig{TokenManager: cfg.JWTManager})
	rbacCfg := sharedmiddleware.RBACConfig{Loader: rbacLoader}

	r.Get("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		if cfg.JWTManager == nil {
			apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL", "JWT manager not configured")
			return
		}
		jwks, err := cfg.JWTManager.MarshalJWKS()
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL", "failed to marshal JWKS")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwks)
	})

	localLimiter := sharedmiddleware.NewLocalRateLimiter()
	var limiter sharedmiddleware.RateLimiter = localLimiter
	if cfg.RedisAddr != "" {
		limiter = sharedmiddleware.NewRedisRateLimiter(cfg.RedisAddr, localLimiter)
	}
	rateLimitCfg := sharedmiddleware.RateLimitConfig{Limiter: limiter}
	authRateLimitCfg := sharedmiddleware.RateLimitConfig{Limiter: limiter, FailClosed: true}

	auditLogFn := func(input sharedmiddleware.AuditLogInput) {
		// Use a bounded, detached context so audit persistence is not tied to
		// the already-completed request lifecycle but still cannot hang.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		auditSvc.Log(ctx, audit.LogInput{
			UserID:     input.UserID,
			Action:     input.Action,
			Resource:   input.Resource,
			ResourceID: input.ResourceID,
			IPAddress:  input.IPAddress,
			Metadata:   map[string]string{"status": fmt.Sprintf("%d", input.StatusCode)},
		})
	}

	apiKeyValidator := &apiKeyValidatorAdapter{service: apiKeyModule.Service}

	csrfMiddleware := sharedmiddleware.CSRF(sharedmiddleware.CSRFConfig{
		Skip: func(r *http.Request) bool {
			path := r.URL.Path
			return strings.HasPrefix(path, "/api/v1/auth/login") ||
				strings.HasPrefix(path, "/api/v1/auth/register") ||
				strings.HasPrefix(path, "/api/v1/auth/refresh") ||
				strings.HasPrefix(path, "/api/v1/auth/password-reset")
		},
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(apiSecurityHeaders)

		r.Group(func(r chi.Router) {
			r.Use(sharedmiddleware.RateLimit(authRateLimitCfg, sharedmiddleware.RateLimitByIP(), sharedmiddleware.RateLimitRule{Limit: 20, Window: time.Minute}))
			r.Use(csrfMiddleware)
			r.Get("/auth/csrf", identityModule.CSRF)
			r.Post("/auth/register", identityModule.Register)
			r.Post("/auth/login", identityModule.Login)
			r.Post("/auth/refresh", identityModule.RefreshToken)
			r.Post("/auth/logout", identityModule.Logout)
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
			r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
			r.Post("/ingest", automationHubModule.Handler.Ingest)
		})

		r.Group(func(r chi.Router) {
			r.Use(csrfMiddleware)
			r.Use(authMiddleware)
			r.Get("/auth/me", identityModule.Me)
			r.Post("/auth/logout-all", identityModule.LogoutAllDevices)
			r.Post("/auth/mfa/setup", identityModule.SetupMFA)
			r.Post("/auth/mfa/verify", identityModule.VerifyMFA)
			r.Post("/auth/mfa/disable", identityModule.DisableMFA)

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.OrgIDFromURLParam("id"),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "orgs:read")).Get("/organizations/{id}", organizationModule.Get)
			})

			r.Post("/organizations", organizationModule.Create)
			r.Get("/organizations", organizationModule.List)

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.OrgIDFromBody("organization_id"),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "workspaces:create")).Post("/workspaces", workspaceModule.Create)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.OrgIDFromQuery("organization_id"),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "workspaces:read")).Get("/workspaces", workspaceModule.List)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "workspaces:read")).Get("/workspaces/{id}", workspaceModule.Get)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "projects:read")).Get("/projects", projectModule.List)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "apikeys:read")).Get("/api-keys", apiKeyModule.Handler.List)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "projects:read")).Get("/projects/{id}", projectModule.Get)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-folders", testMgmtModule.Handler.ListFolders)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-suites", testMgmtModule.Handler.ListSuites)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-cases/search", testMgmtModule.Handler.SearchCases)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromQuery("project_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "tests:read")).Get("/test-cases", testMgmtModule.Handler.ListCases)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "runs:read")).Get("/test-runs", resultsModule.Handler.ListRuns)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.RunToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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
					sharedmiddleware.RequirePermission(rbacCfg, "runs:create"),
					sharedmiddleware.AuditLog("test_run.clone", "test_run",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Post("/test-runs/{id}/clone", resultsModule.Handler.CloneRun)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:create"),
					sharedmiddleware.AuditLog("test_run.rerun", "test_run",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Post("/test-runs/{id}/rerun", resultsModule.Handler.RerunRun)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:execute"),
					sharedmiddleware.AuditLog("test_run_item.bulk_update", "test_run_item",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Post("/test-runs/{id}/bulk", resultsModule.Handler.BulkUpdateItems)
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
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:update"),
					sharedmiddleware.AuditLog("test_run_item.update", "test_run_item",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/test-run-items/{id}", resultsModule.Handler.UpdateItemStatus)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:execute"),
					sharedmiddleware.AuditLog("test_run_item.execute", "test_run_item",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Post("/test-run-items/{id}/execute", resultsModule.Handler.ExecuteItem)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "runs:read")).Get("/test-run-items/{id}/history", resultsModule.Handler.ListItemHistory)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:execute"),
					sharedmiddleware.AuditLog("test_run_item.evidence.attach", "test_run_item_evidence",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Post("/test-run-items/{id}/evidence", resultsModule.Handler.AttachEvidence)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "runs:read")).Get("/test-run-items/{id}/evidence", resultsModule.Handler.ListEvidence)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:execute"),
					sharedmiddleware.AuditLog("test_run_item.evidence.delete", "test_run_item_evidence",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "evidenceId") },
						auditLogFn,
					),
				).Delete("/test-run-items/{id}/evidence/{evidenceId}", resultsModule.Handler.DeleteEvidence)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "defects:create"),
					sharedmiddleware.AuditLog("test_run_item.defect.link", "test_run_item_defect",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Post("/test-run-items/{id}/defects", resultsModule.Handler.LinkDefect)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "runs:read")).Get("/test-run-items/{id}/defects", resultsModule.Handler.ListItemDefects)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "defects:delete"),
					sharedmiddleware.AuditLog("test_run_item.defect.unlink", "test_run_item_defect",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "defectId") },
						auditLogFn,
					),
				).Delete("/test-run-items/{id}/defects/{defectId}", resultsModule.Handler.UnlinkDefect)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromBody("project_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "testcases:create"),
					sharedmiddleware.AuditLog("test_plan.create", "test_plan",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/test-plans", resultsModule.Handler.CreatePlan)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromQuery("project_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "testcases:read")).Get("/test-plans", resultsModule.Handler.ListPlans)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.TestPlanToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "testcases:read")).Get("/test-plans/{id}", resultsModule.Handler.GetPlan)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "testcases:update"),
					sharedmiddleware.AuditLog("test_plan.update", "test_plan",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/test-plans/{id}", resultsModule.Handler.UpdatePlan)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "testcases:delete"),
					sharedmiddleware.AuditLog("test_plan.delete", "test_plan",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/test-plans/{id}", resultsModule.Handler.DeletePlan)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "testcases:read")).Get("/test-plans/{id}/items", resultsModule.Handler.ListPlanItems)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "runs:create"),
					sharedmiddleware.AuditLog("test_plan.run.create", "test_run",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Post("/test-plans/{id}/runs", resultsModule.Handler.CreateRunFromPlan)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.ProjectToOrg(sharedmiddleware.OrgIDFromBody("project_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "defects:read")).Get("/defects", defectsModule.List)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.DefectToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
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

			// --- Automation Hub ---
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:create"),
					sharedmiddleware.AuditLog("automation_project.create", "automation_project",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/automation/projects", automationHubModule.Handler.CreateProject)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "automation:read")).Get("/automation/projects", automationHubModule.Handler.ListProjects)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.AutomationProjectToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "automation:read")).Get("/automation/projects/{id}", automationHubModule.Handler.GetProject)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:update"),
					sharedmiddleware.AuditLog("automation_project.update", "automation_project",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/automation/projects/{id}", automationHubModule.Handler.UpdateProject)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:delete"),
					sharedmiddleware.AuditLog("automation_project.delete", "automation_project",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/automation/projects/{id}", automationHubModule.Handler.DeleteProject)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "automation:read")).Get("/automation/executions", automationHubModule.Handler.ListExecutions)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.AutomationProjectToOrg(sharedmiddleware.OrgIDFromURLParam("project_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:execute"),
					sharedmiddleware.AuditLog("automation_execution.import", "automation_execution",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/automation/projects/{project_id}/executions", automationHubModule.Handler.ImportExecution)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "automation:read")).Get("/automation/projects/{project_id}/executions", automationHubModule.Handler.ListExecutions)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.AutomationExecutionToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "automation:read")).Get("/automation/executions/{id}", automationHubModule.Handler.GetExecution)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:execute"),
					sharedmiddleware.AuditLog("automation_execution.rerun", "automation_execution",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Post("/automation/executions/{id}/rerun", automationHubModule.Handler.RerunExecution)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:delete"),
					sharedmiddleware.AuditLog("automation_execution.delete", "automation_execution",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/automation/executions/{id}", automationHubModule.Handler.DeleteExecution)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "automation:read")).Get("/automation/executions/{id}/artifacts", automationHubModule.Handler.ListArtifacts)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:execute"),
					sharedmiddleware.AuditLog("automation_artifact.upload", "automation_artifact",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/automation/executions/{id}/artifacts", automationHubModule.Handler.UploadArtifact)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "automation:read")).Get("/automation/executions/{id}/logs", automationHubModule.Handler.ListLogs)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:create"),
					sharedmiddleware.AuditLog("automation_log.create", "automation_log",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/automation/executions/{id}/logs", automationHubModule.Handler.AddLog)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.AutomationArtifactToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "automation:read")).Get("/automation/artifacts/{id}", automationHubModule.Handler.GetArtifact)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "automation:delete"),
					sharedmiddleware.AuditLog("automation_artifact.delete", "automation_artifact",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/automation/artifacts/{id}", automationHubModule.Handler.DeleteArtifact)
			})

			// --- API Testing ---
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:create"),
					sharedmiddleware.AuditLog("api_collection.create", "api_collection",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/api-collections", apiTestingModule.Handler.CreateCollection)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:create"),
					sharedmiddleware.AuditLog("api_folder.create", "api_folder",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/api-folders", apiTestingModule.Handler.CreateFolder)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:create"),
					sharedmiddleware.AuditLog("api_environment.create", "api_environment",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/api-environments", apiTestingModule.Handler.CreateEnvironment)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:create"),
					sharedmiddleware.AuditLog("api_request.create", "api_request",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/api-requests", apiTestingModule.Handler.CreateRequest)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:execute"),
					sharedmiddleware.AuditLog("api_request.execute", "api_request",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/api-executions", apiTestingModule.Handler.ExecuteRequest)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-collections", apiTestingModule.Handler.ListCollections)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-environments", apiTestingModule.Handler.ListEnvironments)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-executions", apiTestingModule.Handler.ListRequestHistory)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-requests/search", apiTestingModule.Handler.SearchRequests)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.APICollectionToOrg(sharedmiddleware.OrgIDFromQuery("collection_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-folders", apiTestingModule.Handler.ListFolders)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-requests", apiTestingModule.Handler.ListRequests)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.APICollectionToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-collections/{id}", apiTestingModule.Handler.GetCollection)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:update"),
					sharedmiddleware.AuditLog("api_collection.update", "api_collection",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/api-collections/{id}", apiTestingModule.Handler.UpdateCollection)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:delete"),
					sharedmiddleware.AuditLog("api_collection.delete", "api_collection",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/api-collections/{id}", apiTestingModule.Handler.DeleteCollection)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.APIFolderToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-folders/{id}", apiTestingModule.Handler.GetFolder)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:update"),
					sharedmiddleware.AuditLog("api_folder.update", "api_folder",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/api-folders/{id}", apiTestingModule.Handler.UpdateFolder)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:delete"),
					sharedmiddleware.AuditLog("api_folder.delete", "api_folder",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/api-folders/{id}", apiTestingModule.Handler.DeleteFolder)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.APIEnvironmentToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-environments/{id}", apiTestingModule.Handler.GetEnvironment)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:update"),
					sharedmiddleware.AuditLog("api_environment.update", "api_environment",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/api-environments/{id}", apiTestingModule.Handler.UpdateEnvironment)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:delete"),
					sharedmiddleware.AuditLog("api_environment.delete", "api_environment",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/api-environments/{id}", apiTestingModule.Handler.DeleteEnvironment)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.APIRequestToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-requests/{id}", apiTestingModule.Handler.GetRequest)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-requests/{id}/history", apiTestingModule.Handler.ListRequestHistory)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:update"),
					sharedmiddleware.AuditLog("api_request.update", "api_request",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/api-requests/{id}", apiTestingModule.Handler.UpdateRequest)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "api_tests:delete"),
					sharedmiddleware.AuditLog("api_request.delete", "api_request",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/api-requests/{id}", apiTestingModule.Handler.DeleteRequest)
			})

			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.APIRequestHistoryToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "api_tests:read")).Get("/api-executions/{id}", apiTestingModule.Handler.GetRequestHistory)
			})

			// --- Notification Center ---
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))

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

				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notifications:read")).Get("/notification-templates", notificationModule.Handler.ListTemplates)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notifications:read")).Get("/notification-templates/{id}", notificationModule.Handler.GetTemplate)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "notifications:create"),
					sharedmiddleware.AuditLog("notification_template.create", "notification_template",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return "" },
						auditLogFn,
					),
				).Post("/notification-templates", notificationModule.Handler.CreateTemplate)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "notifications:update"),
					sharedmiddleware.AuditLog("notification_template.update", "notification_template",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Put("/notification-templates/{id}", notificationModule.Handler.UpdateTemplate)
				r.With(
					sharedmiddleware.RequirePermission(rbacCfg, "notifications:delete"),
					sharedmiddleware.AuditLog("notification_template.delete", "notification_template",
						func(r *http.Request) uuid.UUID { uid, _ := sharedmiddleware.UserIDFromContext(r.Context()); return uid },
						func(r *http.Request) string { return chi.URLParam(r, "id") },
						auditLogFn,
					),
				).Delete("/notification-templates/{id}", notificationModule.Handler.DeleteTemplate)

				r.With(sharedmiddleware.RequirePermission(rbacCfg, "notifications:read")).Get("/notification-history", notificationModule.Handler.ListHistory)
			})

			// --- Analytics ---
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:create")).Post("/analytics/dashboards", analyticsModule.Handler.CreateDashboard)
			})
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:read")).Get("/analytics/dashboards", analyticsModule.Handler.ListDashboards)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:read")).Get("/analytics/summary", analyticsModule.Handler.GetSummary)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:read")).Get("/analytics/trends", analyticsModule.Handler.GetTrends)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:read")).Get("/analytics/metrics", analyticsModule.Handler.GetMetrics)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:read")).Get("/analytics/activity", analyticsModule.Handler.GetRecentActivity)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:read")).Get("/analytics/export/csv", analyticsModule.Handler.ExportMetricsCSV)
			})
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:read")).Get("/analytics/dashboards/{id}", analyticsModule.Handler.GetDashboard)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:update")).Put("/analytics/dashboards/{id}", analyticsModule.Handler.UpdateDashboard)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "analytics:delete")).Delete("/analytics/dashboards/{id}", analyticsModule.Handler.DeleteDashboard)
			})

			// --- Intelligence ---
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "intelligence:create")).Post("/intelligence/predict-flaky", intelligenceModule.Handler.PredictFlaky)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "intelligence:create")).Post("/intelligence/classify-failure", intelligenceModule.Handler.ClassifyFailure)
			})
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "intelligence:read")).Get("/intelligence/flaky-tests", intelligenceModule.Handler.ListPredictions)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "intelligence:read")).Get("/intelligence/failure-clusters", intelligenceModule.Handler.ListClusters)
			})
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "intelligence:read")).Get("/intelligence/flaky-tests/{id}", intelligenceModule.Handler.GetPrediction)
			})

			// --- Integration Hub ---
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:create")).Post("/integrations", integrationhubModule.Handler.CreateIntegration)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:create")).Post("/integrations/dispatch", integrationhubModule.Handler.DispatchEvent)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:update")).Post("/integrations/{id}/enable", integrationhubModule.Handler.EnableIntegration)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:update")).Post("/integrations/{id}/disable", integrationhubModule.Handler.DisableIntegration)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:update")).Post("/integration-events/{id}/retry", integrationhubModule.Handler.RetryEvent)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:update")).Post("/integration-events/{id}/replay", integrationhubModule.Handler.ReplayDeadLetter)
			})
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:read")).Get("/integrations", integrationhubModule.Handler.ListIntegrations)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:read")).Get("/integration-events", integrationhubModule.Handler.ListEvents)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:read")).Get("/integration-events/dead-letter", integrationhubModule.Handler.ListDeadLetterEvents)
			})
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromURLParam("id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:read")).Get("/integrations/{id}", integrationhubModule.Handler.GetIntegration)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:read")).Get("/integrations/{id}/health", integrationhubModule.Handler.GetIntegrationHealth)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:update")).Put("/integrations/{id}", integrationhubModule.Handler.UpdateIntegration)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:delete")).Delete("/integrations/{id}", integrationhubModule.Handler.DeleteIntegration)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "integrations:update")).Post("/integrations/{id}/test", integrationhubModule.Handler.TestIntegration)
			})

			// Public incoming webhook endpoint (no auth; signature verified by provider).
			r.Post("/webhooks/{provider}/{integration_id}", integrationhubModule.Handler.ReceiveIncomingWebhook)

			// --- Billing ---
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromBody("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "billing:update")).Put("/billing/subscription", billingModule.Handler.UpdateSubscription)
			})
			r.Group(func(r chi.Router) {
				r.Use(sharedmiddleware.TenantContext(cfg.DB,
					sharedmiddleware.WorkspaceToOrg(sharedmiddleware.OrgIDFromQuery("workspace_id"), tenantResolver),
					tenantResolver,
				))
				r.Use(sharedmiddleware.IdempotencyKey(idempotencyStore, cfg.IdempotencyKeyTTL))
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "billing:read")).Get("/billing/subscription", billingModule.Handler.GetSubscription)
				r.With(sharedmiddleware.RequirePermission(rbacCfg, "billing:read")).Get("/billing/invoices", billingModule.Handler.ListInvoices)
			})
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		healthCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := cfg.DB.PingContext(healthCtx); err != nil {
			apihttp.ErrorJSON(w, http.StatusServiceUnavailable, "UNHEALTHY", "database unavailable")
			return
		}
		apihttp.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return r
}
