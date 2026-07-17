package organizations

import (
	"time"

	"gorm.io/gorm"
)

type OrganizationRole string

const (
	OrgRoleOwner  OrganizationRole = "OWNER"
	OrgRoleAdmin  OrganizationRole = "ADMIN"
	OrgRoleMember OrganizationRole = "MEMBER"
)

type Organization struct {
	ID          string               `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string               `gorm:"column:name;type:varchar(255);not null;index" json:"name"`
	Description string               `gorm:"column:description;type:text" json:"description,omitempty"`
	OwnerID     string               `gorm:"column:owner_id;type:uuid;not null;index" json:"owner_id"`
	Members     []OrganizationMember `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"members,omitempty"`
	CreatedAt   time.Time            `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time            `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt       `gorm:"column:deleted_at;index" json:"-"`
}

func (Organization) TableName() string {
	return "organizations"
}

type OrganizationMember struct {
	ID             string           `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID string           `gorm:"column:organization_id;type:uuid;not null;uniqueIndex:idx_org_user" json:"organization_id"`
	UserID         string           `gorm:"column:user_id;type:uuid;not null;uniqueIndex:idx_org_user" json:"user_id"`
	Role           OrganizationRole `gorm:"column:role;type:varchar(20);not null;default:'MEMBER'" json:"role"`
	JoinedAt       time.Time        `gorm:"column:joined_at;not null;autoCreateTime" json:"joined_at"`
	UpdatedAt      time.Time        `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}
