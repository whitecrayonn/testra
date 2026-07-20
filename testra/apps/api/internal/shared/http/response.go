package http

import (
	"encoding/json"
	"errors"
	"net/http"

	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type Envelope struct {
	Data  any    `json:"data"`
	Meta  any    `json:"meta,omitempty"`
	Error *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data})
}

func ErrorJSON(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Error: &Error{Code: code, Message: message}})
}

// MapError maps sentinel errors from the shared errors package to a JSON
// HTTP error response. It uses errors.Is so wrapped errors are still matched.
func MapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, sharederrors.ErrConflict):
		ErrorJSON(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, sharederrors.ErrNotFound):
		ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, sharederrors.ErrInvalidInput):
		ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case errors.Is(err, sharederrors.ErrForbidden):
		ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, sharederrors.ErrUnauthorized):
		ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, sharederrors.ErrInvalidCredential):
		ErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
	case errors.Is(err, sharederrors.ErrMFARequired):
		ErrorJSON(w, http.StatusUnauthorized, "MFA_REQUIRED", err.Error())
	case errors.Is(err, sharederrors.ErrTokenExpired):
		ErrorJSON(w, http.StatusUnauthorized, "TOKEN_EXPIRED", err.Error())
	case errors.Is(err, sharederrors.ErrTokenRevoked):
		ErrorJSON(w, http.StatusUnauthorized, "TOKEN_REVOKED", err.Error())
	case errors.Is(err, sharederrors.ErrTooManyRequests):
		ErrorJSON(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", err.Error())
	default:
		ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
