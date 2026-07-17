package organizations

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, org *Organization) error
	FindByID(ctx context.Context, id string) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
	GetMembership(ctx context.Context, orgID string, userID string) (*OrganizationMember, error)
	AddMember(ctx context.Context, member *OrganizationMember) error
	RemoveMember(ctx context.Context, orgID string, userID string) error
	UpdateMemberRole(ctx context.Context, orgID string, userID string, role OrganizationRole) error
	ListMembers(ctx context.Context, orgID string) ([]OrganizationMember, error)
	UpdateOwner(ctx context.Context, orgID string, newOwnerID string) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, org *Organization) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(org).Error; err != nil {
			return err
		}

		member := &OrganizationMember{
			OrganizationID: org.ID,
			UserID:         org.OwnerID,
			Role:           OrgRoleOwner,
		}
		return tx.Create(member).Error
	})
}

func (r *GormRepository) FindByID(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	if err := r.db.WithContext(ctx).First(&org, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

func (r *GormRepository) Update(ctx context.Context, org *Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

func (r *GormRepository) GetMembership(ctx context.Context, orgID string, userID string) (*OrganizationMember, error) {
	var member OrganizationMember
	if err := r.db.WithContext(ctx).First(&member, "organization_id = ? AND user_id = ?", orgID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMembershipNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *GormRepository) AddMember(ctx context.Context, member *OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *GormRepository) RemoveMember(ctx context.Context, orgID string, userID string) error {
	return r.db.WithContext(ctx).Delete(&OrganizationMember{}, "organization_id = ? AND user_id = ?", orgID, userID).Error
}

func (r *GormRepository) UpdateMemberRole(ctx context.Context, orgID string, userID string, role OrganizationRole) error {
	return r.db.WithContext(ctx).Model(&OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role", role).Error
}

func (r *GormRepository) ListMembers(ctx context.Context, orgID string) ([]OrganizationMember, error) {
	var members []OrganizationMember
	if err := r.db.WithContext(ctx).Order("joined_at ASC").Find(&members, "organization_id = ?", orgID).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *GormRepository) UpdateOwner(ctx context.Context, orgID string, newOwnerID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update OwnerID in organization
		if err := tx.Model(&Organization{}).Where("id = ?", orgID).Update("owner_id", newOwnerID).Error; err != nil {
			return err
		}

		// Update old owner's role to ADMIN
		if err := tx.Model(&OrganizationMember{}).
			Where("organization_id = ? AND role = ?", orgID, OrgRoleOwner).
			Update("role", OrgRoleAdmin).Error; err != nil {
			return err
		}

		// Update new owner's role to OWNER
		return tx.Model(&OrganizationMember{}).
			Where("organization_id = ? AND user_id = ?", orgID, newOwnerID).
			Update("role", OrgRoleOwner).Error
	})
}
