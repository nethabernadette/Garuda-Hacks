package notifications

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"garuda-hacks/backend/auth"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (h *Controller) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		notificationError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	filter := notificationFilter(c)
	items, err := h.service.List(c.Request.Context(), userID, filter)
	if err != nil {
		handleNotificationError(c, err)
		return
	}
	notificationJSON(c, http.StatusOK, "", ListResponse{Items: items, Page: filter.Page, Limit: filter.Limit})
}

func (h *Controller) UnreadCount(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		notificationError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	count, err := h.service.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		handleNotificationError(c, err)
		return
	}
	notificationJSON(c, http.StatusOK, "", CountResponse{Unread: count})
}

func (h *Controller) MarkRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		notificationError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	notification, err := h.service.MarkRead(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		handleNotificationError(c, err)
		return
	}
	notificationJSON(c, http.StatusOK, "notification marked as read", notification)
}

func (h *Controller) MarkAllRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		notificationError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	if err := h.service.MarkAllRead(c.Request.Context(), userID); err != nil {
		handleNotificationError(c, err)
		return
	}
	notificationJSON(c, http.StatusOK, "all notifications marked as read", nil)
}

func (h *Controller) Delete(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		notificationError(c, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	if err := h.service.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		handleNotificationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func handleNotificationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		notificationError(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrNotificationForbidden):
		notificationError(c, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrNotificationNotFound):
		notificationError(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidNotification):
		notificationError(c, http.StatusBadRequest, err.Error())
	default:
		notificationError(c, http.StatusInternalServerError, "internal server error")
	}
}

func currentUserID(c *gin.Context) (string, bool) {
	if claims, ok := auth.GetCurrentUser(c.Request.Context()); ok {
		return claims.UserID, true
	}
	for _, key := range []string{"CurrentUserID", "currentUserID", "user_id", "userID"} {
		value, exists := c.Get(key)
		if !exists {
			continue
		}
		if id, ok := value.(string); ok && strings.TrimSpace(id) != "" {
			return strings.TrimSpace(id), true
		}
	}
	return "", false
}

func notificationFilter(c *gin.Context) QueryFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	filter := QueryFilter{Page: page, Limit: limit}
	if value := strings.TrimSpace(c.Query("unread")); value != "" {
		unread := value == "true" || value == "1"
		filter.Unread = &unread
	}
	filter.Page, filter.Limit, filter.Offset = normalizePagination(filter.Page, filter.Limit)
	return filter
}

func notificationJSON(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func notificationError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   message,
	})
}
