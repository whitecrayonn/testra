package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
)

type Repository interface {
	Insert(ctx context.Context, event *Event) error

	// ListByUser returns a page of the given user's own audit events, most
	// recent first. Scoped to a single user rather than an organization
	// because audit_events currently has no organization_id column (tracked
	// as SBL-080); a workspace/org-wide view needs that column added first
	// to avoid leaking other tenants' events.
	ListByUser(ctx context.Context, userID uuid.UUID, beforeCreatedAt *time.Time, beforeID uuid.UUID, limit int) ([]Event, error)
}

type SQLRepository struct {
	db db.DBTX
}

func NewSQLRepository(db db.DBTX) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Insert(ctx context.Context, event *Event) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	var metadataJSON []byte
	if event.Metadata != nil {
		metadataJSON, _ = json.Marshal(event.Metadata)
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_events (id, user_id, action, resource, resource_id, ip_address, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.ID, event.UserID, event.Action, event.Resource, event.ResourceID, event.IPAddress, metadataJSON, event.CreatedAt,
	)
	return err
}

func (r *SQLRepository) ListByUser(ctx context.Context, userID uuid.UUID, beforeCreatedAt *time.Time, beforeID uuid.UUID, limit int) ([]Event, error) {
	var rows *sql.Rows
	var err error

	if beforeCreatedAt != nil {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, user_id, action, resource, resource_id, ip_address, metadata, created_at
			 FROM audit_events
			 WHERE user_id = $1 AND (created_at, id) < ($2, $3)
			 ORDER BY created_at DESC, id DESC
			 LIMIT $4`,
			userID, *beforeCreatedAt, beforeID, limit,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, user_id, action, resource, resource_id, ip_address, metadata, created_at
			 FROM audit_events
			 WHERE user_id = $1
			 ORDER BY created_at DESC, id DESC
			 LIMIT $2`,
			userID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]Event, 0, limit)
	for rows.Next() {
		var e Event
		var resourceID sql.NullString
		var ipAddress sql.NullString
		var metadataJSON []byte
		var uid sql.NullString
		if err := rows.Scan(&e.ID, &uid, &e.Action, &e.Resource, &resourceID, &ipAddress, &metadataJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		if uid.Valid {
			if parsed, err := uuid.Parse(uid.String); err == nil {
				e.UserID = parsed
			}
		}
		e.ResourceID = resourceID.String
		e.IPAddress = ipAddress.String
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &e.Metadata)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
