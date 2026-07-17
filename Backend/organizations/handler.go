package organizations

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"garuda-hacks/backend/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateOrgRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidOrgRequest.Error())
		return
	}

	response, err := h.service.Create(r.Context(), claims.UserID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, JSONResponse{
		Success: true,
		Message: "organization created successfully",
		Data:    response,
	})
}

func (h *Handler) FindByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, ErrInvalidOrgRequest.Error())
		return
	}

	response, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		writeError(w, http.StatusBadRequest, ErrInvalidOrgRequest.Error())
		return
	}

	var req UpdateOrgRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidOrgRequest.Error())
		return
	}

	response, err := h.service.Update(r.Context(), id, claims.UserID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "organization updated successfully",
		Data:    response,
	})
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	response, err := h.service.Join(r.Context(), id, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "joined organization successfully",
		Data:    response,
	})
}

func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	err := h.service.Leave(r.Context(), id, claims.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "left organization successfully",
	})
}

func (h *Handler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	var req TransferOwnershipRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidOrgRequest.Error())
		return
	}

	response, err := h.service.TransferOwnership(r.Context(), id, claims.UserID, req.NewOwnerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: "ownership transferred successfully",
		Data:    response,
	})
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	response, err := h.service.ListMembers(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func extractIDFromPath(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) > 1 {
		return segments[1]
	}
	if len(segments) > 0 {
		return segments[0]
	}
	return ""
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrOrgNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrUnauthorizedAction):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrAlreadyMember),
		errors.Is(err, ErrCannotLeaveOwner),
		errors.Is(err, ErrAlreadyOwner):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidOrgRequest),
		errors.Is(err, ErrRequiredOrgName):
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
