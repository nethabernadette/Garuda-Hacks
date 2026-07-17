package auth

import (
	"context"
	"errors"

	"garuda-hacks/backend/users"
	"gorm.io/gorm"
)

var ErrCredentialsNotFound = errors.New("credentials not found")

type Repository interface {
	CreateUser(ctx context.Context, user *users.User) error
	FindUserByEmail(ctx context.Context, email string) (*users.User, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func NewGormCredentialRepository(db *gorm.DB) *GormRepository {
	return NewGormRepository(db)
}

func (r *GormRepository) CreateUser(ctx context.Context, user *users.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormRepository) FindUserByEmail(ctx context.Context, email string) (*users.User, error) {
	var user users.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCredentialsNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *GormRepository) FindCredentialsByEmail(ctx context.Context, email string) (*users.User, error) {
	return r.FindUserByEmail(ctx, email)
}
