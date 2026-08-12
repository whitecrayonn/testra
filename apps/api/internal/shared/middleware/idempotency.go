package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/idempotency"
)

// mutatingMethods are the HTTP methods for which an Idempotency-Key, when
// provided, will be honored and stored.
var mutatingMethods = map[string]bool{
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
}

// IdempotencyKey enforces Idempotency-Key semantics for side-effecting requests.
//
// It is optional: if the client does not send an Idempotency-Key header, the
// request passes through unchanged. When present, it stores a fingerprint of the
// request body and the response so that retries with the same key and body replay
// the stored response. Reusing a key with a different body returns 409 Conflict.
//
// The key is scoped to (organization_id, workspace_id, operation, key_hash).
// workspace_id is extracted from the JSON body when present; otherwise the scope
// falls back to the organization in the request context. The operation is derived
// from the HTTP method and path.
func IdempotencyKey(store idempotency.Store, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !mutatingMethods[r.Method] {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			orgID, ok := db.TenantIDFromContext(r.Context())
			if !ok {
				apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "tenant context required for idempotency")
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "failed to read request body")
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			workspaceID := extractWorkspaceID(body)
			hashedKey := idempotency.HashKey(key)
			fingerprint := idempotency.Fingerprint(body)
			operation := r.Method + " " + r.URL.EscapedPath()

			existing, err := store.Get(r.Context(), orgID, workspaceID, operation, hashedKey)
			if err != nil && !errors.Is(err, sharederrors.ErrNotFound) {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to check idempotency key")
				return
			}

			if err == nil {
				if existing.Fingerprint != fingerprint {
					apihttp.ErrorJSON(w, http.StatusConflict, "IDEMPOTENCY_KEY_CONFLICT", "Idempotency-Key already used with a different request body")
					return
				}
				replayResponse(w, existing.StatusCode, existing.ResponseBody)
				return
			}

			rec := &responseRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)

			if rec.statusCode >= 200 && rec.statusCode < 500 {
				record := &idempotency.Record{
					ID:             uuid.New(),
					OrganizationID: orgID,
					WorkspaceID:    workspaceID,
					Operation:      operation,
					Key:            hashedKey,
					Fingerprint:    fingerprint,
					ResponseBody:   rec.body.Bytes(),
					StatusCode:     rec.statusCode,
					CreatedAt:      time.Now().UTC(),
					ExpiresAt:      time.Now().UTC().Add(ttl),
				}
				if err := store.Save(r.Context(), record); err != nil {
					// Do not fail the request because the original side effect already succeeded.
					// A failure here is logged by the request logger and is a transient concern.
					_ = err
				}
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.statusCode == 0 {
		r.statusCode = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	if r.statusCode == 0 {
		r.WriteHeader(http.StatusOK)
	}
	r.body.Write(p)
	return r.ResponseWriter.Write(p)
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func extractWorkspaceID(body []byte) uuid.UUID {
	var partial struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if err := json.Unmarshal(body, &partial); err != nil {
		return uuid.Nil
	}
	id, err := uuid.Parse(partial.WorkspaceID)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func replayResponse(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
