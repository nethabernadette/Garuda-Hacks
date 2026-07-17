package agreement

import (
	"time"

	"garuda-hacks/backend/users"
	"gorm.io/gorm"
)

// AgreementStatus represents the lifecycle state of an agreement.
type AgreementStatus string

const (
	AgreementStatusDraft     AgreementStatus = "DRAFT"
	AgreementStatusPending   AgreementStatus = "PENDING"
	AgreementStatusConfirmed AgreementStatus = "CONFIRMED"
	AgreementStatusCancelled AgreementStatus = "CANCELLED"
)

// Agreement stores the procurement agreement created from a successful match.
type Agreement struct {
	ID                  string          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MatchID             string          `gorm:"column:match_id;type:uuid;not null;index" json:"match_id"`
	CreatedBy           string          `gorm:"column:created_by;type:uuid;not null;index" json:"created_by"`
	Status              AgreementStatus `gorm:"column:status;type:varchar(20);not null;default:'DRAFT';index" json:"status"`
	BuyerConfirmed      bool            `gorm:"column:buyer_confirmed;not null;default:false" json:"buyer_confirmed"`
	ProducerConfirmed   bool            `gorm:"column:producer_confirmed;not null;default:false" json:"producer_confirmed"`
	BuyerConfirmedAt    *time.Time      `gorm:"column:buyer_confirmed_at" json:"buyer_confirmed_at,omitempty"`
	ProducerConfirmedAt *time.Time      `gorm:"column:producer_confirmed_at" json:"producer_confirmed_at,omitempty"`
	Items               []AgreementItem `gorm:"foreignKey:AgreementID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"items,omitempty"`
	Creator             users.User      `gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	CreatedAt           time.Time       `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt  `gorm:"column:deleted_at;index" json:"-"`
}

// TableName returns the database table name for agreements.
func (Agreement) TableName() string {
	return "agreements"
}

// AgreementItem stores a single procurement line item.
type AgreementItem struct {
	ID              string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AgreementID     string         `gorm:"column:agreement_id;type:uuid;not null;index" json:"agreement_id"`
	ProductName     string         `gorm:"column:product_name;type:varchar(255);not null" json:"product_name"`
	Quantity        float64        `gorm:"column:quantity;not null" json:"quantity"`
	Unit            string         `gorm:"column:unit;type:varchar(50);not null" json:"unit"`
	UnitPrice       float64        `gorm:"column:unit_price;not null" json:"unit_price"`
	Currency        string         `gorm:"column:currency;type:varchar(10);not null;default:'IDR'" json:"currency"`
	DeliveryDate    time.Time      `gorm:"column:delivery_date;not null;index" json:"delivery_date"`
	DeliveryAddress string         `gorm:"column:delivery_address;type:text;not null" json:"delivery_address"`
	PaymentTerms    string         `gorm:"column:payment_terms;type:text;not null" json:"payment_terms"`
	Specification   string         `gorm:"column:specification;type:text" json:"specification,omitempty"`
	AdditionalNotes string         `gorm:"column:additional_notes;type:text" json:"additional_notes,omitempty"`
	Agreement       Agreement      `gorm:"foreignKey:AgreementID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	CreatedAt       time.Time      `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName returns the database table name for agreement items.
func (AgreementItem) TableName() string {
	return "agreement_items"
}

type matchRecord struct {
	ID         string `gorm:"column:id"`
	BuyerID    string `gorm:"column:buyer_id"`
	ProducerID string `gorm:"column:producer_id"`
}

func (matchRecord) TableName() string {
	return "matches"
}
