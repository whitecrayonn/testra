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
	ListChannels(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]NotificationChannel, error)
	UpdateChannel(ctx context.Context, ch *NotificationChannel) error
	DeleteChannel(ctx context.Context, id uuid.UUID) error

	// Templates
	CreateTemplate(ctx context.Context, t *NotificationTemplate) error
	GetTemplate(ctx context.Context, id uuid.UUID) (*NotificationTemplate, error)
	ListTemplates(ctx context.Context, orgID uuid.UUID, eventType, channelType string, limit int) ([]NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, t *NotificationTemplate) error
	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	// History
	CreateHistory(ctx context.Context, h *NotificationHistory) error
	UpdateHistory(ctx context.Context, h *NotificationHistory) error
	ListHistory(ctx context.Context, notificationID uuid.UUID, limit int) ([]NotificationHistory, error)
}
