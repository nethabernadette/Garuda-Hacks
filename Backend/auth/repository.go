package auth

import (
	"context"
	"errors"

	"garuda-hacks/backend/users"
	"gorm.io/gorm"
)

var ErrCredentialsNotFound = errors.New("credentials not found")

type CredentialRepository interface {
	FindCredentialsByEmail(ctx context.Context, email string) (*users.User, error)
}

type GormCredentialRepository struct {
	db *gorm.DB
}

func NewGormCredentialRepository(db *gorm.DB) *GormCredentialRepository {
	return &GormCredentialRepository{db: db}
}

func (r *GormCredentialRepository) FindCredentialsByEmail(ctx context.Context, email string) (*users.User, error) {
	var user users.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCredentialsNotFound
		}
		return nil, err
	}

	return &user, nil
}
