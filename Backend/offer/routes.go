package offer

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router gin.IRouter, db *gorm.DB) {
	repository := NewGormOfferRepository(db)
	service := NewOfferService(repository)
	controller := NewOfferController(service)

	router.POST("/offers", controller.Create)
	router.GET("/offers", controller.ListProducerOffers)
	router.GET("/offers/:id", controller.GetByID)
	router.PUT("/offers/:id", controller.Update)
	router.DELETE("/offers/:id", controller.Cancel)
	router.GET("/demand-groups/:id/offers", controller.ListByDemandGroup)
}
