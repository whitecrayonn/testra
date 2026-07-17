package notification

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Notifications
	CreateNotification(ctx context.Context, n *Notification) error
	GetNotification(ctx context.Context, id, userID uuid.UUID) (*Notification, error)
	ListNotifications(ctx context.Context, userID uuid.UUID, read *bool, cursor string, limit int) ([]Notification, string, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID, read bool) error
	DeleteNotification(ctx context.Context, id, userID uuid.UUID) error

	// Preferences
	GetPreferences(ctx context.Context, orgID, userID uuid.UUID) (*NotificationPreferences, error)
	UpsertPreferences(ctx context.Context, p *NotificationPreferences) error

	// Channels
	CreateChannel(ctx context.Context, ch *NotificationChannel) error
	GetChannel(ctx context.Context, id uuid.UUID) (*NotificationChannel, error)
	ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]NotificationChannel, error)
	UpdateChannel(ctx context.Context, ch *NotificationChannel) error
	DeleteChannel(ctx context.Context, id uuid.UUID) error
}
