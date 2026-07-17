package matches

import (
	"net/http"

	"gorm.io/gorm"
)

func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, authenticate func(http.Handler) http.Handler) {
	repository := NewGormRepository(db)
	service := NewService(repository)
	controller := NewController(service)

	mux.Handle("POST /matches/interest", authenticate(http.HandlerFunc(controller.CreateInterest)))
	mux.Handle("GET /matches", authenticate(http.HandlerFunc(controller.ListMatches)))
	mux.Handle("GET /matches/{id}", authenticate(http.HandlerFunc(controller.GetMatch)))
}
