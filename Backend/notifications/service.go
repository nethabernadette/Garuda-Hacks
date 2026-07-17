package notifications

import (
	"context"
	"time"
)

type Service interface {
	CreateUnique(ctx context.Context, userID string, notificationType string, title string, message string, referenceType string, referenceID string) error
	List(ctx context.Context, userID string, filter QueryFilter) ([]Notification, error)
	UnreadCount(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, userID string, id string) (*Notification, error)
	MarkAllRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, userID string, id string) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) CreateUnique(ctx context.Context, userID string, notificationType string, title string, message string, referenceType string, referenceID string) error {
	notification := &Notification{
		UserID:        userID,
		Type:          notificationType,
		Title:         title,
		Message:       message,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
	}
	if err := validateNotification(notification); err != nil {
		return err
	}
	return s.repository.CreateUnique(ctx, notification)
}

func (s *service) List(ctx context.Context, userID string, filter QueryFilter) ([]Notification, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	filter.Page, filter.Limit, filter.Offset = normalizePagination(filter.Page, filter.Limit)
	return s.repository.ListByUser(ctx, userID, filter)
}

func (s *service) UnreadCount(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, ErrUnauthorized
	}
	return s.repository.UnreadCount(ctx, userID)
}

func (s *service) MarkRead(ctx context.Context, userID string, id string) (*Notification, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	notification, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if notification.UserID != userID {
		return nil, ErrNotificationForbidden
	}
	if !notification.IsRead {
		now := time.Now().UTC()
		notification.IsRead = true
		notification.ReadAt = &now
		if err := s.repository.Update(ctx, notification); err != nil {
			return nil, err
		}
	}
	return notification, nil
}

func (s *service) MarkAllRead(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrUnauthorized
	}
	return s.repository.MarkAllRead(ctx, userID)
}

func (s *service) Delete(ctx context.Context, userID string, id string) error {
	if userID == "" {
		return ErrUnauthorized
	}
	notification, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if notification.UserID != userID {
		return ErrNotificationForbidden
	}
	return s.repository.Delete(ctx, notification)
}
