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
			header := r.Header.Get("Authorization")
			if header == "" {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
				return
			}

			claims, err := jwt.Parse(parts[1], cfg.JWTSecret)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
				return
			}

			ctx := WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
