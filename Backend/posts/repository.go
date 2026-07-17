package posts

import (
	"context"
	"errors"
	"strings"

	"garuda-hacks/backend/users"
	"gorm.io/gorm"
)

type Repository interface {
	CreateSupply(ctx context.Context, post *SupplyPost) error
	CreateDemand(ctx context.Context, post *DemandPost) error
	GetSupplyByID(ctx context.Context, id string) (*SupplyPost, error)
	GetDemandByID(ctx context.Context, id string) (*DemandPost, error)
	UpdateSupply(ctx context.Context, post *SupplyPost) error
	UpdateDemand(ctx context.Context, post *DemandPost) error
	DeleteSupply(ctx context.Context, post *SupplyPost) error
	DeleteDemand(ctx context.Context, post *DemandPost) error
	ListSupply(ctx context.Context, filter QueryFilter) ([]SupplyPost, error)
	ListDemand(ctx context.Context, filter QueryFilter) ([]DemandPost, error)
	ListSupplyByProducer(ctx context.Context, producerID string, filter QueryFilter) ([]SupplyPost, error)
	ListDemandByBuyer(ctx context.Context, buyerID string, filter QueryFilter) ([]DemandPost, error)
	ListUsersByRole(ctx context.Context, roles []users.UserRole) ([]users.User, error)
	GetUserByID(ctx context.Context, userID string) (*users.User, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) CreateSupply(ctx context.Context, post *SupplyPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *GormRepository) CreateDemand(ctx context.Context, post *DemandPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *GormRepository) GetSupplyByID(ctx context.Context, id string) (*SupplyPost, error) {
	var post SupplyPost
	err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *GormRepository) GetDemandByID(ctx context.Context, id string) (*DemandPost, error) {
	var post DemandPost
	err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *GormRepository) UpdateSupply(ctx context.Context, post *SupplyPost) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *GormRepository) UpdateDemand(ctx context.Context, post *DemandPost) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *GormRepository) DeleteSupply(ctx context.Context, post *SupplyPost) error {
	return r.db.WithContext(ctx).Delete(post).Error
}

func (r *GormRepository) DeleteDemand(ctx context.Context, post *DemandPost) error {
	return r.db.WithContext(ctx).Delete(post).Error
}

func (r *GormRepository) ListSupply(ctx context.Context, filter QueryFilter) ([]SupplyPost, error) {
	var posts []SupplyPost
	order, err := normalizeSort(filter.Sort)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(filter.Sort, "budget_") {
		order = "created_at DESC"
	}

	query := r.db.WithContext(ctx).Model(&SupplyPost{})
	query = applySupplyFilter(query, filter)
	err = query.Order(order).Limit(filter.Limit).Offset(filter.Offset).Find(&posts).Error
	return posts, err
}

func (r *GormRepository) ListDemand(ctx context.Context, filter QueryFilter) ([]DemandPost, error) {
	var posts []DemandPost
	order, err := normalizeSort(filter.Sort)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(filter.Sort, "price_") {
		order = "created_at DESC"
	}

	query := r.db.WithContext(ctx).Model(&DemandPost{})
	query = applyDemandFilter(query, filter)
	err = query.Order(order).Limit(filter.Limit).Offset(filter.Offset).Find(&posts).Error
	return posts, err
}

func (r *GormRepository) ListSupplyByProducer(ctx context.Context, producerID string, filter QueryFilter) ([]SupplyPost, error) {
	filter.Status = strings.TrimSpace(filter.Status)
	var posts []SupplyPost
	order, err := normalizeSort(filter.Sort)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(filter.Sort, "budget_") {
		order = "created_at DESC"
	}

	query := r.db.WithContext(ctx).Model(&SupplyPost{}).Where("producer_id = ?", producerID)
	query = applySupplyFilter(query, filter)
	err = query.Order(order).Limit(filter.Limit).Offset(filter.Offset).Find(&posts).Error
	return posts, err
}

func (r *GormRepository) ListDemandByBuyer(ctx context.Context, buyerID string, filter QueryFilter) ([]DemandPost, error) {
	var posts []DemandPost
	order, err := normalizeSort(filter.Sort)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(filter.Sort, "price_") {
		order = "created_at DESC"
	}

	query := r.db.WithContext(ctx).Model(&DemandPost{}).Where("buyer_id = ?", buyerID)
	query = applyDemandFilter(query, filter)
	err = query.Order(order).Limit(filter.Limit).Offset(filter.Offset).Find(&posts).Error
	return posts, err
}

func (r *GormRepository) ListUsersByRole(ctx context.Context, roles []users.UserRole) ([]users.User, error) {
	var records []users.User
	err := r.db.WithContext(ctx).Preload("Profile").Where("role IN ?", roles).Find(&records).Error
	return records, err
}

func (r *GormRepository) GetUserByID(ctx context.Context, userID string) (*users.User, error) {
	var record users.User
	err := r.db.WithContext(ctx).Preload("Profile").First(&record, "id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func applySupplyFilter(query *gorm.DB, filter QueryFilter) *gorm.DB {
	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(product_name) LIKE ? OR LOWER(category) LIKE ? OR LOWER(subcategory) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ? OR LOWER(delivery_area) LIKE ?", like, like, like, like, like, like)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Subcategory != "" {
		query = query.Where("subcategory = ?", filter.Subcategory)
	}
	if filter.Location != "" {
		like := "%" + strings.ToLower(filter.Location) + "%"
		query = query.Where("LOWER(location) LIKE ? OR LOWER(delivery_area) LIKE ?", like, like)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Unit != "" {
		query = query.Where("unit = ?", filter.Unit)
	}
	if filter.AvailabilityStatus != "" {
		query = query.Where("availability_status = ?", filter.AvailabilityStatus)
	}
	if filter.PriceMin != nil {
		query = query.Where("price_max >= ?", *filter.PriceMin)
	}
	if filter.PriceMax != nil {
		query = query.Where("price_min <= ?", *filter.PriceMax)
	}
	if filter.QuantityMin != nil {
		query = query.Where("quantity >= ?", *filter.QuantityMin)
	}
	if filter.QuantityMax != nil {
		query = query.Where("quantity <= ?", *filter.QuantityMax)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedUntil != nil {
		query = query.Where("created_at <= ?", *filter.CreatedUntil)
	}
	return query
}

func applyDemandFilter(query *gorm.DB, filter QueryFilter) *gorm.DB {
	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(product_name) LIKE ? OR LOWER(category) LIKE ? OR LOWER(subcategory) LIKE ? OR LOWER(description) LIKE ? OR LOWER(delivery_location) LIKE ? OR LOWER(additional_requirements) LIKE ?", like, like, like, like, like, like)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Subcategory != "" {
		query = query.Where("subcategory = ?", filter.Subcategory)
	}
	if filter.Location != "" {
		like := "%" + strings.ToLower(filter.Location) + "%"
		query = query.Where("LOWER(delivery_location) LIKE ?", like)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Unit != "" {
		query = query.Where("unit = ?", filter.Unit)
	}
	if filter.BudgetMin != nil {
		query = query.Where("budget_max >= ?", *filter.BudgetMin)
	}
	if filter.BudgetMax != nil {
		query = query.Where("budget_min <= ?", *filter.BudgetMax)
	}
	if filter.QuantityMin != nil {
		query = query.Where("quantity >= ?", *filter.QuantityMin)
	}
	if filter.QuantityMax != nil {
		query = query.Where("quantity <= ?", *filter.QuantityMax)
	}
	if filter.NeededFrom != nil {
		query = query.Where("needed_date >= ?", *filter.NeededFrom)
	}
	if filter.NeededUntil != nil {
		query = query.Where("needed_date <= ?", *filter.NeededUntil)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedUntil != nil {
		query = query.Where("created_at <= ?", *filter.CreatedUntil)
	}
	return query
}
