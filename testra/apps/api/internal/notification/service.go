package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type SMTPConfig struct {
	Host     string
	Port     string
	From     string
	Password string
}

type Service struct {
	repo      Repository
	smtp      SMTPConfig
	httpClient *http.Client
}

func NewService(repo Repository, smtpCfg SMTPConfig) *Service {
	return &Service{
		repo:       repo,
		smtp:       smtpCfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Service) CreateNotification(ctx context.Context, input CreateNotificationInput) (*Notification, error) {
	if input.OrganizationID == uuid.Nil || input.UserID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if !validation.IsValidName(input.Title) || !IsValidNotificationType(input.Type) {
		return nil, sharederrors.ErrInvalidInput
	}

	n := &Notification{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		UserID:         input.UserID,
		Type:           NotificationType(input.Type),
		Title:          strings.TrimSpace(input.Title),
		Body:           strings.TrimSpace(input.Body),
		Link:           strings.TrimSpace(input.Link),
		Read:           false,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := s.repo.CreateNotification(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Service) GetNotification(ctx context.Context, id, userID uuid.UUID) (*Notification, error) {
	if id == uuid.Nil || userID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetNotification(ctx, id, userID)
}

func (s *Service) ListNotifications(ctx context.Context, userID uuid.UUID, read *bool, cursor string, limit int) ([]Notification, string, error) {
	if userID == uuid.Nil {
		return nil, "", sharederrors.ErrInvalidInput
	}
	return s.repo.ListNotifications(ctx, userID, read, cursor, limit)
}

func (s *Service) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	if userID == uuid.Nil {
		return 0, sharederrors.ErrInvalidInput
	}
	return s.repo.CountUnread(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, id, userID uuid.UUID, read bool) error {
	if id == uuid.Nil || userID == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.MarkRead(ctx, id, userID, read)
}

func (s *Service) DeleteNotification(ctx context.Context, id, userID uuid.UUID) error {
	if id == uuid.Nil || userID == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteNotification(ctx, id, userID)
}

func (s *Service) GetPreferences(ctx context.Context, orgID, userID uuid.UUID) (*NotificationPreferences, error) {
	if orgID == uuid.Nil || userID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetPreferences(ctx, orgID, userID)
}

func (s *Service) UpdatePreferences(ctx context.Context, input UpdatePreferencesInput, userID uuid.UUID) (*NotificationPreferences, error) {
	if input.OrganizationID == uuid.Nil || userID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	p := &NotificationPreferences{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		UserID:         userID,
		InAppEnabled:   input.InAppEnabled,
		EmailEnabled:   input.EmailEnabled,
		SlackEnabled:   input.SlackEnabled,
		TeamsEnabled:   input.TeamsEnabled,
		WebhookEnabled: input.WebhookEnabled,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := s.repo.UpsertPreferences(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) CreateChannel(ctx context.Context, input CreateChannelInput, createdBy uuid.UUID) (*NotificationChannel, error) {
	if input.OrganizationID == uuid.Nil || input.WorkspaceID == uuid.Nil || createdBy == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if !validation.IsValidName(input.Name) || !IsValidChannelType(input.Type) {
		return nil, sharederrors.ErrInvalidInput
	}
	if err := validateChannelConfig(NotificationChannelType(input.Type), input.Config); err != nil {
		return nil, err
	}

	ch := &NotificationChannel{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		WorkspaceID:    input.WorkspaceID,
		Type:           NotificationChannelType(input.Type),
		Name:           strings.TrimSpace(input.Name),
		Config:         input.Config,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := s.repo.CreateChannel(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *Service) GetChannel(ctx context.Context, id uuid.UUID) (*NotificationChannel, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetChannel(ctx, id)
}

func (s *Service) ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]NotificationChannel, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListChannels(ctx, workspaceID)
}

func (s *Service) UpdateChannel(ctx context.Context, id uuid.UUID, input UpdateChannelInput) (*NotificationChannel, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	ch, err := s.repo.GetChannel(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validation.IsValidName(input.Name) {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.Type != "" && !IsValidChannelType(input.Type) {
		return nil, sharederrors.ErrInvalidInput
	}
	if input.Type != "" {
		ch.Type = NotificationChannelType(input.Type)
	}
	ch.Name = strings.TrimSpace(input.Name)
	if input.Config != nil {
		ch.Config = input.Config
	}
	if err := validateChannelConfig(ch.Type, ch.Config); err != nil {
		return nil, err
	}
	ch.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateChannel(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *Service) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteChannel(ctx, id)
}

// Send creates an in-app notification and dispatches it to enabled channels.
func (s *Service) Send(ctx context.Context, input SendInput) (*Notification, error) {
	if input.OrganizationID == uuid.Nil || input.UserID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	if !validation.IsValidName(input.Title) || !IsValidNotificationType(input.Type) {
		return nil, sharederrors.ErrInvalidInput
	}

	prefs, err := s.repo.GetPreferences(ctx, input.OrganizationID, input.UserID)
	if err != nil {
		return nil, err
	}

	var notification *Notification
	if prefs.InAppEnabled {
		n := &Notification{
			ID:             uuid.New(),
			OrganizationID: input.OrganizationID,
			UserID:         input.UserID,
			Type:           NotificationType(input.Type),
			Title:          strings.TrimSpace(input.Title),
			Body:           strings.TrimSpace(input.Body),
			Link:           strings.TrimSpace(input.Link),
			Read:           false,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		if err := s.repo.CreateNotification(ctx, n); err != nil {
			return nil, err
		}
		notification = n
	}

	if input.WorkspaceID == uuid.Nil {
		return notification, nil
	}

	channels, err := s.repo.ListChannels(ctx, input.WorkspaceID)
	if err != nil {
		return notification, err
	}

	for _, ch := range channels {
		switch ch.Type {
		case ChannelTypeEmail:
			if prefs.EmailEnabled {
				s.dispatchEmail(ctx, ch, input)
			}
		case ChannelTypeSlack, ChannelTypeTeams, ChannelTypeWebhook:
			if (ch.Type == ChannelTypeSlack && prefs.SlackEnabled) ||
				(ch.Type == ChannelTypeTeams && prefs.TeamsEnabled) ||
				(ch.Type == ChannelTypeWebhook && prefs.WebhookEnabled) {
				s.dispatchHTTP(ctx, ch, input)
			}
		}
	}

	return notification, nil
}

func (s *Service) dispatchEmail(ctx context.Context, ch NotificationChannel, input SendInput) {
	if s.smtp.Host == "" || s.smtp.From == "" {
		return
	}
	to, ok := ch.Config["to"]
	if !ok || to == "" {
		return
	}
	subject := fmt.Sprintf("Subject: %s\r\n", input.Title)
	from := fmt.Sprintf("From: %s\r\n", s.smtp.From)
	recipients := strings.Split(to, ",")
	msg := []byte(from + "To: " + to + "\r\n" + subject + "\r\n" + input.Body + "\r\n")
	addr := s.smtp.Host + ":" + s.smtp.Port
	_ = smtp.SendMail(addr, nil, s.smtp.From, recipients, msg)
}

func (s *Service) dispatchHTTP(ctx context.Context, ch NotificationChannel, input SendInput) {
	url := ch.Config["url"]
	if url == "" {
		return
	}
	payload := map[string]string{
		"title": input.Title,
		"body":  input.Body,
		"link":  input.Link,
		"type":  input.Type,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
}

func validateChannelConfig(t NotificationChannelType, config map[string]string) error {
	switch t {
	case ChannelTypeEmail:
		if config["to"] == "" {
			return fmt.Errorf("%w: email channel requires a 'to' address", sharederrors.ErrInvalidInput)
		}
	case ChannelTypeSlack, ChannelTypeTeams, ChannelTypeWebhook:
		if config["url"] == "" {
			return fmt.Errorf("%w: %s channel requires a 'url'", sharederrors.ErrInvalidInput, t)
		}
	default:
		return sharederrors.ErrInvalidInput
	}
	return nil
}

// Input structs

type CreateNotificationInput struct {
	OrganizationID uuid.UUID
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	Type           string
	Title          string
	Body           string
	Link           string
}

type UpdatePreferencesInput struct {
	OrganizationID uuid.UUID
	InAppEnabled   bool
	EmailEnabled   bool
	SlackEnabled   bool
	TeamsEnabled   bool
	WebhookEnabled bool
}

type CreateChannelInput struct {
	OrganizationID uuid.UUID
	WorkspaceID    uuid.UUID
	Type           string
	Name           string
	Config         map[string]string
}

type UpdateChannelInput struct {
	Type   string
	Name   string
	Config map[string]string
}

type SendInput struct {
	OrganizationID uuid.UUID
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	Type           string
	Title          string
	Body           string
	Link           string
}
