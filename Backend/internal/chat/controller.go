package chat

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"garuda-hacks/backend/auth"
)

// Controller handles chat HTTP requests.
type Controller struct {
	service *Service
}

// NewController creates a chat controller.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// CreateChatRoom handles POST /matches/{matchId}/chat.
func (c *Controller) CreateChatRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := c.service.CreateChatRoom(r.Context(), r.PathValue("matchId"), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, JSONResponse{
		Success: true,
		Message: "chat room ready",
		Data:    response,
	})
}

// GetChatRoom handles GET /matches/{matchId}/chat.
func (c *Controller) GetChatRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	response, err := c.service.GetChatRoom(r.Context(), r.PathValue("matchId"), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

// ListMessages handles GET /matches/{matchId}/chat/messages.
func (c *Controller) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	page, limit := paginationFromRequest(r)
	response, err := c.service.ListMessages(r.Context(), r.PathValue("matchId"), userID, page, limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Data:    response,
	})
}

// SendMessage handles POST /matches/{matchId}/chat/messages.
func (c *Controller) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthorized.Error())
		return
	}

	var req SendMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest.Error())
		return
	}

	response, err := c.service.SendMessage(r.Context(), r.PathValue("matchId"), userID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, JSONResponse{
		Success: true,
		Message: "message sent",
		Data:    response,
	})
}

func currentUserID(r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func paginationFromRequest(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	return page, limit
}

func decodeJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrMatchNotFound),
		errors.Is(err, ErrChatRoomNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidMatchID),
		errors.Is(err, ErrInvalidRequest),
		errors.Is(err, ErrEmptyMessage),
		errors.Is(err, ErrMessageTooLong):
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
