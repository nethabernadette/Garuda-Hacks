package users

import (
	"context"
	"errors"

	"gorm.io/gorm/clause"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
	UpdateProfile(ctx context.Context, user *User) error
	FindVerificationByUserID(ctx context.Context, userID string) (*NIBVerification, error)
	FindVerificationByID(ctx context.Context, id string) (*NIBVerification, error)
	UpsertVerification(ctx context.Context, verification *NIBVerification) error
	ListVerifications(ctx context.Context) ([]NIBVerification, error)
	UpdateVerification(ctx context.Context, verification *NIBVerification) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormRepository) FindByID(ctx context.Context, id string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Preload("Profile").Preload("Verification").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *GormRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Preload("Profile").Preload("Verification").First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *GormRepository) List(ctx context.Context) ([]User, error) {
	var users []User
	if err := r.db.WithContext(ctx).Preload("Profile").Preload("Verification").Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (r *GormRepository) UpdateProfile(ctx context.Context, user *User) error {
	if user.Profile == nil {
		return nil
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"company_name",
				"phone",
				"city",
				"business_type",
				"product_category",
				"capacity",
				"moq",
				"certifications",
				"delivery_area",
				"availability",
				"purchase_frequency",
				"updated_at",
			}),
		}).
		Create(user.Profile).Error
}

func (r *GormRepository) FindVerificationByUserID(ctx context.Context, userID string) (*NIBVerification, error) {
	var verification NIBVerification
	if err := r.db.WithContext(ctx).First(&verification, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVerificationNotFound
		}
		return nil, err
	}

	return &verification, nil
}

func (r *GormRepository) FindVerificationByID(ctx context.Context, id string) (*NIBVerification, error) {
	var verification NIBVerification
	if err := r.db.WithContext(ctx).First(&verification, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVerificationNotFound
		}
		return nil, err
	}

	return &verification, nil
}

func (r *GormRepository) UpsertVerification(ctx context.Context, verification *NIBVerification) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"nib_number":       verification.NIBNumber,
				"status":           verification.Status,
				"verified_at":      verification.VerifiedAt,
				"rejected_at":      verification.RejectedAt,
				"rejection_reason": verification.RejectionReason,
				"updated_at":       gorm.Expr("NOW()"),
			}),
		}).
		Create(verification).Error
}

func (r *GormRepository) ListVerifications(ctx context.Context) ([]NIBVerification, error) {
	var verifications []NIBVerification
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&verifications).Error; err != nil {
		return nil, err
	}

	return verifications, nil
}

func (r *GormRepository) UpdateVerification(ctx context.Context, verification *NIBVerification) error {
	return r.db.WithContext(ctx).Save(verification).Error
}
