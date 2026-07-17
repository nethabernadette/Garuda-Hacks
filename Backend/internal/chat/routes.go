package chat

import (
	"net/http"

	"gorm.io/gorm"
)

// RegisterRoutes registers authenticated chat routes.
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, authenticate func(http.Handler) http.Handler) {
	repository := NewGormRepository(db)
	service := NewService(repository)
	controller := NewController(service)

	mux.Handle("POST /matches/{matchId}/chat", authenticate(http.HandlerFunc(controller.CreateChatRoom)))
	mux.Handle("GET /matches/{matchId}/chat", authenticate(http.HandlerFunc(controller.GetChatRoom)))
	mux.Handle("GET /matches/{matchId}/chat/messages", authenticate(http.HandlerFunc(controller.ListMessages)))
	mux.Handle("POST /matches/{matchId}/chat/messages", authenticate(http.HandlerFunc(controller.SendMessage)))
}

// Migrate runs GORM migrations for chat tables only.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&ChatRoom{}, &Message{})
}
