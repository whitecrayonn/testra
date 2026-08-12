package integrationhub

import (
	"time"

	"github.com/google/uuid"
)

type IntegrationType string

const (
	TypeJira        IntegrationType = "jira"
	TypeGitHub      IntegrationType = "github"
	TypeGitLab      IntegrationType = "gitlab"
	TypeBitbucket   IntegrationType = "bitbucket"
	TypeAzureDevOps IntegrationType = "azure_devops"
	TypeLinear      IntegrationType = "linear"
	TypeSlack       IntegrationType = "slack"
	TypeDiscord     IntegrationType = "discord"
	TypeSMTP        IntegrationType = "smtp"
	TypeWebhook     IntegrationType = "webhook"
)

func IsValidIntegrationType(s string) bool {
	switch IntegrationType(s) {
	case TypeJira, TypeGitHub, TypeGitLab, TypeBitbucket, TypeAzureDevOps, TypeLinear, TypeSlack, TypeDiscord, TypeSMTP, TypeWebhook:
		return true
	}
	return false
}

type Integration struct {
	ID           uuid.UUID
	WorkspaceID  uuid.UUID
	Type         IntegrationType
	Name         string
	Config       map[string]string
	Enabled      bool
	HealthStatus string // healthy, degraded, error, unknown
	LastTestedAt *time.Time
	LastError    string
	SyncStatus   string // synced, pending, error
	RetryCount   int
	CreatedBy    uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type IntegrationEvent struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	IntegrationID *uuid.UUID
	EventType     string
	Payload       map[string]interface{}
	Status        string
	ExternalID    string
	RetryCount    int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type DispatchInput struct {
	WorkspaceID uuid.UUID
	Integration *Integration
	EventType   string
	Payload     map[string]interface{}
}
