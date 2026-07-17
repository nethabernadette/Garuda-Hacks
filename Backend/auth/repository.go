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
	FindUserByID(ctx context.Context, id string) (*users.User, error)
	SaveRefreshToken(ctx context.Context, rt *RefreshToken) error
	FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
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

func (r *GormRepository) FindUserByID(ctx context.Context, id string) (*users.User, error) {
	var user users.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormRepository) SaveRefreshToken(ctx context.Context, rt *RefreshToken) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *GormRepository) FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	var rt RefreshToken
	if err := r.db.WithContext(ctx).First(&rt, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}
	return &rt, nil
}

func (r *GormRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&RefreshToken{}).Where("token = ?", token).Update("revoked_at", gorm.Expr("NOW()")).Error
}
