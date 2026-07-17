package product

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uint) (*Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, product *Product) error
	ListByProducer(ctx context.Context, producerID uint, limit int, offset int) ([]Product, error)
	Search(ctx context.Context, filter ProductSearchFilter) ([]Product, error)
}

type gormProductRepository struct {
	db *gorm.DB
}

func NewGormProductRepository(db *gorm.DB) ProductRepository {
	return &gormProductRepository{db: db}
}

func (r *gormProductRepository) Create(ctx context.Context, product *Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *gormProductRepository) GetByID(ctx context.Context, id uint) (*Product, error) {
	var product Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *gormProductRepository) Update(ctx context.Context, product *Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *gormProductRepository) Delete(ctx context.Context, product *Product) error {
	return r.db.WithContext(ctx).Delete(product).Error
}

func (r *gormProductRepository) ListByProducer(ctx context.Context, producerID uint, limit int, offset int) ([]Product, error) {
	var products []Product
	err := r.db.WithContext(ctx).
		Where("producer_id = ?", producerID).
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error
	return products, err
}

func (r *gormProductRepository) Search(ctx context.Context, filter ProductSearchFilter) ([]Product, error) {
	var products []Product
	query := r.db.WithContext(ctx).Model(&Product{})

	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}
	if category := strings.TrimSpace(filter.Category); category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.
		Order("updated_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&products).Error
	return products, err
}
