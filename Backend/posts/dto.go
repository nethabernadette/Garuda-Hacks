package posts

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		d.Time = time.Time{}
		return nil
	}

	layouts := []string{"2006-01-02", time.RFC3339}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			d.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("invalid date %q, expected YYYY-MM-DD or RFC3339", value)
}

type CreateSupplyPostRequest struct {
	ProductName          string           `json:"product_name" binding:"required"`
	Category             string           `json:"category" binding:"required"`
	Subcategory          string           `json:"subcategory"`
	Description          string           `json:"description"`
	Quantity             float64          `json:"quantity" binding:"required,gt=0"`
	Unit                 string           `json:"unit" binding:"required"`
	MinimumOrderQuantity float64          `json:"minimum_order_quantity"`
	PriceMin             float64          `json:"price_min"`
	PriceMax             float64          `json:"price_max"`
	Location             string           `json:"location" binding:"required"`
	DeliveryArea         string           `json:"delivery_area"`
	AvailabilityStatus   string           `json:"availability_status"`
	AvailableFrom        *Date            `json:"available_from"`
	AvailableUntil       *Date            `json:"available_until"`
	Status               SupplyPostStatus `json:"status"`
}

type UpdateSupplyPostRequest struct {
	ProductName          *string           `json:"product_name"`
	Category             *string           `json:"category"`
	Subcategory          *string           `json:"subcategory"`
	Description          *string           `json:"description"`
	Quantity             *float64          `json:"quantity"`
	Unit                 *string           `json:"unit"`
	MinimumOrderQuantity *float64          `json:"minimum_order_quantity"`
	PriceMin             *float64          `json:"price_min"`
	PriceMax             *float64          `json:"price_max"`
	Location             *string           `json:"location"`
	DeliveryArea         *string           `json:"delivery_area"`
	AvailabilityStatus   *string           `json:"availability_status"`
	AvailableFrom        *Date             `json:"available_from"`
	AvailableUntil       *Date             `json:"available_until"`
	Status               *SupplyPostStatus `json:"status"`
}

type CreateDemandPostRequest struct {
	ProductName            string           `json:"product_name" binding:"required"`
	Category               string           `json:"category" binding:"required"`
	Subcategory            string           `json:"subcategory"`
	Description            string           `json:"description"`
	Quantity               float64          `json:"quantity" binding:"required,gt=0"`
	Unit                   string           `json:"unit" binding:"required"`
	BudgetMin              float64          `json:"budget_min"`
	BudgetMax              float64          `json:"budget_max"`
	DeliveryLocation       string           `json:"delivery_location" binding:"required"`
	NeededDate             *Date            `json:"needed_date"`
	Frequency              string           `json:"frequency"`
	AdditionalRequirements string           `json:"additional_requirements"`
	Status                 DemandPostStatus `json:"status"`
}

type UpdateDemandPostRequest struct {
	ProductName            *string           `json:"product_name"`
	Category               *string           `json:"category"`
	Subcategory            *string           `json:"subcategory"`
	Description            *string           `json:"description"`
	Quantity               *float64          `json:"quantity"`
	Unit                   *string           `json:"unit"`
	BudgetMin              *float64          `json:"budget_min"`
	BudgetMax              *float64          `json:"budget_max"`
	DeliveryLocation       *string           `json:"delivery_location"`
	NeededDate             *Date             `json:"needed_date"`
	Frequency              *string           `json:"frequency"`
	AdditionalRequirements *string           `json:"additional_requirements"`
	Status                 *DemandPostStatus `json:"status"`
}

type QueryFilter struct {
	Type               string
	Query              string
	Category           string
	Subcategory        string
	Location           string
	Status             string
	Unit               string
	AvailabilityStatus string
	PriceMin           *float64
	PriceMax           *float64
	BudgetMin          *float64
	BudgetMax          *float64
	QuantityMin        *float64
	QuantityMax        *float64
	NeededFrom         *time.Time
	NeededUntil        *time.Time
	CreatedFrom        *time.Time
	CreatedUntil       *time.Time
	Sort               string
	Page               int
	Limit              int
	Offset             int
}

type FeedItem struct {
	PostType         string      `json:"post_type"`
	ID               string      `json:"id"`
	OwnerID          string      `json:"owner_id"`
	ProductName      string      `json:"product_name"`
	Category         string      `json:"category"`
	Subcategory      string      `json:"subcategory,omitempty"`
	Description      string      `json:"description,omitempty"`
	Quantity         float64     `json:"quantity"`
	Unit             string      `json:"unit"`
	Location         string      `json:"location"`
	PriceMin         float64     `json:"price_min,omitempty"`
	PriceMax         float64     `json:"price_max,omitempty"`
	BudgetMin        float64     `json:"budget_min,omitempty"`
	BudgetMax        float64     `json:"budget_max,omitempty"`
	Status           string      `json:"status"`
	RelevanceScore   int         `json:"relevance_score"`
	RelevanceReasons []string  `json:"relevance_reasons,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	Post             interface{} `json:"post,omitempty"`
}

type ListResponse struct {
	Items []FeedItem `json:"items"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}
