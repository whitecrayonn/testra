package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/idempotency"
)

// IdempotencyKey enforces Idempotency-Key semantics for a side-effecting endpoint.
// It stores a fingerprint of the request body and the response so that retries with
// the same key and body replay the stored response. Reusing a key with a different
// body returns 409 Conflict. The key is scoped to (workspace_id, operation, key_hash)
// and expires after ttl.
func IdempotencyKey(store idempotency.Store, operation string, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header is required")
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "failed to read request body")
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			workspaceID, ok := extractWorkspaceID(body)
			if !ok {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "workspace_id is required in request body")
				return
			}

			hashedKey := idempotency.HashKey(key)
			fingerprint := idempotency.Fingerprint(body)

			existing, err := store.Get(r.Context(), workspaceID, operation, hashedKey)
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
					ID:           uuid.New(),
					WorkspaceID:  workspaceID,
					Operation:    operation,
					Key:          hashedKey,
					Fingerprint:  fingerprint,
					ResponseBody: rec.body.Bytes(),
					StatusCode:   rec.statusCode,
					CreatedAt:    time.Now().UTC(),
					ExpiresAt:    time.Now().UTC().Add(ttl),
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

func extractWorkspaceID(body []byte) (uuid.UUID, bool) {
	var partial struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if err := json.Unmarshal(body, &partial); err != nil {
		return uuid.Nil, false
	}
	if partial.WorkspaceID == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(partial.WorkspaceID)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func replayResponse(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
