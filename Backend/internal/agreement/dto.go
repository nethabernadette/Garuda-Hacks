package agreement

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Date accepts YYYY-MM-DD or RFC3339 dates in request bodies.
type Date struct {
	time.Time
}

// UnmarshalJSON parses a date from JSON.
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

	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			d.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("invalid date %q", value)
}

// CreateAgreementRequest creates an agreement draft for a match.
type CreateAgreementRequest struct {
	MatchID string                 `json:"match_id" validate:"required"`
	Items   []AgreementItemRequest `json:"items" validate:"required"`
}

// UpdateAgreementRequest replaces the editable contents of a draft agreement.
type UpdateAgreementRequest struct {
	Items []AgreementItemRequest `json:"items" validate:"required"`
}

// AgreementItemRequest creates or updates an agreement item.
type AgreementItemRequest struct {
	ProductName     string  `json:"product_name" validate:"required"`
	Quantity        float64 `json:"quantity" validate:"required"`
	Unit            string  `json:"unit" validate:"required"`
	UnitPrice       float64 `json:"unit_price" validate:"required"`
	Currency        string  `json:"currency" validate:"required"`
	DeliveryDate    *Date   `json:"delivery_date" validate:"required"`
	DeliveryAddress string  `json:"delivery_address" validate:"required"`
	PaymentTerms    string  `json:"payment_terms" validate:"required"`
	Specification   string  `json:"specification"`
	AdditionalNotes string  `json:"additional_notes"`
}

// ConfirmAgreementRequest is the request DTO for confirming an agreement.
type ConfirmAgreementRequest struct{}

// AgreementResponse is returned by agreement endpoints.
type AgreementResponse struct {
	ID                  string                  `json:"id"`
	MatchID             string                  `json:"match_id"`
	CreatedBy           string                  `json:"created_by"`
	Status              AgreementStatus         `json:"status"`
	BuyerConfirmed      bool                    `json:"buyer_confirmed"`
	ProducerConfirmed   bool                    `json:"producer_confirmed"`
	BuyerConfirmedAt    *time.Time              `json:"buyer_confirmed_at,omitempty"`
	ProducerConfirmedAt *time.Time              `json:"producer_confirmed_at,omitempty"`
	Items               []AgreementItemResponse `json:"items,omitempty"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
}

// AgreementItemResponse is returned for agreement item endpoints.
type AgreementItemResponse struct {
	ID              string    `json:"id"`
	AgreementID     string    `json:"agreement_id"`
	ProductName     string    `json:"product_name"`
	Quantity        float64   `json:"quantity"`
	Unit            string    `json:"unit"`
	UnitPrice       float64   `json:"unit_price"`
	Currency        string    `json:"currency"`
	DeliveryDate    time.Time `json:"delivery_date"`
	DeliveryAddress string    `json:"delivery_address"`
	PaymentTerms    string    `json:"payment_terms"`
	Specification   string    `json:"specification,omitempty"`
	AdditionalNotes string    `json:"additional_notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// JSONResponse matches the response envelope used by existing modules.
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func newAgreementResponse(agreement *Agreement) AgreementResponse {
	return AgreementResponse{
		ID:                  agreement.ID,
		MatchID:             agreement.MatchID,
		CreatedBy:           agreement.CreatedBy,
		Status:              agreement.Status,
		BuyerConfirmed:      agreement.BuyerConfirmed,
		ProducerConfirmed:   agreement.ProducerConfirmed,
		BuyerConfirmedAt:    agreement.BuyerConfirmedAt,
		ProducerConfirmedAt: agreement.ProducerConfirmedAt,
		Items:               newAgreementItemResponses(agreement.Items),
		CreatedAt:           agreement.CreatedAt,
		UpdatedAt:           agreement.UpdatedAt,
	}
}

func newAgreementItemResponse(item *AgreementItem) AgreementItemResponse {
	return AgreementItemResponse{
		ID:              item.ID,
		AgreementID:     item.AgreementID,
		ProductName:     item.ProductName,
		Quantity:        item.Quantity,
		Unit:            item.Unit,
		UnitPrice:       item.UnitPrice,
		Currency:        item.Currency,
		DeliveryDate:    item.DeliveryDate,
		DeliveryAddress: item.DeliveryAddress,
		PaymentTerms:    item.PaymentTerms,
		Specification:   item.Specification,
		AdditionalNotes: item.AdditionalNotes,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func newAgreementItemResponses(items []AgreementItem) []AgreementItemResponse {
	responses := make([]AgreementItemResponse, 0, len(items))
	for i := range items {
		responses = append(responses, newAgreementItemResponse(&items[i]))
	}
	return responses
}
