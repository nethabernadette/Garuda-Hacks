package document

import (
	"context"
	"errors"
	"time"

	agreementmodule "garuda-hacks/backend/internal/agreement"
	"garuda-hacks/backend/users"
	"gorm.io/gorm"
)

// Repository defines document data access.
type Repository interface {
	FindAgreementByID(ctx context.Context, id string) (*agreementmodule.Agreement, error)
	FindMatchByID(ctx context.Context, id string) (*MatchRecord, error)
	FindUserContactByID(ctx context.Context, userID string) (*users.User, error)
	CountAgreementsInYearThrough(ctx context.Context, yearStart time.Time, yearEnd time.Time, createdAt time.Time) (int64, error)
}

// MatchRecord is a minimal match projection for authorization and parties.
type MatchRecord struct {
	ID         string `gorm:"column:id"`
	BuyerID    string `gorm:"column:buyer_id"`
	ProducerID string `gorm:"column:producer_id"`
}

func (MatchRecord) TableName() string {
	return "matches"
}

// GormRepository implements Repository with GORM.
type GormRepository struct {
	db                  *gorm.DB
	agreementRepository *agreementmodule.GormRepository
}

// NewGormRepository creates a document repository.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db:                  db,
		agreementRepository: agreementmodule.NewGormRepository(db),
	}
}

// FindAgreementByID reuses the agreement repository for agreement loading.
func (r *GormRepository) FindAgreementByID(ctx context.Context, id string) (*agreementmodule.Agreement, error) {
	return r.agreementRepository.FindAgreementByID(ctx, id)
}

// FindUserContactByID reuses the agreement repository's contact projection.
func (r *GormRepository) FindUserContactByID(ctx context.Context, userID string) (*users.User, error) {
	return r.agreementRepository.FindUserContactByID(ctx, userID)
}

// FindMatchByID loads the match participants.
func (r *GormRepository) FindMatchByID(ctx context.Context, id string) (*MatchRecord, error) {
	var match MatchRecord
	err := r.db.WithContext(ctx).
		Table(match.TableName()).
		Select("id", "buyer_id", "producer_id").
		First(&match, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMatchNotFound
		}
		return nil, err
	}
	return &match, nil
}

// CountAgreementsInYearThrough returns a deterministic sequence for document numbers.
func (r *GormRepository) CountAgreementsInYearThrough(ctx context.Context, yearStart time.Time, yearEnd time.Time, createdAt time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&agreementmodule.Agreement{}).
		Where("created_at >= ? AND created_at < ? AND created_at <= ?", yearStart, yearEnd, createdAt).
		Count(&count).Error
	return count, err
}
