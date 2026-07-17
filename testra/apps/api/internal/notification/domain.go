package notification

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeSystem           NotificationType = "system"
	NotificationTypeTestRunCompleted NotificationType = "test_run_completed"
	NotificationTypeDefectAssigned   NotificationType = "defect_assigned"
	NotificationTypeMention          NotificationType = "mention"
)

func IsValidNotificationType(s string) bool {
	switch NotificationType(s) {
	case NotificationTypeSystem, NotificationTypeTestRunCompleted, NotificationTypeDefectAssigned, NotificationTypeMention:
		return true
	}
	return false
}

type Notification struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	Type           NotificationType
	Title          string
	Body           string
	Link           string
	Read           bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type NotificationPreferences struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	UserID          uuid.UUID
	InAppEnabled    bool
	EmailEnabled    bool
	SlackEnabled    bool
	TeamsEnabled    bool
	WebhookEnabled  bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NotificationChannelType string

const (
	ChannelTypeEmail   NotificationChannelType = "email"
	ChannelTypeSlack   NotificationChannelType = "slack"
	ChannelTypeTeams   NotificationChannelType = "teams"
	ChannelTypeWebhook NotificationChannelType = "webhook"
)

func IsValidChannelType(s string) bool {
	switch NotificationChannelType(s) {
	case ChannelTypeEmail, ChannelTypeSlack, ChannelTypeTeams, ChannelTypeWebhook:
		return true
	}
	return false
}

type NotificationChannel struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	WorkspaceID    uuid.UUID
	Type           NotificationChannelType
	Name           string
	Config         map[string]string
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
