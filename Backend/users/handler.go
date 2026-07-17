package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type ClaimsExtractor func(context.Context) (Principal, bool)

type Handler struct {
	service       *Service
	extractClaims ClaimsExtractor
}

func NewHandler(service *Service, extractClaims ClaimsExtractor) *Handler {
	return &Handler{
		service:       service,
		extractClaims: extractClaims,
	}
}

func (h *Handler) GetCurrentProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := h.service.GetCurrentProfile(r.Context(), principal)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	id, err := userIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidUserID.Error())
		return
	}

	response, err := h.service.GetUserByID(r.Context(), principal, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := h.service.ListUsers(r.Context(), principal)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

func (h *Handler) UpdateCurrentProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := h.service.UpdateCurrentProfile(r.Context(), principal, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "profile updated successfully",
		Data:    response,
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	id, err := userIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidUserID.Error())
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := h.service.UpdateProfile(r.Context(), principal, id, req, false)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "profile updated successfully",
		Data:    response,
	})
}

func (h *Handler) GetCurrentVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := h.service.GetCurrentVerification(r.Context(), principal)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

func (h *Handler) SubmitCurrentVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req NIBVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := h.service.SubmitCurrentVerification(r.Context(), principal, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "verification submitted successfully",
		Data:    response,
	})
}

func (h *Handler) ListVerifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := h.service.ListVerifications(r.Context(), principal)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

func (h *Handler) ReviewVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := h.principalFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	id, err := userIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	var req ReviewNIBVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := h.service.ReviewVerification(r.Context(), principal, id, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "verification reviewed successfully",
		Data:    response,
	})
}

func (h *Handler) principalFromRequest(r *http.Request) (Principal, bool) {
	if h.extractClaims == nil {
		return Principal{}, false
	}

	return h.extractClaims(r.Context())
}

func userIDFromRequest(r *http.Request) (string, error) {
	rawID := strings.TrimSpace(r.URL.Query().Get("id"))
	if rawID == "" {
		rawID = strings.Trim(strings.TrimPrefix(r.URL.Path, "/"), "/")
		segments := strings.Split(rawID, "/")
		if len(segments) > 0 {
			rawID = segments[len(segments)-1]
		}
	}

	if rawID == "" {
		return "", ErrInvalidUserID
	}

	return rawID, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrUserNotFound),
		errors.Is(err, ErrVerificationNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidUserID),
		errors.Is(err, ErrInvalidRequest),
		errors.Is(err, ErrRequiredCompanyName),
		errors.Is(err, ErrRequiredPhone),
		errors.Is(err, ErrRequiredCity),
		errors.Is(err, ErrRequiredNIBNumber),
		errors.Is(err, ErrInvalidNIBNumber),
		errors.Is(err, ErrInvalidVerificationStatus),
		errors.Is(err, ErrRequiredRejectionReason):
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
