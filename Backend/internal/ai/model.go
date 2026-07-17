package ai

import "time"

// SearchHistory stores non-sensitive search intent for future recommendations.
type SearchHistory struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"column:user_id;type:uuid;not null;index" json:"user_id"`
	Query     string    `gorm:"column:query;type:varchar(255);not null" json:"query"`
	Category  string    `gorm:"column:category;type:varchar(120);index" json:"category,omitempty"`
	Location  string    `gorm:"column:location;type:varchar(180);index" json:"location,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime;index" json:"created_at"`
}

func (SearchHistory) TableName() string {
	return "ai_search_histories"
}

type matchRecord struct {
	ID         string `gorm:"column:id"`
	BuyerID    string `gorm:"column:buyer_id"`
	ProducerID string `gorm:"column:producer_id"`
}

func (matchRecord) TableName() string {
	return "matches"
}
