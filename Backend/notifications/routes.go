package notifications

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router gin.IRouter, db *gorm.DB) Service {
	repository := NewGormRepository(db)
	service := NewService(repository)
	controller := NewController(service)

	router.GET("/notifications", controller.List)
	router.GET("/notifications/unread-count", controller.UnreadCount)
	router.PATCH("/notifications/read-all", controller.MarkAllRead)
	router.PATCH("/notifications/:id/read", controller.MarkRead)
	router.DELETE("/notifications/:id", controller.Delete)

	return service
}
