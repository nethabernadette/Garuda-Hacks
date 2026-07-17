package chat

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository defines database operations required by the chat service.
type Repository interface {
	FindMatchByID(ctx context.Context, matchID string) (*matchRecord, error)
	FindChatRoomByMatchID(ctx context.Context, matchID string) (*ChatRoom, error)
	FindOrCreateChatRoom(ctx context.Context, matchID string) (*ChatRoom, error)
	CreateMessage(ctx context.Context, message *Message) error
	ListMessages(ctx context.Context, chatRoomID string, limit int, offset int) ([]Message, error)
	CountMessages(ctx context.Context, chatRoomID string) (int64, error)
}

// GormRepository implements Repository using GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a GORM-backed chat repository.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// FindMatchByID returns a match projection used for chat authorization.
func (r *GormRepository) FindMatchByID(ctx context.Context, matchID string) (*matchRecord, error) {
	var match matchRecord
	err := r.db.WithContext(ctx).
		Table(match.TableName()).
		Select("id", "buyer_id", "producer_id").
		First(&match, "id = ?", matchID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMatchNotFound
		}
		return nil, err
	}
	return &match, nil
}

// FindChatRoomByMatchID returns the chat room for a match.
func (r *GormRepository) FindChatRoomByMatchID(ctx context.Context, matchID string) (*ChatRoom, error) {
	var room ChatRoom
	err := r.db.WithContext(ctx).First(&room, "match_id = ?", matchID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatRoomNotFound
		}
		return nil, err
	}
	return &room, nil
}

// FindOrCreateChatRoom returns the existing room for a match or creates one.
func (r *GormRepository) FindOrCreateChatRoom(ctx context.Context, matchID string) (*ChatRoom, error) {
	var room ChatRoom
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&room, "match_id = ?", matchID).Error; err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		room = ChatRoom{MatchID: matchID}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "match_id"}},
			DoNothing: true,
		}).Create(&room).Error; err != nil {
			return err
		}

		return tx.First(&room, "match_id = ?", matchID).Error
	})
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// CreateMessage persists a chat message.
func (r *GormRepository) CreateMessage(ctx context.Context, message *Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// ListMessages returns messages sorted oldest first.
func (r *GormRepository) ListMessages(ctx context.Context, chatRoomID string, limit int, offset int) ([]Message, error) {
	var messages []Message
	err := r.db.WithContext(ctx).
		Where("chat_room_id = ?", chatRoomID).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

// CountMessages returns the total number of messages in a chat room.
func (r *GormRepository) CountMessages(ctx context.Context, chatRoomID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&Message{}).
		Where("chat_room_id = ?", chatRoomID).
		Count(&total).Error
	return total, err
}
