package posts

import (
	"time"

	"gorm.io/gorm"
)

type SupplyPostStatus string

const (
	SupplyPostStatusDraft   SupplyPostStatus = "draft"
	SupplyPostStatusActive  SupplyPostStatus = "active"
	SupplyPostStatusClosed  SupplyPostStatus = "closed"
	SupplyPostStatusExpired SupplyPostStatus = "expired"
)

type DemandPostStatus string

const (
	DemandPostStatusDraft     DemandPostStatus = "draft"
	DemandPostStatusOpen      DemandPostStatus = "open"
	DemandPostStatusMatched   DemandPostStatus = "matched"
	DemandPostStatusClosed    DemandPostStatus = "closed"
	DemandPostStatusExpired   DemandPostStatus = "expired"
	DemandPostStatusCancelled DemandPostStatus = "cancelled"
)

type SupplyPost struct {
	ID                   string           `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProducerID           string           `gorm:"column:producer_id;type:uuid;not null;index" json:"producer_id"`
	ProductName          string           `gorm:"column:product_name;type:varchar(150);not null;index" json:"product_name"`
	Category             string           `gorm:"column:category;type:varchar(120);not null;index" json:"category"`
	Subcategory          string           `gorm:"column:subcategory;type:varchar(120);index" json:"subcategory,omitempty"`
	Description          string           `gorm:"column:description;type:text" json:"description,omitempty"`
	Quantity             float64          `gorm:"column:quantity;not null" json:"quantity"`
	Unit                 string           `gorm:"column:unit;type:varchar(50);not null;index" json:"unit"`
	MinimumOrderQuantity float64          `gorm:"column:minimum_order_quantity;not null;default:0" json:"minimum_order_quantity"`
	PriceMin             float64          `gorm:"column:price_min;not null;default:0;index" json:"price_min"`
	PriceMax             float64          `gorm:"column:price_max;not null;default:0;index" json:"price_max"`
	Location             string           `gorm:"column:location;type:varchar(180);not null;index" json:"location"`
	DeliveryArea         string           `gorm:"column:delivery_area;type:text" json:"delivery_area,omitempty"`
	AvailabilityStatus   string           `gorm:"column:availability_status;type:varchar(60);not null;default:'available';index" json:"availability_status"`
	AvailableFrom        *time.Time       `gorm:"column:available_from;index" json:"available_from,omitempty"`
	AvailableUntil       *time.Time       `gorm:"column:available_until;index" json:"available_until,omitempty"`
	Status               SupplyPostStatus `gorm:"column:status;type:varchar(30);not null;default:'active';index" json:"status"`
	CreatedAt            time.Time        `gorm:"column:created_at;not null;autoCreateTime;index" json:"created_at"`
	UpdatedAt            time.Time        `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt   `gorm:"column:deleted_at;index" json:"-"`
}

func (SupplyPost) TableName() string {
	return "supply_posts"
}

type DemandPost struct {
	ID                     string           `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BuyerID                string           `gorm:"column:buyer_id;type:uuid;not null;index" json:"buyer_id"`
	ProductName            string           `gorm:"column:product_name;type:varchar(150);not null;index" json:"product_name"`
	Category               string           `gorm:"column:category;type:varchar(120);not null;index" json:"category"`
	Subcategory            string           `gorm:"column:subcategory;type:varchar(120);index" json:"subcategory,omitempty"`
	Description            string           `gorm:"column:description;type:text" json:"description,omitempty"`
	Quantity               float64          `gorm:"column:quantity;not null" json:"quantity"`
	Unit                   string           `gorm:"column:unit;type:varchar(50);not null;index" json:"unit"`
	BudgetMin              float64          `gorm:"column:budget_min;not null;default:0;index" json:"budget_min"`
	BudgetMax              float64          `gorm:"column:budget_max;not null;default:0;index" json:"budget_max"`
	DeliveryLocation       string           `gorm:"column:delivery_location;type:varchar(180);not null;index" json:"delivery_location"`
	NeededDate             *time.Time       `gorm:"column:needed_date;index" json:"needed_date,omitempty"`
	Frequency              string           `gorm:"column:frequency;type:varchar(80);index" json:"frequency,omitempty"`
	AdditionalRequirements string           `gorm:"column:additional_requirements;type:text" json:"additional_requirements,omitempty"`
	Status                 DemandPostStatus `gorm:"column:status;type:varchar(30);not null;default:'open';index" json:"status"`
	CreatedAt              time.Time        `gorm:"column:created_at;not null;autoCreateTime;index" json:"created_at"`
	UpdatedAt              time.Time        `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt              gorm.DeletedAt   `gorm:"column:deleted_at;index" json:"-"`
}

func (DemandPost) TableName() string {
	return "demand_posts"
}
