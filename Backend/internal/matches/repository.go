package matches

import (
	"context"
	"errors"

	"garuda-hacks/backend/users"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	FindUserByID(ctx context.Context, userID string) (*users.User, error)
	FindProducerIDBySupplyPostID(ctx context.Context, postID string) (string, error)
	FindBuyerIDByDemandPostID(ctx context.Context, postID string) (string, error)
	FindFirstActiveSupplyProducerID(ctx context.Context, excludeUserID string) (string, error)
	FindFirstOpenDemandBuyerID(ctx context.Context, excludeUserID string) (string, error)
	FindOrCreate(ctx context.Context, buyerID string, producerID string) (*Match, error)
	FindByID(ctx context.Context, matchID string) (*Match, error)
	ListForUser(ctx context.Context, userID string, limit int, offset int) ([]Match, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindUserByID(ctx context.Context, userID string) (*users.User, error) {
	var user users.User
	err := r.db.WithContext(ctx).Preload("Profile").First(&user, "id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPartnerNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormRepository) FindProducerIDBySupplyPostID(ctx context.Context, postID string) (string, error) {
	var row struct {
		ProducerID string `gorm:"column:producer_id"`
	}
	err := r.db.WithContext(ctx).Table("supply_posts").Select("producer_id").First(&row, "id = ?", postID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrPartnerNotFound
	}
	if err != nil {
		return "", err
	}
	return row.ProducerID, nil
}

func (r *GormRepository) FindBuyerIDByDemandPostID(ctx context.Context, postID string) (string, error) {
	var row struct {
		BuyerID string `gorm:"column:buyer_id"`
	}
	err := r.db.WithContext(ctx).Table("demand_posts").Select("buyer_id").First(&row, "id = ?", postID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrPartnerNotFound
	}
	if err != nil {
		return "", err
	}
	return row.BuyerID, nil
}

func (r *GormRepository) FindFirstActiveSupplyProducerID(ctx context.Context, excludeUserID string) (string, error) {
	var row struct {
		ProducerID string `gorm:"column:producer_id"`
	}
	err := r.db.WithContext(ctx).
		Table("supply_posts").
		Select("producer_id").
		Where("producer_id <> ? AND status = ?", excludeUserID, "active").
		Order("created_at DESC, id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrPartnerNotFound
	}
	if err != nil {
		return "", err
	}
	return row.ProducerID, nil
}

func (r *GormRepository) FindFirstOpenDemandBuyerID(ctx context.Context, excludeUserID string) (string, error) {
	var row struct {
		BuyerID string `gorm:"column:buyer_id"`
	}
	err := r.db.WithContext(ctx).
		Table("demand_posts").
		Select("buyer_id").
		Where("buyer_id <> ? AND status = ?", excludeUserID, "open").
		Order("created_at DESC, id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrPartnerNotFound
	}
	if err != nil {
		return "", err
	}
	return row.BuyerID, nil
}

func (r *GormRepository) FindOrCreate(ctx context.Context, buyerID string, producerID string) (*Match, error) {
	match := &Match{BuyerID: buyerID, ProducerID: producerID, Status: MatchStatusActive}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "buyer_id"}, {Name: "producer_id"}},
		DoNothing: true,
	}).Create(match).Error
	if err != nil {
		return nil, err
	}
	return r.findByPair(ctx, buyerID, producerID)
}

func (r *GormRepository) findByPair(ctx context.Context, buyerID string, producerID string) (*Match, error) {
	var match Match
	err := r.db.WithContext(ctx).First(&match, "buyer_id = ? AND producer_id = ?", buyerID, producerID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMatchNotFound
	}
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *GormRepository) FindByID(ctx context.Context, matchID string) (*Match, error) {
	var match Match
	err := r.db.WithContext(ctx).First(&match, "id = ?", matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMatchNotFound
	}
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *GormRepository) ListForUser(ctx context.Context, userID string, limit int, offset int) ([]Match, error) {
	var records []Match
	err := r.db.WithContext(ctx).
		Where("buyer_id = ? OR producer_id = ?", userID, userID).
		Order("updated_at DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Match{})
}
