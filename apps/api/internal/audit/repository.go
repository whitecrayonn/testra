package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
)

type Repository interface {
	Insert(ctx context.Context, event *Event) error
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
