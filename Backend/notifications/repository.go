package notifications

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	CreateUnique(ctx context.Context, notification *Notification) error
	ListByUser(ctx context.Context, userID string, filter QueryFilter) ([]Notification, error)
	GetByID(ctx context.Context, id string) (*Notification, error)
	UnreadCount(ctx context.Context, userID string) (int64, error)
	Update(ctx context.Context, notification *Notification) error
	MarkAllRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, notification *Notification) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) CreateUnique(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "type"},
				{Name: "reference_type"},
				{Name: "reference_id"},
			},
			DoNothing: true,
		}).
		Create(notification).Error
}

func (r *GormRepository) ListByUser(ctx context.Context, userID string, filter QueryFilter) ([]Notification, error) {
	var notifications []Notification
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if filter.Unread != nil {
		query = query.Where("is_read = ?", !*filter.Unread)
	}
	err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&notifications).Error
	return notifications, err
}

func (r *GormRepository) GetByID(ctx context.Context, id string) (*Notification, error) {
	var notification Notification
	err := r.db.WithContext(ctx).First(&notification, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotificationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *GormRepository) UnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *GormRepository) Update(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

func (r *GormRepository) MarkAllRead(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *GormRepository) Delete(ctx context.Context, notification *Notification) error {
	return r.db.WithContext(ctx).Delete(notification).Error
}
