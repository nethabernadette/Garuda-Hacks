package notifications

import (
	"errors"
	"strings"
)

var (
	ErrNotificationNotFound  = errors.New("notification not found")
	ErrNotificationForbidden = errors.New("user is not allowed to access this notification")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrInvalidNotification   = errors.New("invalid notification")
)

func validateNotification(notification *Notification) error {
	notification.UserID = strings.TrimSpace(notification.UserID)
	notification.Type = strings.TrimSpace(notification.Type)
	notification.Title = strings.TrimSpace(notification.Title)
	notification.Message = strings.TrimSpace(notification.Message)
	notification.ReferenceType = strings.TrimSpace(notification.ReferenceType)
	notification.ReferenceID = strings.TrimSpace(notification.ReferenceID)
	if notification.UserID == "" || notification.Type == "" || notification.Title == "" || notification.Message == "" || notification.ReferenceType == "" || notification.ReferenceID == "" {
		return ErrInvalidNotification
	}
	return nil
}

func normalizePagination(page, limit int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit, (page - 1) * limit
}
