package audit

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
	"github.com/testra/testra/apps/api/internal/shared/middleware"
	"github.com/testra/testra/apps/api/internal/shared/pagination"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type eventResponse struct {
	ID         uuid.UUID         `json:"id"`
	Action     string            `json:"action"`
	Resource   string            `json:"resource"`
	ResourceID string            `json:"resource_id,omitempty"`
	IPAddress  string            `json:"ip_address,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

func mapEventResponse(e *Event) eventResponse {
	return eventResponse{
		ID:         e.ID,
		Action:     e.Action,
		Resource:   e.Resource,
		ResourceID: e.ResourceID,
		IPAddress:  e.IPAddress,
		Metadata:   e.Metadata,
		CreatedAt:  e.CreatedAt,
	}
}

// auditCursor is a composite (created_at, id) cursor, since audit events are
// ordered chronologically rather than by id. Encoded separately from the
// shared pagination package's single-id cursor helper because that helper
// only supports one field.
type auditCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

func encodeAuditCursor(e Event) (string, error) {
	b, err := json.Marshal(auditCursor{CreatedAt: e.CreatedAt, ID: e.ID})
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func decodeAuditCursor(cursor string) (*auditCursor, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	var c auditCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// List returns a page of the authenticated user's own audit trail.
//
// Scoped to the current user rather than the whole organization/workspace:
// audit_events has no organization_id column yet (SBL-080 in the backlog), so
// listing "everyone's" events here would leak other tenants' activity. Once
// that column exists, this can grow an org-wide view gated behind an
// audit:read-style permission for owners/admins.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
		return
	}

	limit := pagination.DefaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			if n > pagination.MaxLimit {
				n = pagination.MaxLimit
			}
			limit = n
		}
	}

	var beforeCreatedAt *time.Time
	var beforeID uuid.UUID
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		c, err := decodeAuditCursor(cursor)
		if err != nil {
			apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid cursor")
			return
		}
		beforeCreatedAt = &c.CreatedAt
		beforeID = c.ID
	}

	events, err := h.service.ListByUser(r.Context(), userID, beforeCreatedAt, beforeID, limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}

	resp := make([]eventResponse, len(events))
	for i, e := range events {
		resp[i] = mapEventResponse(&e)
	}

	meta := pagination.Meta{HasMore: len(events) == limit}
	if meta.HasMore && len(events) > 0 {
		nextCursor, err := encodeAuditCursor(events[len(events)-1])
		if err == nil {
			meta.NextCursor = nextCursor
		}
	}

	apihttp.JSONWithMeta(w, http.StatusOK, resp, meta)
}
