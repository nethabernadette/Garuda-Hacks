package product

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router gin.IRouter, db *gorm.DB) {
	repository := NewGormProductRepository(db)
	service := NewProductService(repository)
	controller := NewProductController(service)

	router.POST("/products", controller.Create)
	router.GET("/products", controller.Search)
	router.GET("/producer/products", controller.ListProducerProducts)
	router.GET("/products/:id", controller.GetByID)
	router.PUT("/products/:id", controller.Update)
	router.DELETE("/products/:id", controller.Delete)
	router.PATCH("/products/:id/stock", controller.UpdateStock)
}
