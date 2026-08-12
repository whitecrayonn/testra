package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/testra/testra/apps/api/internal/shared/errors"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
)

type AuthConfig struct {
	TokenManager *jwt.Manager

	// IsRevoked, if set, is called with the access token's jti (claims.ID) on
	// every authenticated request to check a server-side denylist (e.g. tokens
	// revoked by logout before their natural expiry). Optional: a nil value
	// skips the check entirely, so existing tests and callers that don't wire
	// a denylist keep working unchanged.
	IsRevoked func(ctx context.Context, jti string) (bool, error)
}

func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
				return
			}

			if cfg.TokenManager == nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL", errors.ErrInternal.Error())
				return
			}

			claims, err := cfg.TokenManager.Parse(token)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
				return
			}

			if cfg.IsRevoked != nil {
				revoked, err := cfg.IsRevoked(r.Context(), claims.ID)
				if err != nil {
					apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL", errors.ErrInternal.Error())
					return
				}
				if revoked {
					apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
					return
				}
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

	if token, ok := AccessTokenFromCookie(r); ok {
		return token
	}

	return ""
}
