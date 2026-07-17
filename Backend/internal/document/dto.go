package document

import "time"

// ProcurementSummaryResponse contains the structured RFQ/procurement summary.
type ProcurementSummaryResponse struct {
	DocumentNumber                string                           `json:"document_number"`
	GeneratedDate                 time.Time                        `json:"generated_date"`
	AgreementID                   string                           `json:"agreement_id"`
	ProducerCompany               string                           `json:"producer_company"`
	BuyerCompany                  string                           `json:"buyer_company"`
	ProductList                   []ProcurementSummaryItemResponse `json:"product_list"`
	TotalValue                    float64                          `json:"total_value"`
	Currency                      string                           `json:"currency"`
	DeliveryAddress               string                           `json:"delivery_address"`
	PaymentTerms                  string                           `json:"payment_terms"`
	AdditionalNotes               string                           `json:"additional_notes,omitempty"`
	AgreementStatus               string                           `json:"agreement_status"`
	ProducerConfirmationTimestamp *time.Time                       `json:"producer_confirmation_timestamp,omitempty"`
	BuyerConfirmationTimestamp    *time.Time                       `json:"buyer_confirmation_timestamp,omitempty"`
}

// ProcurementSummaryItemResponse represents one product line in the summary.
type ProcurementSummaryItemResponse struct {
	ProductName     string    `json:"product_name"`
	Quantity        float64   `json:"quantity"`
	Unit            string    `json:"unit"`
	UnitPrice       float64   `json:"unit_price"`
	Currency        string    `json:"currency"`
	TotalValue      float64   `json:"total_value"`
	Specifications  string    `json:"specifications,omitempty"`
	DeliveryDate    time.Time `json:"delivery_date"`
	DeliveryAddress string    `json:"delivery_address"`
	PaymentTerms    string    `json:"payment_terms"`
	AdditionalNotes string    `json:"additional_notes,omitempty"`
}

// DocumentResponse returns metadata plus printable HTML.
type DocumentResponse struct {
	DocumentNumber string                     `json:"document_number"`
	AgreementID    string                     `json:"agreement_id"`
	GeneratedDate  time.Time                  `json:"generated_date"`
	Summary        ProcurementSummaryResponse `json:"summary"`
	HTML           string                     `json:"html"`
}

// ContactResponse returns company contacts after both parties confirmed.
type ContactResponse struct {
	AgreementID string               `json:"agreement_id"`
	MatchID     string               `json:"match_id"`
	Buyer       ContactPartyResponse `json:"buyer"`
	Producer    ContactPartyResponse `json:"producer"`
}

// ContactPartyResponse exposes confirmed post-agreement company contact data.
type ContactPartyResponse struct {
	UserID                 string `json:"user_id"`
	CompanyName            string `json:"company_name"`
	BusinessAddress        string `json:"business_address"`
	Email                  string `json:"email"`
	PhoneNumber            string `json:"phone_number"`
	Website                string `json:"website"`
	BusinessRepresentative string `json:"business_representative"`
}

// JSONResponse matches the response envelope used by existing modules.
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
