package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/secrets"
	"github.com/testra/testra/apps/api/internal/shared/security"
	"github.com/testra/testra/apps/api/internal/shared/validation"
)

type SMTPConfig struct {
	Host           string
	Port           string
	From           string
	Username       string
	Password       string // Deprecated: use SecretProvider + PasswordSecret
	SecretProvider secrets.Provider
	PasswordSecret string // Key passed to SecretProvider to retrieve the SMTP password
}

type Service struct {
	repo         Repository
	smtp         SMTPConfig
	httpClient   *http.Client
	smtpSender   func(string, smtp.Auth, string, []string, []byte) error
	urlValidator func(context.Context, string) error
}

func NewService(repo Repository, smtpCfg SMTPConfig) *Service {
	return &Service{
		repo:         repo,
		smtp:         smtpCfg,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		smtpSender:   smtp.SendMail,
		urlValidator: security.ValidateURL,
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

func (s *Service) ListChannels(ctx context.Context, workspaceID uuid.UUID, cursor string, limit int) ([]NotificationChannel, error) {
	if workspaceID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListChannels(ctx, workspaceID, cursor, limit)
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

	channels, err := s.repo.ListChannels(ctx, input.WorkspaceID, "", 1000)
	if err != nil {
		return notification, err
	}

	for _, ch := range channels {
		if !s.isChannelEnabled(ch.Type, prefs) {
			continue
		}

		history := &NotificationHistory{
			ID:             uuid.New(),
			OrganizationID: input.OrganizationID,
			ChannelID:      &ch.ID,
			ChannelType:    string(ch.Type),
			Status:         "pending",
			RetryCount:     0,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		if notification != nil {
			history.NotificationID = notification.ID
		}
		if err := s.repo.CreateHistory(ctx, history); err != nil {
			log.Printf("notification: failed to create history: %v", err)
			continue
		}

		attempts, err := s.deliver(ctx, ch, input)
		history.UpdatedAt = time.Now().UTC()
		history.RetryCount = attempts
		if err != nil {
			history.Status = "failed"
			history.ErrorMessage = err.Error()
			log.Printf("notification: failed to dispatch %s for channel %s: %v", ch.Type, ch.ID, err)
		} else {
			history.Status = "sent"
		}
		if updateErr := s.repo.UpdateHistory(ctx, history); updateErr != nil {
			log.Printf("notification: failed to update history: %v", updateErr)
		}
	}

	return notification, nil
}

func (s *Service) isChannelEnabled(t NotificationChannelType, prefs *NotificationPreferences) bool {
	switch t {
	case ChannelTypeEmail:
		return prefs.EmailEnabled
	case ChannelTypeSlack:
		return prefs.SlackEnabled
	case ChannelTypeTeams:
		return prefs.TeamsEnabled
	case ChannelTypeWebhook:
		return prefs.WebhookEnabled
	}
	return false
}

func (s *Service) deliver(ctx context.Context, ch NotificationChannel, input SendInput) (int, error) {
	switch ch.Type {
	case ChannelTypeEmail:
		return s.dispatchEmail(ctx, ch, input)
	case ChannelTypeSlack, ChannelTypeTeams, ChannelTypeWebhook:
		return s.dispatchHTTP(ctx, ch, input)
	}
	return 0, fmt.Errorf("unknown channel type %s", ch.Type)
}

func (s *Service) dispatchEmail(ctx context.Context, ch NotificationChannel, input SendInput) (int, error) {
	if s.smtp.Host == "" || s.smtp.From == "" {
		return 0, nil
	}
	to, ok := ch.Config["to"]
	if !ok || to == "" {
		return 0, nil
	}

	password := s.smtp.Password
	if s.smtp.SecretProvider != nil && s.smtp.PasswordSecret != "" {
		v, err := s.smtp.SecretProvider.Get(s.smtp.PasswordSecret)
		if err != nil {
			return 0, fmt.Errorf("resolve smtp password: %w", err)
		}
		password = v
	}

	var auth smtp.Auth
	if s.smtp.Username != "" && password != "" {
		auth = smtp.PlainAuth("", s.smtp.Username, password, s.smtp.Host)
	}

	subject := fmt.Sprintf("Subject: %s\r\n", input.Title)
	from := fmt.Sprintf("From: %s\r\n", s.smtp.From)
	recipients := strings.Split(to, ",")
	msg := []byte(from + "To: " + to + "\r\n" + subject + "\r\n" + input.Body + "\r\n")
	addr := s.smtp.Host + ":" + s.smtp.Port

	return s.retry(ctx, func() error {
		return s.smtpSender(addr, auth, s.smtp.From, recipients, msg)
	})
}

func (s *Service) dispatchHTTP(ctx context.Context, ch NotificationChannel, input SendInput) (int, error) {
	url := ch.Config["url"]
	if url == "" {
		return 0, nil
	}
	if s.urlValidator != nil {
		if err := s.urlValidator(ctx, url); err != nil {
			return 0, fmt.Errorf("blocked by SSRF guard: %w", err)
		}
	}
	payload := map[string]string{
		"title": input.Title,
		"body":  input.Body,
		"link":  input.Link,
		"type":  input.Type,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	return s.retry(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return nil
	})
}

// retry attempts f up to maxAttempts times with exponential backoff. The first
// attempt is immediate, subsequent attempts wait 500ms, 1s, 2s, ... up to a
// 5 second cap. The context is checked before every attempt.
// It returns the number of attempts made and the final error, if any.
func (s *Service) retry(ctx context.Context, f func() error) (int, error) {
	const maxAttempts = 3
	delay := 500 * time.Millisecond
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := ctx.Err(); err != nil {
			return i + 1, err
		}
		if err := f(); err == nil {
			return i + 1, nil
		} else {
			lastErr = err
		}
		if i < maxAttempts-1 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return i + 1, ctx.Err()
			case <-timer.C:
			}
			if delay < 5*time.Second {
				delay *= 2
			}
		}
	}
	return maxAttempts, lastErr
}

func (s *Service) CreateTemplate(ctx context.Context, input CreateTemplateInput, createdBy uuid.UUID) (*NotificationTemplate, error) {
	if input.OrganizationID == uuid.Nil || !validation.IsValidName(input.Name) || input.EventType == "" || input.ChannelType == "" {
		return nil, sharederrors.ErrInvalidInput
	}
	now := time.Now().UTC()
	t := &NotificationTemplate{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		Name:           strings.TrimSpace(input.Name),
		EventType:      input.EventType,
		ChannelType:    input.ChannelType,
		Subject:        strings.TrimSpace(input.Subject),
		Body:           strings.TrimSpace(input.Body),
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.CreateTemplate(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) GetTemplate(ctx context.Context, id uuid.UUID) (*NotificationTemplate, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.GetTemplate(ctx, id)
}

func (s *Service) ListTemplates(ctx context.Context, orgID uuid.UUID, eventType, channelType string, limit int) ([]NotificationTemplate, error) {
	if orgID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListTemplates(ctx, orgID, eventType, channelType, limit)
}

func (s *Service) UpdateTemplate(ctx context.Context, id uuid.UUID, input UpdateTemplateInput) (*NotificationTemplate, error) {
	if id == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	t, err := s.repo.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		t.Name = input.Name
	}
	if input.EventType != "" {
		t.EventType = input.EventType
	}
	if input.ChannelType != "" {
		t.ChannelType = input.ChannelType
	}
	if input.Subject != "" {
		t.Subject = input.Subject
	}
	if input.Body != "" {
		t.Body = input.Body
	}
	t.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateTemplate(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return sharederrors.ErrInvalidInput
	}
	return s.repo.DeleteTemplate(ctx, id)
}

func (s *Service) ListHistory(ctx context.Context, notificationID uuid.UUID, limit int) ([]NotificationHistory, error) {
	if notificationID == uuid.Nil {
		return nil, sharederrors.ErrInvalidInput
	}
	return s.repo.ListHistory(ctx, notificationID, limit)
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

type CreateTemplateInput struct {
	OrganizationID uuid.UUID
	Name           string
	EventType      string
	ChannelType    string
	Subject        string
	Body           string
}

type UpdateTemplateInput struct {
	Name        string
	EventType   string
	ChannelType string
	Subject     string
	Body        string
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
