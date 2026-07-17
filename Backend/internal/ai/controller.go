package ai

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"garuda-hacks/backend/auth"
)

const maxRequestBodyBytes = 1 << 20

// Controller handles AI HTTP requests.
type Controller struct {
	service *Service
}

// NewController creates an AI controller.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// TrackSearch handles POST /ai/search-history.
func (c *Controller) TrackSearch(w http.ResponseWriter, r *http.Request) {
	claims, ok := currentClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req SearchHistoryInput
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}
	response, err := c.service.TrackSearch(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, JSONResponse{
		Success: true,
		Message: "search history recorded successfully",
		Data:    response,
	})
}

// Recommendations handles GET /recommendations and GET /ai/recommendations.
func (c *Controller) Recommendations(w http.ResponseWriter, r *http.Request) {
	claims, ok := currentClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	response, err := c.service.Recommendations(r.Context(), claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

// Matchmaking handles POST /ai/matchmaking.
func (c *Controller) Matchmaking(w http.ResponseWriter, r *http.Request) {
	claims, ok := currentClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	req := MatchmakingRequest{}
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
			return
		}
	}

	response, err := c.service.Matchmaking(r.Context(), claims.UserID, claims.Role, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

// VerifyAgreement handles GET /agreements/{id}/ai-verification.
func (c *Controller) VerifyAgreement(w http.ResponseWriter, r *http.Request) {
	claims, ok := currentClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	response, err := c.service.VerifyAgreement(r.Context(), claims.UserID, r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

// CompareAgreementSubmissions handles POST /agreements/{id}/ai-verification.
func (c *Controller) CompareAgreementSubmissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := currentClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req AgreementVerificationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}
	response, err := c.service.CompareAgreementSubmissions(r.Context(), claims.UserID, r.PathValue("id"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

// SummarizeNegotiation handles GET /agreements/{id}/negotiation-summary.
func (c *Controller) SummarizeNegotiation(w http.ResponseWriter, r *http.Request) {
	claims, ok := currentClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}
	response, err := c.service.SummarizeNegotiation(r.Context(), claims.UserID, r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

func currentClaims(r *http.Request) (*auth.Claims, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return nil, false
	}
	return claims, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return ErrInvalidRequest
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ErrInvalidRequest
	}
	return nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrUserNotFound),
		errors.Is(err, ErrPostNotFound),
		errors.Is(err, ErrAgreementNotFound),
		errors.Is(err, ErrMatchNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAgreementNotReady):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidRequest),
		errors.Is(err, ErrMissingAgreementID):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, JSONResponse{
		Success: false,
		Error:   message,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, response JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}
