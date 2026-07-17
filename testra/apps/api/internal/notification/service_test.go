package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type fakeRepo struct {
	notifications []Notification
	prefs         map[string]*NotificationPreferences
	channels      map[uuid.UUID]*NotificationChannel
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		prefs:    make(map[string]*NotificationPreferences),
		channels: make(map[uuid.UUID]*NotificationChannel),
	}
}

func (f *fakeRepo) CreateNotification(_ context.Context, n *Notification) error {
	f.notifications = append(f.notifications, *n)
	return nil
}

func (f *fakeRepo) GetNotification(_ context.Context, id, userID uuid.UUID) (*Notification, error) {
	for _, n := range f.notifications {
		if n.ID == id && n.UserID == userID {
			return &n, nil
		}
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepo) ListNotifications(_ context.Context, userID uuid.UUID, read *bool, _ string, _ int) ([]Notification, string, error) {
	var out []Notification
	for i := len(f.notifications) - 1; i >= 0; i-- {
		n := f.notifications[i]
		if n.UserID != userID {
			continue
		}
		if read != nil && n.Read != *read {
			continue
		}
		out = append(out, n)
	}
	return out, "", nil
}

func (f *fakeRepo) CountUnread(_ context.Context, userID uuid.UUID) (int, error) {
	count := 0
	for _, n := range f.notifications {
		if n.UserID == userID && !n.Read {
			count++
		}
	}
	return count, nil
}

func (f *fakeRepo) MarkRead(_ context.Context, id, userID uuid.UUID, read bool) error {
	for i, n := range f.notifications {
		if n.ID == id && n.UserID == userID {
			f.notifications[i].Read = read
			return nil
		}
	}
	return sharederrors.ErrNotFound
}

func (f *fakeRepo) DeleteNotification(_ context.Context, id, userID uuid.UUID) error {
	for i, n := range f.notifications {
		if n.ID == id && n.UserID == userID {
			f.notifications = append(f.notifications[:i], f.notifications[i+1:]...)
			return nil
		}
	}
	return sharederrors.ErrNotFound
}

func (f *fakeRepo) GetPreferences(_ context.Context, orgID, userID uuid.UUID) (*NotificationPreferences, error) {
	key := orgID.String() + ":" + userID.String()
	if p, ok := f.prefs[key]; ok {
		return p, nil
	}
	return &NotificationPreferences{
		OrganizationID: orgID,
		UserID:         userID,
		InAppEnabled:   true,
	}, nil
}

func (f *fakeRepo) UpsertPreferences(_ context.Context, p *NotificationPreferences) error {
	key := p.OrganizationID.String() + ":" + p.UserID.String()
	f.prefs[key] = p
	return nil
}

func (f *fakeRepo) CreateChannel(_ context.Context, ch *NotificationChannel) error {
	f.channels[ch.ID] = ch
	return nil
}

func (f *fakeRepo) GetChannel(_ context.Context, id uuid.UUID) (*NotificationChannel, error) {
	if ch, ok := f.channels[id]; ok {
		return ch, nil
	}
	return nil, sharederrors.ErrNotFound
}

func (f *fakeRepo) ListChannels(_ context.Context, workspaceID uuid.UUID) ([]NotificationChannel, error) {
	var out []NotificationChannel
	for _, ch := range f.channels {
		if ch.WorkspaceID == workspaceID {
			out = append(out, *ch)
		}
	}
	return out, nil
}

func (f *fakeRepo) UpdateChannel(_ context.Context, ch *NotificationChannel) error {
	if _, ok := f.channels[ch.ID]; !ok {
		return sharederrors.ErrNotFound
	}
	f.channels[ch.ID] = ch
	return nil
}

func (f *fakeRepo) DeleteChannel(_ context.Context, id uuid.UUID) error {
	if _, ok := f.channels[id]; !ok {
		return sharederrors.ErrNotFound
	}
	delete(f.channels, id)
	return nil
}

func newTestService() *Service {
	return NewService(newFakeRepo(), SMTPConfig{})
}

func TestCreateNotification(t *testing.T) {
	svc := newTestService()
	orgID := uuid.New()
	userID := uuid.New()

	n, err := svc.CreateNotification(context.Background(), CreateNotificationInput{
		OrganizationID: orgID,
		UserID:         userID,
		Type:           "system",
		Title:          "Test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.ID == uuid.Nil {
		t.Error("expected notification id to be set")
	}
	if n.Read {
		t.Error("new notification should be unread")
	}

	_, err = svc.CreateNotification(context.Background(), CreateNotificationInput{
		OrganizationID: orgID,
		UserID:         userID,
		Type:           "bad",
		Title:          "Bad",
	})
	if err != sharederrors.ErrInvalidInput {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestMarkReadAndDelete(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()

	n, _ := svc.CreateNotification(context.Background(), CreateNotificationInput{
		OrganizationID: uuid.New(),
		UserID:         userID,
		Type:           "system",
		Title:          "Mark me",
	})

	if err := svc.MarkRead(context.Background(), n.ID, userID, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := svc.GetNotification(context.Background(), n.ID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Read {
		t.Error("expected notification to be read")
	}

	if err := svc.DeleteNotification(context.Background(), n.ID, userID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetNotification(context.Background(), n.ID, userID)
	if err != sharederrors.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestPreferences(t *testing.T) {
	svc := newTestService()
	orgID := uuid.New()
	userID := uuid.New()

	updated, err := svc.UpdatePreferences(context.Background(), UpdatePreferencesInput{
		OrganizationID: orgID,
		EmailEnabled:   true,
		SlackEnabled:   true,
	}, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated.EmailEnabled || !updated.SlackEnabled {
		t.Error("expected preferences to be updated")
	}

	got, err := svc.GetPreferences(context.Background(), orgID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("expected user id %s, got %s", userID, got.UserID)
	}
}

func TestChannels(t *testing.T) {
	svc := newTestService()
	orgID := uuid.New()
	wsID := uuid.New()
	createdBy := uuid.New()

	ch, err := svc.CreateChannel(context.Background(), CreateChannelInput{
		OrganizationID: orgID,
		WorkspaceID:    wsID,
		Type:           "email",
		Name:           "Team email",
		Config:         map[string]string{"to": "team@example.com"},
	}, createdBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	channels, err := svc.ListChannels(context.Background(), wsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}

	updated, err := svc.UpdateChannel(context.Background(), ch.ID, UpdateChannelInput{
		Name: "Updated name",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Updated name" {
		t.Errorf("expected name Updated name, got %s", updated.Name)
	}

	if err := svc.DeleteChannel(context.Background(), ch.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetChannel(context.Background(), ch.ID)
	if err != sharederrors.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestChannelConfigValidation(t *testing.T) {
	svc := newTestService()
	_, err := svc.CreateChannel(context.Background(), CreateChannelInput{
		OrganizationID: uuid.New(),
		WorkspaceID:    uuid.New(),
		Type:           "email",
		Name:           "Missing to",
		Config:         map[string]string{},
	}, uuid.New())
	if err == nil {
		t.Fatal("expected error for missing 'to' config")
	}

	_, err = svc.CreateChannel(context.Background(), CreateChannelInput{
		OrganizationID: uuid.New(),
		WorkspaceID:    uuid.New(),
		Type:           "slack",
		Name:           "Missing url",
		Config:         map[string]string{},
	}, uuid.New())
	if err == nil {
		t.Fatal("expected error for missing 'url' config")
	}
}

func TestListNotifications(t *testing.T) {
	svc := newTestService()
	userID := uuid.New()

	for i := 0; i < 3; i++ {
		_, err := svc.CreateNotification(context.Background(), CreateNotificationInput{
			OrganizationID: uuid.New(),
			UserID:         userID,
			Type:           "system",
			Title:          "n",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		time.Sleep(time.Millisecond)
	}

	list, _, err := svc.ListNotifications(context.Background(), userID, nil, "", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 notifications, got %d", len(list))
	}
}
