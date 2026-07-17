package middleware

import (
	"net/http"
	"strings"

	"github.com/testra/testra/apps/api/internal/shared/errors"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
)

type AuthConfig struct {
	JWTSecret string
}

func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
				return
			}

			claims, err := jwt.Parse(token, cfg.JWTSecret)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
				return
			}

			ctx := WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// EventSource cannot send custom headers, so clients may pass a short-lived
	// JWT via the access_token query parameter for SSE endpoints.
	if token := r.URL.Query().Get("access_token"); token != "" {
		return token
	}
	return r.URL.Query().Get("token")
}
