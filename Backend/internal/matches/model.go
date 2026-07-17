package matches

import (
	"time"
)

type MatchStatus string

const (
	MatchStatusActive MatchStatus = "ACTIVE"
)

type Match struct {
	ID         string      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BuyerID    string      `gorm:"column:buyer_id;type:uuid;not null;index;uniqueIndex:idx_matches_buyer_producer" json:"buyer_id"`
	ProducerID string      `gorm:"column:producer_id;type:uuid;not null;index;uniqueIndex:idx_matches_buyer_producer" json:"producer_id"`
	Status     MatchStatus `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE';index" json:"status"`
	CreatedAt  time.Time   `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

func (Match) TableName() string {
	return "matches"
}
