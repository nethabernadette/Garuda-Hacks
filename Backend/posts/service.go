package posts

import (
	"context"
	"sort"
	"strings"
)

type NotificationCreator interface {
	CreateUnique(ctx context.Context, userID string, notificationType string, title string, message string, referenceType string, referenceID string) error
}

type Service interface {
	CreateSupply(ctx context.Context, principal Principal, req CreateSupplyPostRequest) (*SupplyPost, error)
	UpdateSupply(ctx context.Context, principal Principal, id string, req UpdateSupplyPostRequest) (*SupplyPost, error)
	DeleteSupply(ctx context.Context, principal Principal, id string) error
	CloseSupply(ctx context.Context, principal Principal, id string) (*SupplyPost, error)
	GetSupply(ctx context.Context, principal Principal, id string) (*SupplyPost, error)
	ListSupply(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error)
	ListMySupply(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error)
	CreateDemand(ctx context.Context, principal Principal, req CreateDemandPostRequest) (*DemandPost, error)
	UpdateDemand(ctx context.Context, principal Principal, id string, req UpdateDemandPostRequest) (*DemandPost, error)
	DeleteDemand(ctx context.Context, principal Principal, id string) error
	CloseDemand(ctx context.Context, principal Principal, id string) (*DemandPost, error)
	GetDemand(ctx context.Context, principal Principal, id string) (*DemandPost, error)
	ListDemand(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error)
	ListMyDemand(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error)
	Feed(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error)
	Search(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error)
}

type service struct {
	repository    Repository
	notifications NotificationCreator
}

func NewService(repository Repository, notifications NotificationCreator) Service {
	return &service{repository: repository, notifications: notifications}
}

func (s *service) CreateSupply(ctx context.Context, principal Principal, req CreateSupplyPostRequest) (*SupplyPost, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if !canCreateSupply(principal.Role) {
		return nil, ErrInvalidRole
	}

	post := &SupplyPost{
		ProducerID:           principal.UserID,
		ProductName:          req.ProductName,
		Category:             req.Category,
		Subcategory:          req.Subcategory,
		Description:          req.Description,
		Quantity:             req.Quantity,
		Unit:                 req.Unit,
		MinimumOrderQuantity: req.MinimumOrderQuantity,
		PriceMin:             req.PriceMin,
		PriceMax:             req.PriceMax,
		Location:             req.Location,
		DeliveryArea:         req.DeliveryArea,
		AvailabilityStatus:   req.AvailabilityStatus,
		Status:               req.Status,
	}
	if req.AvailableFrom != nil && !req.AvailableFrom.IsZero() {
		value := req.AvailableFrom.Time
		post.AvailableFrom = &value
	}
	if req.AvailableUntil != nil && !req.AvailableUntil.IsZero() {
		value := req.AvailableUntil.Time
		post.AvailableUntil = &value
	}
	if err := validateSupplyPost(post); err != nil {
		return nil, err
	}
	if err := s.repository.CreateSupply(ctx, post); err != nil {
		return nil, err
	}
	if post.Status == SupplyPostStatusActive {
		_ = s.notifyBuyersForSupply(ctx, post)
	}
	return post, nil
}

func (s *service) UpdateSupply(ctx context.Context, principal Principal, id string, req UpdateSupplyPostRequest) (*SupplyPost, error) {
	post, err := s.repository.GetSupplyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if post.ProducerID != principal.UserID {
		return nil, ErrPostForbidden
	}

	applySupplyUpdates(post, req)
	if err := validateSupplyPost(post); err != nil {
		return nil, err
	}
	if err := s.repository.UpdateSupply(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *service) DeleteSupply(ctx context.Context, principal Principal, id string) error {
	post, err := s.repository.GetSupplyByID(ctx, id)
	if err != nil {
		return err
	}
	if principal.UserID == "" {
		return ErrUnauthorized
	}
	if post.ProducerID != principal.UserID {
		return ErrPostForbidden
	}
	return s.repository.DeleteSupply(ctx, post)
}

func (s *service) CloseSupply(ctx context.Context, principal Principal, id string) (*SupplyPost, error) {
	post, err := s.repository.GetSupplyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if post.ProducerID != principal.UserID {
		return nil, ErrPostForbidden
	}
	post.Status = SupplyPostStatusClosed
	if err := s.repository.UpdateSupply(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *service) GetSupply(ctx context.Context, principal Principal, id string) (*SupplyPost, error) {
	post, err := s.repository.GetSupplyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.Status != SupplyPostStatusActive && post.ProducerID != principal.UserID {
		return nil, ErrPostForbidden
	}
	return post, nil
}

func (s *service) ListSupply(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error) {
	filter = normalizeFilter(filter)
	if filter.Status == "" && !canCreateSupply(principal.Role) {
		filter.Status = string(SupplyPostStatusActive)
	}
	records, err := s.repository.ListSupply(ctx, filter)
	if err != nil {
		return nil, err
	}
	return supplyFeedItems(records), nil
}

func (s *service) ListMySupply(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	filter = normalizeFilter(filter)
	records, err := s.repository.ListSupplyByProducer(ctx, principal.UserID, filter)
	if err != nil {
		return nil, err
	}
	return supplyFeedItems(records), nil
}

func (s *service) CreateDemand(ctx context.Context, principal Principal, req CreateDemandPostRequest) (*DemandPost, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if !canCreateDemand(principal.Role) {
		return nil, ErrInvalidRole
	}

	post := &DemandPost{
		BuyerID:                principal.UserID,
		ProductName:            req.ProductName,
		Category:               req.Category,
		Subcategory:            req.Subcategory,
		Description:            req.Description,
		Quantity:               req.Quantity,
		Unit:                   req.Unit,
		BudgetMin:              req.BudgetMin,
		BudgetMax:              req.BudgetMax,
		DeliveryLocation:       req.DeliveryLocation,
		Frequency:              req.Frequency,
		AdditionalRequirements: req.AdditionalRequirements,
		Status:                 req.Status,
	}
	if req.NeededDate != nil && !req.NeededDate.IsZero() {
		value := req.NeededDate.Time
		post.NeededDate = &value
	}
	if err := validateDemandPost(post, true); err != nil {
		return nil, err
	}
	if err := s.repository.CreateDemand(ctx, post); err != nil {
		return nil, err
	}
	if post.Status == DemandPostStatusOpen {
		_ = s.notifyProducersForDemand(ctx, post)
	}
	return post, nil
}

func (s *service) UpdateDemand(ctx context.Context, principal Principal, id string, req UpdateDemandPostRequest) (*DemandPost, error) {
	post, err := s.repository.GetDemandByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if post.BuyerID != principal.UserID {
		return nil, ErrPostForbidden
	}
	applyDemandUpdates(post, req)
	if err := validateDemandPost(post, false); err != nil {
		return nil, err
	}
	if err := s.repository.UpdateDemand(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *service) DeleteDemand(ctx context.Context, principal Principal, id string) error {
	post, err := s.repository.GetDemandByID(ctx, id)
	if err != nil {
		return err
	}
	if principal.UserID == "" {
		return ErrUnauthorized
	}
	if post.BuyerID != principal.UserID {
		return ErrPostForbidden
	}
	post.Status = DemandPostStatusCancelled
	return s.repository.UpdateDemand(ctx, post)
}

func (s *service) CloseDemand(ctx context.Context, principal Principal, id string) (*DemandPost, error) {
	post, err := s.repository.GetDemandByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if post.BuyerID != principal.UserID {
		return nil, ErrPostForbidden
	}
	post.Status = DemandPostStatusClosed
	if err := s.repository.UpdateDemand(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *service) GetDemand(ctx context.Context, principal Principal, id string) (*DemandPost, error) {
	post, err := s.repository.GetDemandByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.Status != DemandPostStatusOpen && post.BuyerID != principal.UserID {
		return nil, ErrPostForbidden
	}
	return post, nil
}

func (s *service) ListDemand(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error) {
	filter = normalizeFilter(filter)
	if filter.Status == "" && !canCreateDemand(principal.Role) {
		filter.Status = string(DemandPostStatusOpen)
	}
	records, err := s.repository.ListDemand(ctx, filter)
	if err != nil {
		return nil, err
	}
	return demandFeedItems(records), nil
}

func (s *service) ListMyDemand(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	filter = normalizeFilter(filter)
	records, err := s.repository.ListDemandByBuyer(ctx, principal.UserID, filter)
	if err != nil {
		return nil, err
	}
	return demandFeedItems(records), nil
}

func (s *service) Feed(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error) {
	filter = normalizeFilter(filter)
	switch strings.TrimSpace(filter.Type) {
	case "", "all":
		if canCreateDemand(principal.Role) {
			filter.Status = string(SupplyPostStatusActive)
			items, err := s.ListSupply(ctx, principal, filter)
			if err != nil {
				return nil, err
			}
			return s.scoreItemsForPrincipal(ctx, principal, items), nil
		}
		if canCreateSupply(principal.Role) {
			filter.Status = string(DemandPostStatusOpen)
			items, err := s.ListDemand(ctx, principal, filter)
			if err != nil {
				return nil, err
			}
			return s.scoreItemsForPrincipal(ctx, principal, items), nil
		}
		supply, err := s.repository.ListSupply(ctx, withStatus(filter, string(SupplyPostStatusActive)))
		if err != nil {
			return nil, err
		}
		demand, err := s.repository.ListDemand(ctx, withStatus(filter, string(DemandPostStatusOpen)))
		if err != nil {
			return nil, err
		}
		return mergeFeedItems(supplyFeedItems(supply), demandFeedItems(demand), filter.Limit), nil
	case "supply":
		return s.ListSupply(ctx, principal, filter)
	case "demand":
		return s.ListDemand(ctx, principal, filter)
	default:
		return nil, ErrInvalidPostType
	}
}

func (s *service) Search(ctx context.Context, principal Principal, filter QueryFilter) ([]FeedItem, error) {
	return s.Feed(ctx, principal, filter)
}

func (s *service) scoreItemsForPrincipal(ctx context.Context, principal Principal, items []FeedItem) []FeedItem {
	if principal.UserID == "" {
		return items
	}
	user, err := s.repository.GetUserByID(ctx, principal.UserID)
	if err != nil || user.Profile == nil {
		return items
	}

	for i := range items {
		score := 0
		reasons := make([]string, 0, 3)
		if equalFold(user.Profile.ProductCategory, items[i].Category) {
			score += scoreCategory
			reasons = append(reasons, "profile category matches")
		}
		if containsFold(items[i].Location, user.Profile.City) || containsFold(user.Profile.DeliveryArea, items[i].Location) {
			score += scoreLocation
			reasons = append(reasons, "profile location or delivery area matches")
		}
		if score > 0 {
			items[i].RelevanceScore = score
			items[i].RelevanceReasons = reasons
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].RelevanceScore == items[j].RelevanceScore {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].RelevanceScore > items[j].RelevanceScore
	})
	return items
}

func (s *service) notifyBuyersForSupply(ctx context.Context, supply *SupplyPost) error {
	if s.notifications == nil {
		return nil
	}
	buyers, err := s.repository.ListUsersByRole(ctx, []users.UserRole{users.RoleBuyer})
	if err != nil {
		return err
	}
	for _, buyer := range buyers {
		if buyer.Profile == nil {
			continue
		}
		if !profileMatchesSupply(buyer.Profile.ProductCategory, buyer.Profile.City, supply) {
			continue
		}
		_ = s.notifications.CreateUnique(ctx, buyer.ID, "supply_relevant", "New relevant supply post", supply.ProductName+" is available in "+supply.Location, "supply_post", supply.ID)
	}
	return nil
}

func (s *service) notifyProducersForDemand(ctx context.Context, demand *DemandPost) error {
	if s.notifications == nil {
		return nil
	}
	producers, err := s.repository.ListUsersByRole(ctx, []users.UserRole{users.RoleProducer, users.RoleFarmer})
	if err != nil {
		return err
	}
	for _, producer := range producers {
		if producer.Profile == nil {
			continue
		}
		if !profileMatchesDemand(producer.Profile.ProductCategory, producer.Profile.DeliveryArea, demand) {
			continue
		}
		_ = s.notifications.CreateUnique(ctx, producer.ID, "demand_relevant", "New relevant demand post", demand.ProductName+" is needed in "+demand.DeliveryLocation, "demand_post", demand.ID)
	}
	return nil
}

func applySupplyUpdates(post *SupplyPost, req UpdateSupplyPostRequest) {
	if req.ProductName != nil {
		post.ProductName = *req.ProductName
	}
	if req.Category != nil {
		post.Category = *req.Category
	}
	if req.Subcategory != nil {
		post.Subcategory = *req.Subcategory
	}
	if req.Description != nil {
		post.Description = *req.Description
	}
	if req.Quantity != nil {
		post.Quantity = *req.Quantity
	}
	if req.Unit != nil {
		post.Unit = *req.Unit
	}
	if req.MinimumOrderQuantity != nil {
		post.MinimumOrderQuantity = *req.MinimumOrderQuantity
	}
	if req.PriceMin != nil {
		post.PriceMin = *req.PriceMin
	}
	if req.PriceMax != nil {
		post.PriceMax = *req.PriceMax
	}
	if req.Location != nil {
		post.Location = *req.Location
	}
	if req.DeliveryArea != nil {
		post.DeliveryArea = *req.DeliveryArea
	}
	if req.AvailabilityStatus != nil {
		post.AvailabilityStatus = *req.AvailabilityStatus
	}
	if req.AvailableFrom != nil {
		value := req.AvailableFrom.Time
		post.AvailableFrom = &value
	}
	if req.AvailableUntil != nil {
		value := req.AvailableUntil.Time
		post.AvailableUntil = &value
	}
	if req.Status != nil {
		post.Status = *req.Status
	}
}

func applyDemandUpdates(post *DemandPost, req UpdateDemandPostRequest) {
	if req.ProductName != nil {
		post.ProductName = *req.ProductName
	}
	if req.Category != nil {
		post.Category = *req.Category
	}
	if req.Subcategory != nil {
		post.Subcategory = *req.Subcategory
	}
	if req.Description != nil {
		post.Description = *req.Description
	}
	if req.Quantity != nil {
		post.Quantity = *req.Quantity
	}
	if req.Unit != nil {
		post.Unit = *req.Unit
	}
	if req.BudgetMin != nil {
		post.BudgetMin = *req.BudgetMin
	}
	if req.BudgetMax != nil {
		post.BudgetMax = *req.BudgetMax
	}
	if req.DeliveryLocation != nil {
		post.DeliveryLocation = *req.DeliveryLocation
	}
	if req.NeededDate != nil {
		value := req.NeededDate.Time
		post.NeededDate = &value
	}
	if req.Frequency != nil {
		post.Frequency = *req.Frequency
	}
	if req.AdditionalRequirements != nil {
		post.AdditionalRequirements = *req.AdditionalRequirements
	}
	if req.Status != nil {
		post.Status = *req.Status
	}
}

func normalizeFilter(filter QueryFilter) QueryFilter {
	filter.Type = strings.ToLower(strings.TrimSpace(filter.Type))
	filter.Query = strings.TrimSpace(filter.Query)
	filter.Category = strings.TrimSpace(filter.Category)
	filter.Subcategory = strings.TrimSpace(filter.Subcategory)
	filter.Location = strings.TrimSpace(filter.Location)
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.Unit = strings.TrimSpace(filter.Unit)
	filter.AvailabilityStatus = strings.TrimSpace(filter.AvailabilityStatus)
	filter.Sort = strings.ToLower(strings.TrimSpace(filter.Sort))
	filter.Page, filter.Limit, filter.Offset = normalizePagination(filter.Page, filter.Limit)
	return filter
}

func withStatus(filter QueryFilter, status string) QueryFilter {
	if filter.Status == "" {
		filter.Status = status
	}
	return filter
}

func supplyFeedItems(records []SupplyPost) []FeedItem {
	items := make([]FeedItem, 0, len(records))
	for i := range records {
		post := records[i]
		items = append(items, FeedItem{
			PostType:    "supply",
			ID:          post.ID,
			OwnerID:     post.ProducerID,
			ProductName: post.ProductName,
			Category:    post.Category,
			Subcategory: post.Subcategory,
			Description: post.Description,
			Quantity:    post.Quantity,
			Unit:        post.Unit,
			Location:    post.Location,
			PriceMin:    post.PriceMin,
			PriceMax:    post.PriceMax,
			Status:      string(post.Status),
			CreatedAt:   post.CreatedAt,
			UpdatedAt:   post.UpdatedAt,
			Post:        post,
		})
	}
	return items
}

func demandFeedItems(records []DemandPost) []FeedItem {
	items := make([]FeedItem, 0, len(records))
	for i := range records {
		post := records[i]
		items = append(items, FeedItem{
			PostType:    "demand",
			ID:          post.ID,
			OwnerID:     post.BuyerID,
			ProductName: post.ProductName,
			Category:    post.Category,
			Subcategory: post.Subcategory,
			Description: post.Description,
			Quantity:    post.Quantity,
			Unit:        post.Unit,
			Location:    post.DeliveryLocation,
			BudgetMin:   post.BudgetMin,
			BudgetMax:   post.BudgetMax,
			Status:      string(post.Status),
			CreatedAt:   post.CreatedAt,
			UpdatedAt:   post.UpdatedAt,
			Post:        post,
		})
	}
	return items
}

func mergeFeedItems(a, b []FeedItem, limit int) []FeedItem {
	items := append(a, b...)
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

func profileMatchesSupply(category string, city string, supply *SupplyPost) bool {
	return equalFold(category, supply.Category) || containsFold(supply.DeliveryArea, city) || containsFold(supply.Location, city)
}

func profileMatchesDemand(category string, deliveryArea string, demand *DemandPost) bool {
	return equalFold(category, demand.Category) || containsFold(deliveryArea, demand.DeliveryLocation)
}
