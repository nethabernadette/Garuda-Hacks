package offer

import "time"

type OfferStatus string

const (
	OfferStatusPending   OfferStatus = "PENDING"
	OfferStatusCancelled OfferStatus = "CANCELLED"
)

type Offer struct {
	ID                    uint        `gorm:"primaryKey" json:"id"`
	DemandGroupID         uint        `gorm:"column:group_id;not null;index;uniqueIndex:idx_offers_group_producer" json:"demand_group_id"`
	ProducerID            uint        `gorm:"column:producer_id;not null;index;uniqueIndex:idx_offers_group_producer" json:"producer_id"`
	OfferedPrice          float64     `gorm:"column:price;not null" json:"offered_price"`
	EstimatedDeliveryDate time.Time   `gorm:"column:estimated_delivery;not null" json:"estimated_delivery_date"`
	Notes                 string      `gorm:"column:notes;type:text" json:"notes"`
	Status                OfferStatus `gorm:"column:status;type:varchar(30);not null;default:'PENDING';index" json:"status"`
	CreatedAt             time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt             time.Time   `gorm:"column:updated_at" json:"updated_at"`
}

func (Offer) TableName() string {
	return "offers"
}
