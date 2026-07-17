package agreement

import (
	"context"
	"errors"

	"garuda-hacks/backend/users"
	"gorm.io/gorm"
)

// Repository defines persistence operations required by the agreement service.
type Repository interface {
	FindMatchByID(ctx context.Context, matchID string) (*matchRecord, error)
	CreateAgreement(ctx context.Context, agreement *Agreement) error
	FindAgreementByID(ctx context.Context, id string) (*Agreement, error)
	SaveAgreement(ctx context.Context, agreement *Agreement) error
	ReplaceAgreementItems(ctx context.Context, agreement *Agreement, items []AgreementItem) error
	ListAgreementItems(ctx context.Context, agreementID string) ([]AgreementItem, error)
	CountAgreementItems(ctx context.Context, agreementID string) (int64, error)
	CreateAgreementItem(ctx context.Context, item *AgreementItem) error
	FindAgreementItemByID(ctx context.Context, agreementID string, itemID string) (*AgreementItem, error)
	UpdateAgreementItem(ctx context.Context, item *AgreementItem) error
	DeleteAgreementItem(ctx context.Context, item *AgreementItem) error
	FindUserContactByID(ctx context.Context, userID string) (*users.User, error)
}

// GormRepository implements Repository using GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a GORM-backed agreement repository.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// FindMatchByID returns a match projection used for authorization.
func (r *GormRepository) FindMatchByID(ctx context.Context, matchID string) (*matchRecord, error) {
	var match matchRecord
	err := r.db.WithContext(ctx).
		Table(match.TableName()).
		Select("id", "buyer_id", "producer_id").
		First(&match, "id = ?", matchID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMatchNotFound
		}
		return nil, err
	}
	return &match, nil
}

// CreateAgreement stores an agreement and its items if the match has no active agreement.
func (r *GormRepository) CreateAgreement(ctx context.Context, agreement *Agreement) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Agreement{}).
			Where("match_id = ? AND status <> ?", agreement.MatchID, AgreementStatusCancelled).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrActiveAgreementExists
		}
		return tx.Create(agreement).Error
	})
}

// FindAgreementByID returns an agreement with its items.
func (r *GormRepository) FindAgreementByID(ctx context.Context, id string) (*Agreement, error) {
	var agreement Agreement
	err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC, id ASC")
		}).
		First(&agreement, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgreementNotFound
		}
		return nil, err
	}
	return &agreement, nil
}

// SaveAgreement persists agreement state changes.
func (r *GormRepository) SaveAgreement(ctx context.Context, agreement *Agreement) error {
	return r.db.WithContext(ctx).Save(agreement).Error
}

// ReplaceAgreementItems replaces all items for a draft agreement.
func (r *GormRepository) ReplaceAgreementItems(ctx context.Context, agreement *Agreement, items []AgreementItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("agreement_id = ?", agreement.ID).Delete(&AgreementItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].AgreementID = agreement.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		agreement.Items = items
		return tx.Save(agreement).Error
	})
}

// ListAgreementItems returns all items for an agreement.
func (r *GormRepository) ListAgreementItems(ctx context.Context, agreementID string) ([]AgreementItem, error) {
	var items []AgreementItem
	err := r.db.WithContext(ctx).
		Where("agreement_id = ?", agreementID).
		Order("created_at ASC, id ASC").
		Find(&items).Error
	return items, err
}

// CountAgreementItems returns the number of active items for an agreement.
func (r *GormRepository) CountAgreementItems(ctx context.Context, agreementID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&AgreementItem{}).
		Where("agreement_id = ?", agreementID).
		Count(&count).Error
	return count, err
}

// CreateAgreementItem stores a new agreement item.
func (r *GormRepository) CreateAgreementItem(ctx context.Context, item *AgreementItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// FindAgreementItemByID returns an item belonging to an agreement.
func (r *GormRepository) FindAgreementItemByID(ctx context.Context, agreementID string, itemID string) (*AgreementItem, error) {
	var item AgreementItem
	err := r.db.WithContext(ctx).First(&item, "agreement_id = ? AND id = ?", agreementID, itemID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgreementItemNotFound
		}
		return nil, err
	}
	return &item, nil
}

// UpdateAgreementItem persists an agreement item update.
func (r *GormRepository) UpdateAgreementItem(ctx context.Context, item *AgreementItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// DeleteAgreementItem soft-deletes an agreement item.
func (r *GormRepository) DeleteAgreementItem(ctx context.Context, item *AgreementItem) error {
	return r.db.WithContext(ctx).Delete(item).Error
}

// FindUserContactByID returns a user and profile for contact reveal.
func (r *GormRepository) FindUserContactByID(ctx context.Context, userID string) (*users.User, error) {
	var user users.User
	err := r.db.WithContext(ctx).Preload("Profile").First(&user, "id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, err
	}
	return &user, nil
}
