package users

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
	UpdateProfile(ctx context.Context, user *User) error
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

	return r.db.WithContext(ctx).Save(user.Profile).Error
}
