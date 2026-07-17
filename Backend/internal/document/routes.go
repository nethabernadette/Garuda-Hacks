package document

import (
	"net/http"

	"gorm.io/gorm"
)

// RegisterRoutes registers authenticated procurement document routes.
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, authenticate func(http.Handler) http.Handler) {
	repository := NewGormRepository(db)
	service := NewService(repository, NewHTMLGenerator())
	controller := NewController(service)

	mux.Handle("GET /agreements/{id}/document", authenticate(http.HandlerFunc(controller.GetDocument)))
	mux.Handle("GET /agreements/{id}/document/html", authenticate(http.HandlerFunc(controller.GetHTML)))
	mux.Handle("GET /agreements/{id}/contact", authenticate(http.HandlerFunc(controller.GetContact)))
}
