package document

import (
	"encoding/json"
	"errors"
	"net/http"

	"garuda-hacks/backend/auth"
	agreementmodule "garuda-hacks/backend/internal/agreement"
)

// Controller handles procurement document HTTP requests.
type Controller struct {
	service *Service
}

// NewController creates a document controller.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// GetDocument handles GET /agreements/{id}/document.
func (c *Controller) GetDocument(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := c.service.GetDocument(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

// GetHTML handles GET /agreements/{id}/document/html.
func (c *Controller) GetHTML(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	html, err := c.service.GetHTML(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// GetContact handles GET /agreements/{id}/contact.
func (c *Controller) GetContact(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := c.service.GetContact(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

func currentUserID(r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrAgreementNotReady),
		errors.Is(err, ErrAgreementCancelled):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, agreementmodule.ErrAgreementNotFound),
		errors.Is(err, ErrMatchNotFound),
		errors.Is(err, ErrContactNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidAgreementID):
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
