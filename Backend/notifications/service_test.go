package notifications

import (
	"context"
	"errors"
	"testing"
)

type fakeNotificationRepository struct {
	records []Notification
}

func (r *fakeNotificationRepository) CreateUnique(ctx context.Context, notification *Notification) error {
	if notification.ID == "" {
		notification.ID = "notification-1"
	}
	for _, record := range r.records {
		if record.UserID == notification.UserID && record.Type == notification.Type && record.ReferenceType == notification.ReferenceType && record.ReferenceID == notification.ReferenceID {
			return nil
		}
	}
	r.records = append(r.records, *notification)
	return nil
}

func (r *fakeNotificationRepository) ListByUser(ctx context.Context, userID string, filter QueryFilter) ([]Notification, error) {
	return r.records, nil
}

func (r *fakeNotificationRepository) GetByID(ctx context.Context, id string) (*Notification, error) {
	for i := range r.records {
		if r.records[i].ID == id {
			return &r.records[i], nil
		}
	}
	return nil, ErrNotificationNotFound
}

func (r *fakeNotificationRepository) UnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	for _, record := range r.records {
		if record.UserID == userID && !record.IsRead {
			count++
		}
	}
	return count, nil
}

func (r *fakeNotificationRepository) Update(ctx context.Context, notification *Notification) error {
	return nil
}

func (r *fakeNotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	for i := range r.records {
		if r.records[i].UserID == userID {
			r.records[i].IsRead = true
		}
	}
	return nil
}

func (r *fakeNotificationRepository) Delete(ctx context.Context, notification *Notification) error {
	return nil
}

func TestCreateUniquePreventsDuplicateNotifications(t *testing.T) {
	repo := &fakeNotificationRepository{}
	svc := NewService(repo)

	err := svc.CreateUnique(context.Background(), "user-1", "demand_relevant", "Title", "Message", "demand_post", "demand-1")
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	err = svc.CreateUnique(context.Background(), "user-1", "demand_relevant", "Title", "Message", "demand_post", "demand-1")
	if err != nil {
		t.Fatalf("unexpected duplicate create error: %v", err)
	}
	if len(repo.records) != 1 {
		t.Fatalf("expected one notification, got %d", len(repo.records))
	}
}

func TestMarkReadRequiresOwner(t *testing.T) {
	repo := &fakeNotificationRepository{records: []Notification{{
		ID:            "notification-1",
		UserID:        "user-1",
		Type:          "demand_relevant",
		Title:         "Title",
		Message:       "Message",
		ReferenceType: "demand_post",
		ReferenceID:   "demand-1",
	}}}
	svc := NewService(repo)

	_, err := svc.MarkRead(context.Background(), "user-2", "notification-1")
	if !errors.Is(err, ErrNotificationForbidden) {
		t.Fatalf("expected ErrNotificationForbidden, got %v", err)
	}

	notification, err := svc.MarkRead(context.Background(), "user-1", "notification-1")
	if err != nil {
		t.Fatalf("unexpected mark read error: %v", err)
	}
	if !notification.IsRead || notification.ReadAt == nil {
		t.Fatal("expected notification to be marked read")
	}
}
