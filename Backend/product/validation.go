package product

import (
	"errors"
	"strings"
)

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrProductForbidden = errors.New("producer is not allowed to modify this product")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func validateProduct(product *Product) error {
	if strings.TrimSpace(product.Name) == "" {
		return ValidationError{Message: "product name cannot be empty"}
	}
	if strings.TrimSpace(product.Category) == "" {
		return ValidationError{Message: "category cannot be empty"}
	}
	if strings.TrimSpace(product.Unit) == "" {
		return ValidationError{Message: "unit cannot be empty"}
	}
	if product.AvailableStock < 0 {
		return ValidationError{Message: "stock cannot be negative"}
	}
	if product.ProductionCapacity < 0 {
		return ValidationError{Message: "production capacity cannot be negative"}
	}
	if product.MinimumOrder < 0 {
		return ValidationError{Message: "minimum order cannot be negative"}
	}
	if product.MinimumOrder > product.AvailableStock {
		return ValidationError{Message: "minimum order cannot exceed available stock"}
	}

	product.Name = strings.TrimSpace(product.Name)
	product.Category = strings.TrimSpace(product.Category)
	product.Unit = strings.TrimSpace(product.Unit)
	return nil
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return limit, (page - 1) * limit
}
