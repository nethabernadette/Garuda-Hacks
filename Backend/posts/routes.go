package posts

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router gin.IRouter, db *gorm.DB, notifications NotificationCreator) {
	repository := NewGormRepository(db)
	service := NewService(repository, notifications)
	controller := NewController(service)

	router.GET("/posts", controller.Feed)
	router.GET("/posts/search", controller.Search)

	router.POST("/posts/supply", controller.CreateSupply)
	router.GET("/posts/supply", controller.ListSupply)
	router.GET("/posts/supply/me", controller.MySupply)
	router.GET("/posts/supply/:id", controller.GetSupply)
	router.PUT("/posts/supply/:id", controller.UpdateSupply)
	router.PATCH("/posts/supply/:id/close", controller.CloseSupply)
	router.DELETE("/posts/supply/:id", controller.DeleteSupply)

	router.POST("/posts/demand", controller.CreateDemand)
	router.GET("/posts/demand", controller.ListDemand)
	router.GET("/posts/demand/me", controller.MyDemand)
	router.GET("/posts/demand/:id", controller.GetDemand)
	router.PUT("/posts/demand/:id", controller.UpdateDemand)
	router.PATCH("/posts/demand/:id/close", controller.CloseDemand)
	router.DELETE("/posts/demand/:id", controller.DeleteDemand)
}
