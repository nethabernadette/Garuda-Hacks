package offer

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type OfferRepository interface {
	CreateIfNoDuplicate(ctx context.Context, offer *Offer) error
	GetByID(ctx context.Context, id uint) (*Offer, error)
	Update(ctx context.Context, offer *Offer) error
	ListByProducer(ctx context.Context, producerID uint, limit int, offset int) ([]Offer, error)
	ListByDemandGroup(ctx context.Context, demandGroupID uint, limit int, offset int) ([]Offer, error)
}

type gormOfferRepository struct {
	db *gorm.DB
}

func NewGormOfferRepository(db *gorm.DB) OfferRepository {
	return &gormOfferRepository{db: db}
}

func (r *gormOfferRepository) CreateIfNoDuplicate(ctx context.Context, offer *Offer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.
			Model(&Offer{}).
			Where("group_id = ? AND producer_id = ?", offer.DemandGroupID, offer.ProducerID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateOffer
		}
		if err := tx.Create(offer).Error; err != nil {
			if isDuplicateOfferError(err) {
				return ErrDuplicateOffer
			}
			return err
		}
		return nil
	})
}

func (r *gormOfferRepository) GetByID(ctx context.Context, id uint) (*Offer, error) {
	var offer Offer
	err := r.db.WithContext(ctx).First(&offer, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOfferNotFound
	}
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *gormOfferRepository) Update(ctx context.Context, offer *Offer) error {
	return r.db.WithContext(ctx).Save(offer).Error
}

func (r *gormOfferRepository) ListByProducer(ctx context.Context, producerID uint, limit int, offset int) ([]Offer, error) {
	var offers []Offer
	err := r.db.WithContext(ctx).
		Where("producer_id = ?", producerID).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&offers).Error
	return offers, err
}

func (r *gormOfferRepository) ListByDemandGroup(ctx context.Context, demandGroupID uint, limit int, offset int) ([]Offer, error) {
	var offers []Offer
	err := r.db.WithContext(ctx).
		Where("group_id = ?", demandGroupID).
		Order("price ASC, estimated_delivery ASC").
		Limit(limit).
		Offset(offset).
		Find(&offers).Error
	return offers, err
}

func isDuplicateOfferError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "idx_offers_group_producer") ||
		strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint")
}
