package product

import "time"

type Product struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	ProducerID         uint       `gorm:"column:producer_id;not null;index" json:"producer_id"`
	Name               string     `gorm:"column:name;type:varchar(150);not null;index" json:"name"`
	Category           string     `gorm:"column:category;type:varchar(100);not null;index" json:"category"`
	Description        string     `gorm:"column:description;type:text" json:"description"`
	Unit               string     `gorm:"column:unit;type:varchar(50);not null" json:"unit"`
	AvailableStock     float64    `gorm:"column:stock;not null;default:0" json:"available_stock"`
	ProductionCapacity float64    `gorm:"column:capacity;not null;default:0" json:"production_capacity"`
	HarvestDate        *time.Time `gorm:"column:harvest_date" json:"harvest_date,omitempty"`
	MinimumOrder       float64    `gorm:"column:minimum_order;not null;default:0" json:"minimum_order"`
	ImageURL           string     `gorm:"column:image_url;type:text" json:"image_url"`
	CreatedAt          time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (Product) TableName() string {
	return "products"
}
