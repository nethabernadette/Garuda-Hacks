package ai

import (
	"context"
	"errors"

	"garuda-hacks/backend/internal/agreement"
	"garuda-hacks/backend/internal/chat"
	"garuda-hacks/backend/posts"
	"garuda-hacks/backend/users"
	"gorm.io/gorm"
)

// Repository defines data access required by AI features.
type Repository interface {
	CreateSearchHistory(ctx context.Context, history *SearchHistory) error
	ListSearchHistory(ctx context.Context, userID string, limit int) ([]SearchHistory, error)
	FindUserByID(ctx context.Context, userID string) (*users.User, error)
	ListSupplyCandidates(ctx context.Context, excludeProducerID string, limit int) ([]posts.SupplyPost, error)
	ListDemandCandidates(ctx context.Context, excludeBuyerID string, limit int) ([]posts.DemandPost, error)
	ListUserSupplyPosts(ctx context.Context, producerID string, limit int) ([]posts.SupplyPost, error)
	ListUserDemandPosts(ctx context.Context, buyerID string, limit int) ([]posts.DemandPost, error)
	FindSupplyPost(ctx context.Context, id string) (*posts.SupplyPost, error)
	FindDemandPost(ctx context.Context, id string) (*posts.DemandPost, error)
	ListAgreementsForUser(ctx context.Context, userID string, limit int) ([]agreement.Agreement, error)
	FindAgreementByID(ctx context.Context, id string) (*agreement.Agreement, error)
	FindMatchByID(ctx context.Context, matchID string) (*matchRecord, error)
	ListMessagesByMatchID(ctx context.Context, matchID string, limit int) ([]chat.Message, error)
}

// GormRepository implements Repository using GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a GORM-backed AI repository.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) CreateSearchHistory(ctx context.Context, history *SearchHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *GormRepository) ListSearchHistory(ctx context.Context, userID string, limit int) ([]SearchHistory, error) {
	var records []SearchHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *GormRepository) FindUserByID(ctx context.Context, userID string) (*users.User, error) {
	var user users.User
	err := r.db.WithContext(ctx).Preload("Profile").First(&user, "id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormRepository) ListSupplyCandidates(ctx context.Context, excludeProducerID string, limit int) ([]posts.SupplyPost, error) {
	var records []posts.SupplyPost
	err := r.db.WithContext(ctx).
		Where("producer_id <> ? AND status = ?", excludeProducerID, posts.SupplyPostStatusActive).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *GormRepository) ListDemandCandidates(ctx context.Context, excludeBuyerID string, limit int) ([]posts.DemandPost, error) {
	var records []posts.DemandPost
	err := r.db.WithContext(ctx).
		Where("buyer_id <> ? AND status = ?", excludeBuyerID, posts.DemandPostStatusOpen).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *GormRepository) ListUserSupplyPosts(ctx context.Context, producerID string, limit int) ([]posts.SupplyPost, error) {
	var records []posts.SupplyPost
	err := r.db.WithContext(ctx).
		Where("producer_id = ?", producerID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *GormRepository) ListUserDemandPosts(ctx context.Context, buyerID string, limit int) ([]posts.DemandPost, error) {
	var records []posts.DemandPost
	err := r.db.WithContext(ctx).
		Where("buyer_id = ?", buyerID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *GormRepository) FindSupplyPost(ctx context.Context, id string) (*posts.SupplyPost, error) {
	var record posts.SupplyPost
	err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *GormRepository) FindDemandPost(ctx context.Context, id string) (*posts.DemandPost, error) {
	var record posts.DemandPost
	err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *GormRepository) ListAgreementsForUser(ctx context.Context, userID string, limit int) ([]agreement.Agreement, error) {
	if !r.db.Migrator().HasTable("matches") {
		return []agreement.Agreement{}, nil
	}
	var records []agreement.Agreement
	err := r.db.WithContext(ctx).
		Preload("Items").
		Joins("JOIN matches ON matches.id = agreements.match_id").
		Where("matches.buyer_id = ? OR matches.producer_id = ?", userID, userID).
		Order("agreements.created_at DESC, agreements.id DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *GormRepository) FindAgreementByID(ctx context.Context, id string) (*agreement.Agreement, error) {
	var record agreement.Agreement
	err := r.db.WithContext(ctx).Preload("Items").First(&record, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAgreementNotFound
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *GormRepository) FindMatchByID(ctx context.Context, matchID string) (*matchRecord, error) {
	if !r.db.Migrator().HasTable("matches") {
		return nil, ErrMatchNotFound
	}
	var record matchRecord
	err := r.db.WithContext(ctx).
		Table(record.TableName()).
		Select("id", "buyer_id", "producer_id").
		First(&record, "id = ?", matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMatchNotFound
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *GormRepository) ListMessagesByMatchID(ctx context.Context, matchID string, limit int) ([]chat.Message, error) {
	var room chat.ChatRoom
	err := r.db.WithContext(ctx).First(&room, "match_id = ?", matchID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []chat.Message{}, nil
	}
	if err != nil {
		return nil, err
	}

	var messages []chat.Message
	err = r.db.WithContext(ctx).
		Where("chat_room_id = ?", room.ID).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// Migrate runs GORM migrations for AI-owned tables.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&SearchHistory{})
}
