package users

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleBuyer    UserRole = "BUYER"
	RoleProducer UserRole = "PRODUCER"
	RoleAdmin    UserRole = "ADMIN"
)

type NIBVerificationStatus string

const (
	NIBVerificationStatusPending  NIBVerificationStatus = "PENDING"
	NIBVerificationStatusVerified NIBVerificationStatus = "VERIFIED"
	NIBVerificationStatusRejected NIBVerificationStatus = "REJECTED"
)

type User struct {
	ID           string           `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Role         UserRole         `gorm:"column:role;type:varchar(20);not null;index" json:"role"`
	Email        string           `gorm:"column:email;type:varchar(255);not null;uniqueIndex" json:"email"`
	PasswordHash string           `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Profile      *UserProfile     `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"profile,omitempty"`
	Verification *NIBVerification `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"verification,omitempty"`
	CreatedAt    time.Time        `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time        `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"column:deleted_at;index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type UserProfile struct {
	ID                string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            string         `gorm:"column:user_id;type:uuid;not null;uniqueIndex" json:"user_id"`
	CompanyName       string         `gorm:"column:company_name;type:varchar(255);not null" json:"company_name"`
	Phone             string         `gorm:"column:phone;type:varchar(50);not null" json:"phone"`
	City              string         `gorm:"column:city;type:varchar(120);not null;index" json:"city"`
	BusinessType      string         `gorm:"column:business_type;type:varchar(120)" json:"business_type,omitempty"`
	ProductCategory   string         `gorm:"column:product_category;type:varchar(120);index" json:"product_category,omitempty"`
	Capacity          string         `gorm:"column:capacity;type:varchar(120)" json:"capacity,omitempty"`
	MOQ               string         `gorm:"column:moq;type:varchar(120)" json:"moq,omitempty"`
	Certifications    string         `gorm:"column:certifications;type:text" json:"certifications,omitempty"`
	DeliveryArea      string         `gorm:"column:delivery_area;type:text" json:"delivery_area,omitempty"`
	Availability      string         `gorm:"column:availability;type:varchar(120)" json:"availability,omitempty"`
	PurchaseFrequency string         `gorm:"column:purchase_frequency;type:varchar(120)" json:"purchase_frequency,omitempty"`
	CreatedAt         time.Time      `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

type NIBVerification struct {
	ID              string                `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          string                `gorm:"column:user_id;type:uuid;not null;uniqueIndex" json:"user_id"`
	NIBNumber       string                `gorm:"column:nib_number;type:varchar(64);not null;uniqueIndex" json:"nib_number"`
	Status          NIBVerificationStatus `gorm:"column:status;type:varchar(20);not null;default:'PENDING';index" json:"status"`
	VerifiedAt      *time.Time            `gorm:"column:verified_at" json:"verified_at,omitempty"`
	RejectedAt      *time.Time            `gorm:"column:rejected_at" json:"rejected_at,omitempty"`
	RejectionReason string                `gorm:"column:rejection_reason;type:text" json:"rejection_reason,omitempty"`
	CreatedAt       time.Time             `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time             `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt        `gorm:"column:deleted_at;index" json:"-"`
}

func (NIBVerification) TableName() string {
	return "nib_verifications"
}

func (r UserRole) IsValid() bool {
	switch r {
	case RoleBuyer, RoleProducer, RoleAdmin:
		return true
	default:
		return false
	}
}

func (s NIBVerificationStatus) IsValid() bool {
	switch s {
	case NIBVerificationStatusPending, NIBVerificationStatusVerified, NIBVerificationStatusRejected:
		return true
	default:
		return false
	}
}
