package integrationhub

import (
	"time"

	"github.com/google/uuid"
)

type IntegrationType string

const (
	TypeJira    IntegrationType = "jira"
	TypeGitHub  IntegrationType = "github"
	TypeGitLab  IntegrationType = "gitlab"
	TypeSlack   IntegrationType = "slack"
	TypeWebhook IntegrationType = "webhook"
)

func IsValidIntegrationType(s string) bool {
	switch IntegrationType(s) {
	case TypeJira, TypeGitHub, TypeGitLab, TypeSlack, TypeWebhook:
		return true
	}
	return false
}

type Integration struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Type        IntegrationType
	Name        string
	Config      map[string]string
	Enabled     bool
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type IntegrationEvent struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	IntegrationID *uuid.UUID
	EventType     string
	Payload       map[string]interface{}
	Status        string
	ExternalID    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type DispatchInput struct {
	WorkspaceID uuid.UUID
	Integration *Integration
	EventType   string
	Payload     map[string]interface{}
}
