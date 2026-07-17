package agreement

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"garuda-hacks/backend/auth"
)

const maxRequestBodyBytes = 1 << 20

// Controller handles agreement HTTP requests.
type Controller struct {
	service *Service
}

// NewController creates an agreement controller.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// CreateAgreement handles POST /agreements.
func (c *Controller) CreateAgreement(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req CreateAgreementRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := c.service.CreateAgreement(r.Context(), userID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, JSONResponse{
		Success: true,
		Message: "agreement created successfully",
		Data:    response,
	})
}

// GetAgreement handles GET /agreements/{id}.
func (c *Controller) GetAgreement(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := c.service.GetAgreement(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

// UpdateAgreement handles PUT /agreements/{id}.
func (c *Controller) UpdateAgreement(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req UpdateAgreementRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := c.service.UpdateAgreement(r.Context(), userID, r.PathValue("id"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "agreement updated successfully",
		Data:    response,
	})
}

// CancelAgreement handles DELETE /agreements/{id}.
func (c *Controller) CancelAgreement(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	if err := c.service.CancelAgreement(r.Context(), userID, r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "agreement cancelled successfully",
	})
}

// ConfirmAgreement handles POST /agreements/{id}/confirm.
func (c *Controller) ConfirmAgreement(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := c.service.ConfirmAgreement(r.Context(), userID, r.PathValue("id"), ConfirmAgreementRequest{})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "agreement confirmed successfully",
		Data:    response,
	})
}

// ListItems handles GET /agreements/{id}/items.
func (c *Controller) ListItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := c.service.ListItems(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{Success: true, Data: response})
}

// AddItem handles POST /agreements/{id}/items.
func (c *Controller) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req AgreementItemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := c.service.AddItem(r.Context(), userID, r.PathValue("id"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, JSONResponse{
		Success: true,
		Message: "agreement item added successfully",
		Data:    response,
	})
}

// UpdateItem handles PUT /agreements/{id}/items/{itemId}.
func (c *Controller) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req AgreementItemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := c.service.UpdateItem(r.Context(), userID, r.PathValue("id"), r.PathValue("itemId"), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "agreement item updated successfully",
		Data:    response,
	})
}

// DeleteItem handles DELETE /agreements/{id}/items/{itemId}.
func (c *Controller) DeleteItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	if err := c.service.DeleteItem(r.Context(), userID, r.PathValue("id"), r.PathValue("itemId")); err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "agreement item deleted successfully",
	})
}

func currentUserID(r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
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
	case errors.Is(err, ErrForbidden),
		errors.Is(err, ErrContactHidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrMatchNotFound),
		errors.Is(err, ErrAgreementNotFound),
		errors.Is(err, ErrAgreementItemNotFound),
		errors.Is(err, ErrContactNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrActiveAgreementExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrAgreementNotEditable):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidRequest),
		errors.Is(err, ErrInvalidMatchID),
		errors.Is(err, ErrInvalidAgreementID),
		errors.Is(err, ErrInvalidAgreementItemID),
		errors.Is(err, ErrAgreementNeedsItems),
		errors.Is(err, ErrRequiredProductName),
		errors.Is(err, ErrInvalidQuantity),
		errors.Is(err, ErrRequiredUnit),
		errors.Is(err, ErrInvalidUnitPrice),
		errors.Is(err, ErrRequiredCurrency),
		errors.Is(err, ErrRequiredDeliveryDate),
		errors.Is(err, ErrRequiredDeliveryAddress),
		errors.Is(err, ErrRequiredPaymentTerms):
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
