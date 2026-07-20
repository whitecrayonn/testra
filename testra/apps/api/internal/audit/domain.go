package audit

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Action     string
	Resource   string
	ResourceID string
	IPAddress  string
	Metadata   map[string]string
	CreatedAt  time.Time
}
