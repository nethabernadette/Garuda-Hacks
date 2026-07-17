package offer

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

type CreateOfferRequest struct {
	DemandGroupID         uint    `json:"demand_group_id" binding:"required"`
	OfferedPrice          float64 `json:"offered_price" binding:"required,gt=0"`
	EstimatedDeliveryDate *Date   `json:"estimated_delivery_date" binding:"required"`
	Notes                 string  `json:"notes"`
}

type UpdateOfferRequest struct {
	OfferedPrice          *float64 `json:"offered_price"`
	EstimatedDeliveryDate *Date    `json:"estimated_delivery_date"`
	Notes                 *string  `json:"notes"`
}
