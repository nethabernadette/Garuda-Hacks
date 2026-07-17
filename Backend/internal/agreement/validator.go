package agreement

import "strings"

func validateAgreementRequest(req CreateAgreementRequest) error {
	if strings.TrimSpace(req.MatchID) == "" {
		return ErrInvalidMatchID
	}
	return validateItemRequests(req.Items)
}

func validateUpdateAgreementRequest(req UpdateAgreementRequest) error {
	return validateItemRequests(req.Items)
}

func validateItemRequests(items []AgreementItemRequest) error {
	if len(items) == 0 {
		return ErrAgreementNeedsItems
	}
	for i := range items {
		if err := validateItemRequest(items[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateItemRequest(req AgreementItemRequest) error {
	if strings.TrimSpace(req.ProductName) == "" {
		return ErrRequiredProductName
	}
	if req.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if strings.TrimSpace(req.Unit) == "" {
		return ErrRequiredUnit
	}
	if req.UnitPrice <= 0 {
		return ErrInvalidUnitPrice
	}
	if strings.TrimSpace(req.Currency) == "" {
		return ErrRequiredCurrency
	}
	if req.DeliveryDate == nil || req.DeliveryDate.IsZero() {
		return ErrRequiredDeliveryDate
	}
	if strings.TrimSpace(req.DeliveryAddress) == "" {
		return ErrRequiredDeliveryAddress
	}
	if strings.TrimSpace(req.PaymentTerms) == "" {
		return ErrRequiredPaymentTerms
	}
	return nil
}

func agreementItemFromRequest(req AgreementItemRequest) AgreementItem {
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "IDR"
	}

	return AgreementItem{
		ProductName:     strings.TrimSpace(req.ProductName),
		Quantity:        req.Quantity,
		Unit:            strings.TrimSpace(req.Unit),
		UnitPrice:       req.UnitPrice,
		Currency:        currency,
		DeliveryDate:    req.DeliveryDate.Time,
		DeliveryAddress: strings.TrimSpace(req.DeliveryAddress),
		PaymentTerms:    strings.TrimSpace(req.PaymentTerms),
		Specification:   strings.TrimSpace(req.Specification),
		AdditionalNotes: strings.TrimSpace(req.AdditionalNotes),
	}
}

func agreementItemsFromRequests(requests []AgreementItemRequest) []AgreementItem {
	items := make([]AgreementItem, 0, len(requests))
	for _, req := range requests {
		items = append(items, agreementItemFromRequest(req))
	}
	return items
}
