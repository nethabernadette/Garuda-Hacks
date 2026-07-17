package ai

import (
	"net/http"

	"gorm.io/gorm"
)

// RegisterRoutes registers authenticated AI routes.
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, authenticate func(http.Handler) http.Handler) {
	repository := NewGormRepository(db)
	service := NewService(repository, NewGroqClient(LoadConfigFromEnv()))
	controller := NewController(service)

	mux.Handle("GET /recommendations", authenticate(http.HandlerFunc(controller.Recommendations)))
	mux.Handle("GET /ai/recommendations", authenticate(http.HandlerFunc(controller.Recommendations)))
	mux.Handle("POST /ai/search-history", authenticate(http.HandlerFunc(controller.TrackSearch)))
	mux.Handle("POST /ai/matchmaking", authenticate(http.HandlerFunc(controller.Matchmaking)))
	mux.Handle("GET /agreements/{id}/ai-verification", authenticate(http.HandlerFunc(controller.VerifyAgreement)))
	mux.Handle("POST /agreements/{id}/ai-verification", authenticate(http.HandlerFunc(controller.CompareAgreementSubmissions)))
	mux.Handle("GET /agreements/{id}/negotiation-summary", authenticate(http.HandlerFunc(controller.SummarizeNegotiation)))
}
