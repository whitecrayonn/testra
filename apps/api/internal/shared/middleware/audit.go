package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

type AuditLogInput struct {
	UserID     uuid.UUID
	Action     string
	Resource   string
	ResourceID string
	IPAddress  string
	StatusCode int
}

type statusCaptureWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCaptureWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func AuditLog(
	action, resource string,
	extractUserID func(r *http.Request) uuid.UUID,
	extractResourceID func(r *http.Request) string,
	logger func(input AuditLogInput),
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := &statusCaptureWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)

			uid := extractUserID(r)
			if uid != uuid.Nil {
				logger(AuditLogInput{
					UserID:     uid,
					Action:     action,
					Resource:   resource,
					ResourceID: extractResourceID(r),
					IPAddress:  MaskIP(r.RemoteAddr),
					StatusCode: ww.status,
				})
			}
		})
	}
}
