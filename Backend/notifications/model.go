package notifications

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID            string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        string         `gorm:"column:user_id;type:uuid;not null;index;uniqueIndex:idx_notifications_unique_ref" json:"user_id"`
	Type          string         `gorm:"column:type;type:varchar(80);not null;index;uniqueIndex:idx_notifications_unique_ref" json:"type"`
	Title         string         `gorm:"column:title;type:varchar(180);not null" json:"title"`
	Message       string         `gorm:"column:message;type:text;not null" json:"message"`
	ReferenceType string         `gorm:"column:reference_type;type:varchar(80);not null;index;uniqueIndex:idx_notifications_unique_ref" json:"reference_type"`
	ReferenceID   string         `gorm:"column:reference_id;type:uuid;not null;index;uniqueIndex:idx_notifications_unique_ref" json:"reference_id"`
	IsRead        bool           `gorm:"column:is_read;not null;default:false;index" json:"is_read"`
	CreatedAt     time.Time      `gorm:"column:created_at;not null;autoCreateTime;index" json:"created_at"`
	ReadAt        *time.Time     `gorm:"column:read_at" json:"read_at,omitempty"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}
