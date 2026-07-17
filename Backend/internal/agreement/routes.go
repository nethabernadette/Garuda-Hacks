package agreement

import (
	"net/http"

	"gorm.io/gorm"
)

// RegisterRoutes registers authenticated agreement routes.
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, authenticate func(http.Handler) http.Handler) {
	repository := NewGormRepository(db)
	service := NewService(repository)
	controller := NewController(service)

	mux.Handle("POST /agreements", authenticate(http.HandlerFunc(controller.CreateAgreement)))
	mux.Handle("GET /agreements/{id}", authenticate(http.HandlerFunc(controller.GetAgreement)))
	mux.Handle("PUT /agreements/{id}", authenticate(http.HandlerFunc(controller.UpdateAgreement)))
	mux.Handle("DELETE /agreements/{id}", authenticate(http.HandlerFunc(controller.CancelAgreement)))
	mux.Handle("POST /agreements/{id}/confirm", authenticate(http.HandlerFunc(controller.ConfirmAgreement)))
	mux.Handle("GET /agreements/{id}/items", authenticate(http.HandlerFunc(controller.ListItems)))
	mux.Handle("POST /agreements/{id}/items", authenticate(http.HandlerFunc(controller.AddItem)))
	mux.Handle("PUT /agreements/{id}/items/{itemId}", authenticate(http.HandlerFunc(controller.UpdateItem)))
	mux.Handle("DELETE /agreements/{id}/items/{itemId}", authenticate(http.HandlerFunc(controller.DeleteItem)))
	mux.Handle("GET /agreements/{id}/contact", authenticate(http.HandlerFunc(controller.GetContact)))
}

// Migrate runs GORM migrations for agreement tables only.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Agreement{}, &AgreementItem{})
}
