package chat

import (
	"time"

	"garuda-hacks/backend/users"
	"gorm.io/gorm"
)

// ChatRoom stores the one-to-one chat room created for a successful match.
type ChatRoom struct {
	ID        string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MatchID   string         `gorm:"column:match_id;type:uuid;not null;uniqueIndex" json:"match_id"`
	Messages  []Message      `gorm:"foreignKey:ChatRoomID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"messages,omitempty"`
	CreatedAt time.Time      `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName returns the database table name for chat rooms.
func (ChatRoom) TableName() string {
	return "chat_rooms"
}

// Message stores a single negotiation message in a chat room.
type Message struct {
	ID         string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ChatRoomID string         `gorm:"column:chat_room_id;type:uuid;not null;index" json:"chat_room_id"`
	SenderID   string         `gorm:"column:sender_id;type:uuid;not null;index" json:"sender_id"`
	Message    string         `gorm:"column:message;type:text;not null" json:"message"`
	ChatRoom   ChatRoom       `gorm:"foreignKey:ChatRoomID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Sender     users.User     `gorm:"foreignKey:SenderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	CreatedAt  time.Time      `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName returns the database table name for messages.
func (Message) TableName() string {
	return "messages"
}

type matchRecord struct {
	ID         string `gorm:"column:id"`
	BuyerID    string `gorm:"column:buyer_id"`
	ProducerID string `gorm:"column:producer_id"`
}

func (matchRecord) TableName() string {
	return "matches"
}
