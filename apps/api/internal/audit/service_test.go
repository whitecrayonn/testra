package audit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepo struct {
	events []Event
}

func (r *fakeRepo) Insert(ctx context.Context, event *Event) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	r.events = append(r.events, *event)
	return nil
}

// ListByUser mirrors the SQL implementation's ordering/filter semantics
// (WHERE user_id = ..., ORDER BY created_at DESC, id DESC) closely enough to
// exercise the service/handler pagination logic without a real database.
func (r *fakeRepo) ListByUser(ctx context.Context, userID uuid.UUID, beforeCreatedAt *time.Time, beforeID uuid.UUID, limit int) ([]Event, error) {
	var matched []Event
	for _, e := range r.events {
		if e.UserID != userID {
			continue
		}
		if beforeCreatedAt != nil {
			if e.CreatedAt.After(*beforeCreatedAt) {
				continue
			}
			if e.CreatedAt.Equal(*beforeCreatedAt) && e.ID.String() >= beforeID.String() {
				continue
			}
		}
		matched = append(matched, e)
	}
	// Sort desc by created_at, then id, matching `ORDER BY created_at DESC, id DESC`.
	for i := 0; i < len(matched); i++ {
		for j := i + 1; j < len(matched); j++ {
			if matched[j].CreatedAt.After(matched[i].CreatedAt) ||
				(matched[j].CreatedAt.Equal(matched[i].CreatedAt) && matched[j].ID.String() > matched[i].ID.String()) {
				matched[i], matched[j] = matched[j], matched[i]
			}
		}
	}
	if len(matched) > limit {
		matched = matched[:limit]
	}
	return matched, nil
}

func TestService_Log_PersistsEvent(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	userID := uuid.New()

	svc.Log(context.Background(), LogInput{
		UserID:   userID,
		Action:   "login",
		Resource: "session",
	})

	if len(repo.events) != 1 {
		t.Fatalf("expected 1 event persisted, got %d", len(repo.events))
	}
	if repo.events[0].UserID != userID {
		t.Fatalf("expected event for user %s, got %s", userID, repo.events[0].UserID)
	}
}

func TestService_ListByUser_OnlyReturnsOwnEvents(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	userA := uuid.New()
	userB := uuid.New()

	svc.Log(context.Background(), LogInput{UserID: userA, Action: "login", Resource: "session"})
	svc.Log(context.Background(), LogInput{UserID: userB, Action: "login", Resource: "session"})

	events, err := svc.ListByUser(context.Background(), userA, nil, uuid.Nil, 20)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event for userA, got %d", len(events))
	}
	if events[0].UserID != userA {
		t.Fatalf("expected event scoped to userA, got %s", events[0].UserID)
	}
}

func TestService_ListByUser_PaginatesMostRecentFirst(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	userID := uuid.New()
	base := time.Now().UTC().Add(-time.Hour)

	for i := 0; i < 5; i++ {
		repo.events = append(repo.events, Event{
			ID:        uuid.New(),
			UserID:    userID,
			Action:    "action",
			Resource:  "resource",
			CreatedAt: base.Add(time.Duration(i) * time.Minute),
		})
	}

	page1, err := svc.ListByUser(context.Background(), userID, nil, uuid.Nil, 2)
	if err != nil {
		t.Fatalf("list page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 events in page1, got %d", len(page1))
	}
	// Most recent (index 4, i.e. base+4min) should come first.
	if !page1[0].CreatedAt.After(page1[1].CreatedAt) {
		t.Fatalf("expected page1 events in descending created_at order")
	}

	last := page1[len(page1)-1]
	page2, err := svc.ListByUser(context.Background(), userID, &last.CreatedAt, last.ID, 2)
	if err != nil {
		t.Fatalf("list page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 events in page2, got %d", len(page2))
	}
	for _, e := range page2 {
		if !e.CreatedAt.Before(last.CreatedAt) {
			t.Fatalf("expected page2 events to be older than page1's last event")
		}
	}
}
