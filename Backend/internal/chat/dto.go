package chat

import "time"

// SendMessageRequest is the request body for creating a message.
type SendMessageRequest struct {
	Message string `json:"message" validate:"required"`
}

// ChatRoomResponse is returned for chat room endpoints.
type ChatRoomResponse struct {
	ID        string    `json:"id"`
	MatchID   string    `json:"match_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MessageResponse is returned for message endpoints.
type MessageResponse struct {
	ID         string    `json:"id"`
	ChatRoomID string    `json:"chat_room_id"`
	SenderID   string    `json:"sender_id"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PaginationResponse describes the current message page.
type PaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// ConversationResponse returns chat room details with a page of messages.
type ConversationResponse struct {
	ChatRoom   ChatRoomResponse   `json:"chat_room"`
	Messages   []MessageResponse  `json:"messages"`
	Pagination PaginationResponse `json:"pagination"`
}

// JSONResponse matches the response envelope used by existing modules.
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func newChatRoomResponse(room *ChatRoom) ChatRoomResponse {
	return ChatRoomResponse{
		ID:        room.ID,
		MatchID:   room.MatchID,
		CreatedAt: room.CreatedAt,
		UpdatedAt: room.UpdatedAt,
	}
}

func newMessageResponse(message *Message) MessageResponse {
	return MessageResponse{
		ID:         message.ID,
		ChatRoomID: message.ChatRoomID,
		SenderID:   message.SenderID,
		Message:    message.Message,
		CreatedAt:  message.CreatedAt,
		UpdatedAt:  message.UpdatedAt,
	}
}

func newMessageResponses(messages []Message) []MessageResponse {
	responses := make([]MessageResponse, 0, len(messages))
	for i := range messages {
		responses = append(responses, newMessageResponse(&messages[i]))
	}
	return responses
}
