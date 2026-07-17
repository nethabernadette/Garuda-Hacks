package auth

import "time"

type RefreshToken struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"column:user_id;type:uuid;not null;index" json:"user_id"`
	Token     string     `gorm:"column:token;type:varchar(255);not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().UTC().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}
