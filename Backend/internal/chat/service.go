package chat

import (
	"context"
	"errors"
	"strings"
)

const (
	defaultMessageLimit = 20
	maxMessageLimit     = 100
	maxMessageLength    = 2000
)

var (
	ErrUnauthorized     = errors.New("authentication is required")
	ErrForbidden        = errors.New("forbidden: user is not part of this match")
	ErrInvalidMatchID   = errors.New("match id is invalid")
	ErrMatchNotFound    = errors.New("match not found")
	ErrChatRoomNotFound = errors.New("chat room not found")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrEmptyMessage     = errors.New("message cannot be empty")
	ErrMessageTooLong   = errors.New("message is too long")
)

// Service contains chat business logic.
type Service struct {
	repository Repository
}

// NewService creates a chat service.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// CreateChatRoom creates or returns a chat room for a match.
func (s *Service) CreateChatRoom(ctx context.Context, matchID string, userID string) (*ChatRoomResponse, error) {
	if err := validateIdentity(matchID, userID); err != nil {
		return nil, err
	}
	if err := s.authorizeMatchAccess(ctx, matchID, userID); err != nil {
		return nil, err
	}

	room, err := s.repository.FindOrCreateChatRoom(ctx, matchID)
	if err != nil {
		return nil, err
	}
	response := newChatRoomResponse(room)
	return &response, nil
}

// GetChatRoom returns chat room information for a match.
func (s *Service) GetChatRoom(ctx context.Context, matchID string, userID string) (*ChatRoomResponse, error) {
	if err := validateIdentity(matchID, userID); err != nil {
		return nil, err
	}
	if err := s.authorizeMatchAccess(ctx, matchID, userID); err != nil {
		return nil, err
	}

	room, err := s.repository.FindChatRoomByMatchID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	response := newChatRoomResponse(room)
	return &response, nil
}

// ListMessages returns a paginated page of chat messages sorted oldest first.
func (s *Service) ListMessages(ctx context.Context, matchID string, userID string, page int, limit int) (*ConversationResponse, error) {
	if err := validateIdentity(matchID, userID); err != nil {
		return nil, err
	}
	if err := s.authorizeMatchAccess(ctx, matchID, userID); err != nil {
		return nil, err
	}

	room, err := s.repository.FindChatRoomByMatchID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	normalizedPage, normalizedLimit, offset := normalizePagination(page, limit)
	messages, err := s.repository.ListMessages(ctx, room.ID, normalizedLimit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.repository.CountMessages(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	return &ConversationResponse{
		ChatRoom: newChatRoomResponse(room),
		Messages: newMessageResponses(messages),
		Pagination: PaginationResponse{
			Page:  normalizedPage,
			Limit: normalizedLimit,
			Total: total,
		},
	}, nil
}

// SendMessage creates a new message from the authenticated user.
func (s *Service) SendMessage(ctx context.Context, matchID string, userID string, req SendMessageRequest) (*MessageResponse, error) {
	if err := validateIdentity(matchID, userID); err != nil {
		return nil, err
	}
	if err := s.authorizeMatchAccess(ctx, matchID, userID); err != nil {
		return nil, err
	}

	room, err := s.repository.FindChatRoomByMatchID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(req.Message)
	if text == "" {
		return nil, ErrEmptyMessage
	}
	if len([]rune(text)) > maxMessageLength {
		return nil, ErrMessageTooLong
	}

	message := &Message{
		ChatRoomID: room.ID,
		SenderID:   userID,
		Message:    text,
	}
	if err := s.repository.CreateMessage(ctx, message); err != nil {
		return nil, err
	}

	response := newMessageResponse(message)
	return &response, nil
}

func (s *Service) authorizeMatchAccess(ctx context.Context, matchID string, userID string) error {
	match, err := s.repository.FindMatchByID(ctx, matchID)
	if err != nil {
		return err
	}
	if match.BuyerID != userID && match.ProducerID != userID {
		return ErrForbidden
	}
	return nil
}

func validateIdentity(matchID string, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrUnauthorized
	}
	if strings.TrimSpace(matchID) == "" {
		return ErrInvalidMatchID
	}
	return nil
}

func normalizePagination(page int, limit int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}
	return page, limit, (page - 1) * limit
}
