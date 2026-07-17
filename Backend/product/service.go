package product

import "context"

type ProductService interface {
	Create(ctx context.Context, producerID uint, req CreateProductRequest) (*Product, error)
	Update(ctx context.Context, producerID uint, id uint, req UpdateProductRequest) (*Product, error)
	Delete(ctx context.Context, producerID uint, id uint) error
	GetByID(ctx context.Context, id uint) (*Product, error)
	ListProducerProducts(ctx context.Context, producerID uint, page int, limit int) ([]Product, error)
	Search(ctx context.Context, query string, category string, page int, limit int) ([]Product, error)
	UpdateStock(ctx context.Context, producerID uint, id uint, req UpdateStockRequest) (*Product, error)
}

type productService struct {
	repository ProductRepository
}

func NewProductService(repository ProductRepository) ProductService {
	return &productService{repository: repository}
}

func (s *productService) Create(ctx context.Context, producerID uint, req CreateProductRequest) (*Product, error) {
	product := &Product{
		ProducerID:         producerID,
		Name:               req.Name,
		Category:           req.Category,
		Description:        req.Description,
		Unit:               req.Unit,
		AvailableStock:     req.AvailableStock,
		ProductionCapacity: req.ProductionCapacity,
		MinimumOrder:       req.MinimumOrder,
		ImageURL:           req.ImageURL,
	}
	if req.HarvestDate != nil && !req.HarvestDate.IsZero() {
		harvestDate := req.HarvestDate.Time
		product.HarvestDate = &harvestDate
	}

	if err := validateProduct(product); err != nil {
		return nil, err
	}
	if err := s.repository.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *productService) Update(ctx context.Context, producerID uint, id uint, req UpdateProductRequest) (*Product, error) {
	product, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product.ProducerID != producerID {
		return nil, ErrProductForbidden
	}

	applyProductUpdates(product, req)
	if err := validateProduct(product); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *productService) Delete(ctx context.Context, producerID uint, id uint) error {
	product, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product.ProducerID != producerID {
		return ErrProductForbidden
	}
	return s.repository.Delete(ctx, product)
}

func (s *productService) GetByID(ctx context.Context, id uint) (*Product, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *productService) ListProducerProducts(ctx context.Context, producerID uint, page int, limit int) ([]Product, error) {
	normalizedLimit, offset := normalizePagination(page, limit)
	return s.repository.ListByProducer(ctx, producerID, normalizedLimit, offset)
}

func (s *productService) Search(ctx context.Context, query string, category string, page int, limit int) ([]Product, error) {
	normalizedLimit, offset := normalizePagination(page, limit)
	return s.repository.Search(ctx, ProductSearchFilter{
		Query:    query,
		Category: category,
		Limit:    normalizedLimit,
		Offset:   offset,
	})
}

func (s *productService) UpdateStock(ctx context.Context, producerID uint, id uint, req UpdateStockRequest) (*Product, error) {
	if req.AvailableStock == nil {
		return nil, ValidationError{Message: "available stock is required"}
	}

	product, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product.ProducerID != producerID {
		return nil, ErrProductForbidden
	}

	product.AvailableStock = *req.AvailableStock
	if err := validateProduct(product); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func applyProductUpdates(product *Product, req UpdateProductRequest) {
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Unit != nil {
		product.Unit = *req.Unit
	}
	if req.AvailableStock != nil {
		product.AvailableStock = *req.AvailableStock
	}
	if req.ProductionCapacity != nil {
		product.ProductionCapacity = *req.ProductionCapacity
	}
	if req.HarvestDate != nil {
		harvestDate := req.HarvestDate.Time
		product.HarvestDate = &harvestDate
	}
	if req.MinimumOrder != nil {
		product.MinimumOrder = *req.MinimumOrder
	}
	if req.ImageURL != nil {
		product.ImageURL = *req.ImageURL
	}
}
