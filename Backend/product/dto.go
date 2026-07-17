package product

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

type CreateProductRequest struct {
	Name               string  `json:"name" binding:"required"`
	Category           string  `json:"category" binding:"required"`
	Description        string  `json:"description"`
	Unit               string  `json:"unit" binding:"required"`
	AvailableStock     float64 `json:"available_stock" binding:"gte=0"`
	ProductionCapacity float64 `json:"production_capacity" binding:"gte=0"`
	HarvestDate        *Date   `json:"harvest_date"`
	MinimumOrder       float64 `json:"minimum_order" binding:"gte=0"`
	ImageURL           string  `json:"image_url"`
}

type UpdateProductRequest struct {
	Name               *string  `json:"name"`
	Category           *string  `json:"category"`
	Description        *string  `json:"description"`
	Unit               *string  `json:"unit"`
	AvailableStock     *float64 `json:"available_stock"`
	ProductionCapacity *float64 `json:"production_capacity"`
	HarvestDate        *Date    `json:"harvest_date"`
	MinimumOrder       *float64 `json:"minimum_order"`
	ImageURL           *string  `json:"image_url"`
}

type UpdateStockRequest struct {
	AvailableStock *float64 `json:"available_stock" binding:"required,gte=0"`
}

type ProductSearchFilter struct {
	Query    string
	Category string
	Limit    int
	Offset   int
}
