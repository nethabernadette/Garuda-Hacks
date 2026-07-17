package posts

import (
	"errors"
	"strings"
	"time"

	"garuda-hacks/backend/users"
)

var (
	ErrPostNotFound       = errors.New("post not found")
	ErrPostForbidden      = errors.New("user is not allowed to modify this post")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidRole        = errors.New("role is not allowed for this action")
	ErrInvalidPostType    = errors.New("invalid post type")
	ErrInvalidSort        = errors.New("invalid sort")
	ErrInvalidQueryFilter = errors.New("invalid query filter")
)

type Principal struct {
	UserID string
	Role   users.UserRole
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func validateSupplyPost(post *SupplyPost) error {
	post.ProductName = strings.TrimSpace(post.ProductName)
	post.Category = strings.TrimSpace(post.Category)
	post.Subcategory = strings.TrimSpace(post.Subcategory)
	post.Unit = strings.TrimSpace(post.Unit)
	post.Location = strings.TrimSpace(post.Location)
	post.DeliveryArea = strings.TrimSpace(post.DeliveryArea)
	post.AvailabilityStatus = strings.TrimSpace(post.AvailabilityStatus)

	if post.ProductName == "" {
		return ValidationError{Message: "product name cannot be empty"}
	}
	if post.Category == "" {
		return ValidationError{Message: "category cannot be empty"}
	}
	if post.Quantity <= 0 {
		return ValidationError{Message: "quantity must be greater than zero"}
	}
	if post.Unit == "" {
		return ValidationError{Message: "unit cannot be empty"}
	}
	if post.MinimumOrderQuantity < 0 {
		return ValidationError{Message: "minimum order quantity cannot be negative"}
	}
	if post.PriceMin < 0 || post.PriceMax < 0 {
		return ValidationError{Message: "price cannot be negative"}
	}
	if post.PriceMin > post.PriceMax {
		return ValidationError{Message: "price_min cannot be greater than price_max"}
	}
	if post.Location == "" {
		return ValidationError{Message: "location cannot be empty"}
	}
	if post.AvailabilityStatus == "" {
		post.AvailabilityStatus = "available"
	}
	if post.AvailableFrom != nil && post.AvailableUntil != nil && post.AvailableUntil.Before(*post.AvailableFrom) {
		return ValidationError{Message: "available_until cannot be earlier than available_from"}
	}
	if post.Status == "" {
		post.Status = SupplyPostStatusActive
	}
	if !post.Status.IsValid() {
		return ValidationError{Message: "invalid supply post status"}
	}
	return nil
}

func validateDemandPost(post *DemandPost, creating bool) error {
	post.ProductName = strings.TrimSpace(post.ProductName)
	post.Category = strings.TrimSpace(post.Category)
	post.Subcategory = strings.TrimSpace(post.Subcategory)
	post.Unit = strings.TrimSpace(post.Unit)
	post.DeliveryLocation = strings.TrimSpace(post.DeliveryLocation)
	post.Frequency = strings.TrimSpace(post.Frequency)

	if post.ProductName == "" {
		return ValidationError{Message: "product name cannot be empty"}
	}
	if post.Category == "" {
		return ValidationError{Message: "category cannot be empty"}
	}
	if post.Quantity <= 0 {
		return ValidationError{Message: "quantity must be greater than zero"}
	}
	if post.Unit == "" {
		return ValidationError{Message: "unit cannot be empty"}
	}
	if post.BudgetMin < 0 || post.BudgetMax < 0 {
		return ValidationError{Message: "budget cannot be negative"}
	}
	if post.BudgetMin > post.BudgetMax {
		return ValidationError{Message: "budget_min cannot be greater than budget_max"}
	}
	if post.DeliveryLocation == "" {
		return ValidationError{Message: "delivery_location cannot be empty"}
	}
	if creating && post.NeededDate != nil && isBeforeToday(*post.NeededDate) {
		return ValidationError{Message: "needed_date cannot be in the past"}
	}
	if post.Status == "" {
		post.Status = DemandPostStatusOpen
	}
	if !post.Status.IsValid() {
		return ValidationError{Message: "invalid demand post status"}
	}
	return nil
}

func (s SupplyPostStatus) IsValid() bool {
	switch s {
	case SupplyPostStatusDraft, SupplyPostStatusActive, SupplyPostStatusClosed, SupplyPostStatusExpired:
		return true
	default:
		return false
	}
}

func (s DemandPostStatus) IsValid() bool {
	switch s {
	case DemandPostStatusDraft, DemandPostStatusOpen, DemandPostStatusMatched, DemandPostStatusClosed, DemandPostStatusExpired, DemandPostStatusCancelled:
		return true
	default:
		return false
	}
}

func normalizePagination(page, limit int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit, (page - 1) * limit
}

func normalizeSort(sort string) (string, error) {
	switch strings.TrimSpace(sort) {
	case "", "newest":
		return "created_at DESC", nil
	case "oldest":
		return "created_at ASC", nil
	case "updated":
		return "updated_at DESC", nil
	case "price_asc":
		return "price_min ASC, created_at DESC", nil
	case "price_desc":
		return "price_max DESC, created_at DESC", nil
	case "budget_asc":
		return "budget_min ASC, created_at DESC", nil
	case "budget_desc":
		return "budget_max DESC, created_at DESC", nil
	default:
		return "", ErrInvalidSort
	}
}

func isBeforeToday(value time.Time) bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	date := value.In(now.Location())
	startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, now.Location())
	return startOfDate.Before(today)
}

func canCreateSupply(role users.UserRole) bool {
	return role == users.RoleProducer || role == users.RoleFarmer
}

func canCreateDemand(role users.UserRole) bool {
	return role == users.RoleBuyer
}
